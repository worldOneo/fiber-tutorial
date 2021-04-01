package main

import "github.com/gofiber/fiber/v2"

func StartFileServer() error {
	fs := fiber.New()
	fs.Static("/", "./public")
	return fs.Listen("localhost:8080")
}
