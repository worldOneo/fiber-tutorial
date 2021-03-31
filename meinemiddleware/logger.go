package meinemiddleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

func New() fiber.Handler {
	return func(c *fiber.Ctx) error {
		before := time.Now()
		if err := c.Next(); err != nil {
			return err
		}
		diff := time.Since(before)
		log.Printf("%d | %s | %s | %v", c.Response().StatusCode(),
			c.Method(), c.Path(), diff)
		return nil
	}
}
