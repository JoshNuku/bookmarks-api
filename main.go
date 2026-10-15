package main

import (
	"database/sql"
	"errors"
	"log"
	"log/slog"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Config struct {
	Port  string
	Env   string
	DBURL string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()
	port := os.Getenv("PORT")
	env := os.Getenv("APP_ENV")
	dbURL := os.Getenv("DATABASE_URL")

	if port == "" {
		port = "8080"
	}
	if env == "" {
		env = "development"
	}
	if dbURL == "" {
		log.Fatal("Database URL not set")
	}
	return &Config{
		Port:  port,
		Env:   env,
		DBURL: dbURL,
	}, nil
}

//struct tags -> ``

type Bookmark struct {
	ID        int      `json:"id"`
	URL       string   `json:"url" binding:"required,url"`
	Title     string   `json:"title" binding:"required,min=1,max=100"`
	CreatedAt string   `json:"created_at"`
	Tags      []string `json:"tags"`
	// Description string `json:"description" binding:"max=200"`
	// Visibility  string `json:"visibility" binding:"required,oneof=public private"`
}

func main() {
	cfg, _ := LoadConfig()
	db := initDB(cfg.DBURL)
	defer db.Close()

	var logLevel slog.Level
	if cfg.Env == "development" {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}
	handlerOpts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewJSONHandler(os.Stdout, handlerOpts)
	logger := slog.New(handler)

	if err := runMigrations(db); err != nil {
		logger.Error("database migration failed", "error", err.Error())
		os.Exit(1)
	}

	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(structuredLogger(logger))
	router.Use(corsMiddleware())

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	//retrieve all bookmarks
	router.GET("/bookmarks", func(c *gin.Context) {
		query := `SELECT id, url, title, created_at FROM bookmarks ORDER BY id DESC`
		rows, err := db.Query(query)
		if err != nil {
			logger.Error("Error fetching bookmarks from database:", "error", err)
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		defer rows.Close()

		bookmarkList := []Bookmark{}

		for rows.Next() {
			var bookmark Bookmark
			if err := rows.Scan(&bookmark.ID, &bookmark.URL, &bookmark.Title, &bookmark.CreatedAt); err != nil {
				logger.Error("Error scanning bookmark row:", "error", err)
				c.JSON(500, gin.H{"error": "Internal server error"})
				return
			}
			bookmarkList = append(bookmarkList, bookmark)
		}
		if err := rows.Err(); err != nil {
			logger.Error("Error iterating over bookmark rows:", "error", err)
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}
		c.JSON(200, bookmarkList)
	})

	//retrieve one bookmark
	router.GET("/bookmarks/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid ID"})
			return
		}
		query := `SELECT id, url, title, created_at FROM bookmarks WHERE id = $1`
		var b Bookmark
		err = db.QueryRow(query, id).Scan(&b.ID, &b.URL, &b.Title, &b.CreatedAt)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(404, gin.H{"error": "Bookmark not found"})
				return
			}

			logger.Error("Error fetching bookmark from database by ID:", "ID", id, "error", err.Error())
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(200, b)
	})

	//update a bookmark
	router.PUT("/bookmarks/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid ID"})
			return
		}
		var updatedBookmark Bookmark
		if err := c.ShouldBindJSON(&updatedBookmark); err != nil {
			c.JSON(400, gin.H{"errors": formatValidationError(err)})
			return
		}

		query := `UPDATE bookmarks SET url = $1, title = $2 WHERE id = $3 RETURNING id, url, title, created_at`
		err = db.QueryRow(query, updatedBookmark.URL, updatedBookmark.Title, id).Scan(&updatedBookmark.ID, &updatedBookmark.URL, &updatedBookmark.Title, &updatedBookmark.CreatedAt)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(404, gin.H{"error": "Bookmark not found"})
				return
			}
			logger.Error("Error updating bookmark in database:", "id", id, "error", err.Error())
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}
		c.JSON(200, updatedBookmark)
	})

	//delete a bookmark
	router.DELETE("/bookmarks/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid ID"})
			return
		}
		query := `DELETE FROM bookmarks WHERE id = $1`
		result, err := db.Exec(query, id)
		if err != nil {
			logger.Error("Error deleting bookmark from database:", "id", id, "error", err.Error())
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			logger.Error("Error getting rows affected by delete query:", "id", id, "error", err.Error())
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}
		if rowsAffected == 0 {
			c.JSON(404, gin.H{"error": "Bookmark not found"})
			return
		}
		c.JSON(200, gin.H{"message": "Bookmark deleted"})
	})

	//create a new bookmark
	router.POST("/bookmarks", func(c *gin.Context) {
		var newBookmark Bookmark

		if err := c.ShouldBindJSON(&newBookmark); err != nil {
			c.JSON(400, gin.H{"errors": formatValidationError(err)})
			return
		}

		query := `INSERT INTO bookmarks (url, title) VALUES ($1, $2) RETURNING id, created_at`

		err := db.QueryRow(query, newBookmark.URL, newBookmark.Title).Scan(&newBookmark.ID, &newBookmark.CreatedAt)
		if err != nil {
			logger.Error("Error inserting bookmark into database:", "error", err)
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}
		c.JSON(201, newBookmark)
	})

	logger.Info("server starting ", "port", cfg.Port, "env", cfg.Env)
	if err := router.Run(":" + cfg.Port); err != nil {
		logger.Error("Error starting server:", "error", err)
		os.Exit(1)
	}

}
