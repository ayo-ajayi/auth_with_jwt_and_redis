package main

import (
	"github.com/ayo-ajayi/proj/ctrl"
	"github.com/gin-gonic/gin"

	"log"
)


func main() {
	
	router := gin.Default()
	router.GET("/", ctrl.Welcome)
	router.POST("/signup", ctrl.Signup)
	router.POST("/login", ctrl.Login)
	router.POST("/logout", ctrl.TokenAuthMiddleware(), ctrl.Logout)
	router.POST("/newtodo", ctrl.TokenAuthMiddleware(), ctrl.NewTodo)
	router.GET("/todos", ctrl.TokenAuthMiddleware(), ctrl.GetTodoByUser)
	log.Fatal(router.Run(":8080"))
}

//create a /admin type of the API and let it be accessed with certain routes perculiar to it
