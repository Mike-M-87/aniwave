package main

import (
	"aniwave/notifications"
	"aniwave/utils"
	"os"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

var defaultPort = "8082"

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Panic().Err(err)
	}
	utils.InitialiseDB()

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	app := fiber.New(fiber.Config{
		Prefork: false,
	})
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))

	app.Static("/", "./index.html")
	app.Post("/done", notifications.ChangeDone)
	app.Get("/nots", notifications.DisplayNotifications)
	app.Get("/check", notifications.CheckNotifications)

	app.Use(func(c *fiber.Ctx) error {
		return c.Status(418).JSON(&fiber.Map{
			"Message": "🍏 Route not found",
		}) // => 418 "I am a tepot"
	})

	log.Fatal().Err(app.Listen(":" + port))
}
