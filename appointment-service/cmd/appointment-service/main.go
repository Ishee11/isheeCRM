package main

import (
	"appointment-service/internal/app"
	"log"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("app run failed: %v", err)
	}
}
