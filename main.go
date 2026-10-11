package main

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type Bookmark struct {
	ID    int    `json:"id"`
	URL   string `json:"url" binding:"required"`
	Title string `json:"title" binding:"required, url"`
}

var bookmarks = []Bookmark{}
var nextID = 1

func main() {
	router := gin.Default()

	router.GET("/bookmarks", func(c *gin.Context) {
		c.JSON(200, bookmarks)
	})

	router.GET("/bookmarks/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid ID"})
			return
		}
		for _, bookmark := range bookmarks {
			if id == (bookmark.ID) {
				c.JSON(200, bookmark)
				return
			}
		}

		c.JSON(404, gin.H{"error": "Bookmark not found"})
	})

	router.PUT("/bookmarks/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid ID"})
			return
		}
		var updatedBookmark Bookmark
		if err := c.ShouldBindJSON(&updatedBookmark); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		for i, bookmark := range bookmarks {
			if id == (bookmark.ID) {
				updatedBookmark.ID = bookmark.ID
				bookmarks[i] = updatedBookmark

				c.JSON(200, updatedBookmark)
				return
			}
		}
		c.JSON(404, gin.H{"error": "Bookmark not found"})
	})

	//delete a bookmark
	router.DELETE("/bookmarks/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid ID"})
			return
		}
		for i, bookmark := range bookmarks {
			if id == (bookmark.ID) {
				bookmarks = append(bookmarks[:i], bookmarks[i+1:]...)
				c.JSON(200, gin.H{"message": "Bookmark deleted"})
				return
			}
		}
		c.JSON(404, gin.H{"error": "Bookmark not found"})
	})

	//create a new bookmark
	router.POST("/bookmarks", func(c *gin.Context) {
		var newBookmark Bookmark

		if err := c.ShouldBindJSON(&newBookmark); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		newBookmark.ID = nextID
		nextID++
		bookmarks = append(bookmarks, newBookmark)
		c.JSON(201, newBookmark)
	})

	err := router.Run(":8080")
	if err != nil {
		println("Error starting server:", err)
	}

}
