package main

import (
	"fmt"

	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/middleware/cors"
	"github.com/kamil5b/go-auth-main/database"
	"github.com/kamil5b/go-auth-main/routes"
	"github.com/kamil5b/go-auth-main/utils"
)

func main() {
	database.Connect()
	app := fiber.New()
	origin := utils.GoDotEnvVariable("VIEW_URL")
	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowOrigins:     origin,
		AllowMethods:     "GET,POST,PUT,DELETE",
	}))

	routes.Setup(app)
	url_server := utils.GoDotEnvVariable("SERVER_URL")
	err := app.Listen(url_server)
	if err != nil {
		fmt.Println(err)
		fmt.Scan()
	}
}
