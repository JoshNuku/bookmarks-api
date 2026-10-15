package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func formatValidationError(err error) map[string]string {
	var ve validator.ValidationErrors

	// Check if the error is actually a validator.ValidationErrors
	if errors.As(err, &ve) {
		out := make(map[string]string)

		// Loop through every field that failed validation
		for _, fe := range ve {
			out[fe.Field()] = msgForTag(fe)
		}
		return out
	}

	// Fallback if the error wasn't validation (e.g., malformed JSON syntax)
	return map[string]string{"error": err.Error()}
}

// msgForTag returns a human-friendly message based on which rule failed
func msgForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "url":
		return "Must be a valid URL"
	case "email":
		return "Must be a valid email address"
	case "min":
		return fmt.Sprintf("Must be at least %s characters long", fe.Param())
	case "max":
		return fmt.Sprintf("Must not exceed %s characters", fe.Param())
	case "oneof":
		return fmt.Sprintf("Must be one of: %s", fe.Param())
	default:
		return "Invalid value"
	}
}

func initDB(dbURL string) *sql.DB {
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)

	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("Error pinging the database: %v", err)
	}
	return db
}
