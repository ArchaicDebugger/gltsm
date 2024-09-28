package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()
	r.GET("/seed", func(c *gin.Context) {

	})
	r.Run("localhost:4500")
}
