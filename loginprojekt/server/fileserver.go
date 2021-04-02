package main

import "github.com/gofiber/fiber/v2"

func StartFileServer() {
	fs := fiber.New()
	fs.Static("/", "./public")
	fs.Listen("localhost:3000")
}
