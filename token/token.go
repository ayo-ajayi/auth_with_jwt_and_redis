package token

import (
	"log"
	"net/http"
	"os"

	"time"

	"fmt"

	//"github.com/ayo-ajayi/proj/database"
	"github.com/ayo-ajayi/proj/model"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
	"github.com/twinj/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"strings"
)
var client *redis.Client

func init() {
	//os.Setenv("REDIS_DSN", "")
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	dsn := os.Getenv("REDIS_DSN")
	dsn_pass:=os.Getenv("REDIS_PASS")
	if dsn=="" || dsn_pass ==""{
		log.Fatal("error: unset or incorrect redis login credientials")
	}
	client = redis.NewClient(&redis.Options{
		Addr:     dsn, // redis port
		Password: dsn_pass,
	})
	if _, err = client.Ping().Result(); err != nil {
		panic(err)
	}
	
}


// Create Access Token and Refresh Token
func CreateToken(userid primitive.ObjectID) (*model.TokenDetails, error) {
	td := &model.TokenDetails{}
	td.AtExpires = time.Now().Add(time.Minute * 15).Unix()
	td.AccessUuid = uuid.NewV4().String()
	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
	td.RefreshUuid = uuid.NewV4().String()

	var err error
	// Create the Access Token
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUuid
	atClaims["user_id"] = userid
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return nil, err
	}

	// Creating the Refresh Token
	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["user_id"] = userid
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(os.Getenv("REFRESH_SECRET")))
	if err != nil {
		return nil, err
	}

	return td, nil
}


// Save JWTs metadata to Redis
func CreateAuth(userid primitive.ObjectID, td *model.TokenDetails) error {
	at := time.Unix(td.AtExpires, 0) // converts Unix to UTC
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()

	errAccess := client.Set(td.AccessUuid, userid.Hex(), at.Sub(now)).Err()
	if errAccess != nil {
		return errAccess
	}
	errRefresh := client.Set(td.RefreshUuid, userid.Hex(), rt.Sub(now)).Err()
	if errRefresh != nil {
		return errRefresh
	}
	return nil
}

func ExtractToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearerToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func VerifyToken(r *http.Request)(*jwt.Token, error) {
	tokenString:=ExtractToken(r)
	token, err :=jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Make sure that the token method conforms to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		access_Secret:=os.Getenv("ACCESS_SECRET")
		return []byte(access_Secret), nil
	})
	if err != nil {
		return nil, err
	}
	
	return token, nil
}


func TokenValid(r *http.Request) error {
	token, err := VerifyToken(r)
	if err != nil {
		return err
	}
	if _, ok := token.Claims.(jwt.MapClaims); !ok && !token.Valid {
		return err
	}
	return nil
}


// Extract metadata from token so as to look it up in Redis
func ExtractTokenMetadata(r *http.Request) (*model.AccessDetails, error) {
	token, err := VerifyToken(r)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUuid, ok := claims["access_uuid"].(string)
		if !ok {
			return nil, err
		}
		userIdString, _ := claims["user_id"].(string)
		userId, err := primitive.ObjectIDFromHex(userIdString)
		if err != nil {
			return nil, err
		}
		
		return &model.AccessDetails{
			AccessUuid: accessUuid,
			UserId:     userId,
		}, nil
	}
	return nil, err
}

// Delete JWT metadata from redis store
func DeleteAuth(givenUuid string) (int64, error) {
	deleted, err := client.Del(givenUuid).Result()
	if err != nil {
		return 0, err
	}	
	return deleted, nil
}

