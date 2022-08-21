package ctrl

import (
	"fmt"
	"net/http"

	"github.com/ayo-ajayi/proj/database"
	"github.com/ayo-ajayi/proj/model"
	"github.com/ayo-ajayi/proj/token"
	"github.com/gin-gonic/gin"
)

func Welcome(c *gin.Context){
	c.JSON(200, gin.H{"message": "Welcome"})
}
func Signup(c *gin.Context) {
	var newUser model.User
	if err := c.BindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	insertedID, err := database.CreateUser(newUser)
	if err != nil {
		c.JSON(http.StatusForbidden, err.Error())
		return
	}
	// ts, createErr := token.CreateToken(insertedID) //Id from inserted user
	// if createErr != nil {
	// 	c.JSON(http.StatusForbidden, createErr.Error())
	// 	return
	// }
	// //log.Printf("first: %v",ts.AccessToken)

	// // Save the tokens metadata to Redis
	// saveErr := token.CreateAuth(insertedID, ts)
	// if saveErr != nil {
	// 	c.JSON(http.StatusForbidden, saveErr.Error())
	// 	return
	// }
	// //log.Printf("second: %v",ts.AccessToken)

	// tokens := map[string]string{
	// 	"access_token":  ts.AccessToken,
	// 	"refresh_token": ts.RefreshToken,
	// }

	x := insertedID.Hex()
	c.JSON(http.StatusCreated, gin.H{"message": fmt.Sprintf("User with ID: %v created", x)})
}

func Login(c *gin.Context) {
	var currentUser model.User
	if err := c.BindJSON(&currentUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	storedUser, err := database.FindUser(currentUser)
	// storedUser := *stored
	if err != nil {
		c.JSON(http.StatusForbidden, err.Error())
		return
	}
	if storedUser.Username != currentUser.Username || storedUser.Password != currentUser.Password {
		c.JSON(http.StatusUnauthorized, "Incorrect password")
		return
	}

	ts, createErr := token.CreateToken(storedUser.ID) //Id from inserted user
	if createErr != nil {
		c.JSON(http.StatusForbidden, createErr.Error())
		return
	}

	// Save the tokens metadata to Redis
	saveErr := token.CreateAuth(storedUser.ID, ts)
	if saveErr != nil {
		c.JSON(http.StatusForbidden, saveErr.Error())
		return
	}

	c.JSON(http.StatusCreated, map[string]string{
		"access_token":  ts.AccessToken,
		"refresh_token": ts.RefreshToken,
	})
}

func Logout(c *gin.Context) {
	au, err := token.ExtractTokenMetadata(c.Request)

	if err != nil {
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}
	deleted, delErr := token.DeleteAuth(au.AccessUuid)
	if delErr != nil || deleted == 0 { // if any goes wrong
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}
	c.JSON(http.StatusOK, "Sucessfully logged out")
}

// how to embed a curl script that posts the login details automatically and returns the access token...within an html of the verification mail

func NewTodo(c *gin.Context) {
	tokenAuth, err := token.ExtractTokenMetadata(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}
	var td model.Todo
	if err := c.ShouldBindJSON(&td); err != nil {
		c.JSON(http.StatusUnprocessableEntity, "invalid json")
		return
	}
	td.UserID = tokenAuth.UserId
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, "unauthorized")
		return
	}
	if err = database.CreateTodo(td); err != nil {
		c.JSON(http.StatusForbidden, err.Error())
		return
	}
	c.JSON(http.StatusCreated, td)
}



func GetTodoByUser(c *gin.Context) {
	tokenAuth, err := token.ExtractTokenMetadata(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}
	userId := tokenAuth.UserId
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, "unauthorized")
		return
	}
	resArr, err := database.FindUserTodos(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"user_id": userId, "todos": resArr})
}
