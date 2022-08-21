package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)


type User struct {
	ID       primitive.ObjectID `json:"id" bson:"_id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
}

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

type Todo struct {
	UserID primitive.ObjectID `json:"user_id" bson:"user_id"`
	Title  string `json:"title"`
}

type AccessDetails struct {
	AccessUuid string
	UserId     primitive.ObjectID
}