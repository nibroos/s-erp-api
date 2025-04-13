package config

import (
	"fmt"
	"log"
	"os"
)

func GetDatabaseURL() string {
	env := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
	)
	log.Println("Loading environment variables from .env file", env)

	return env
}

func GetTestDatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER_TEST"),
		os.Getenv("POSTGRES_PASSWORD_TEST"),
		os.Getenv("POSTGRES_HOST_TEST"),
		os.Getenv("POSTGRES_PORT_TEST"),
		os.Getenv("POSTGRES_DB_TEST"),
	)
}
