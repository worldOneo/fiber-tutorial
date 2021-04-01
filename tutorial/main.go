package main

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/worldOneo/fiber-tutorial/meinemiddleware"
)

func main() {
	app := fiber.New()

	app.Use(meinemiddleware.New()) // #4 Middleware

	limit := limiter.New(limiter.Config{
		Max:        1,
		Expiration: time.Second,
	})

	// Part 1 - Die Basics
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Send([]byte("Hello, World!"))
	})

	// Part 2 - Daten auf dem server empfangen
	app.Get("/name/:name", func(c *fiber.Ctx) error {
		name := c.Params("name")
		response := fmt.Sprintf("Hello, %s!", name)
		return c.Send([]byte(response))
	})

	app.Get("/name/:name/:greeting", func(c *fiber.Ctx) error {
		name := c.Params("name")
		greeting := c.Params("greeting")
		response := fmt.Sprintf("%s, %s!", greeting, name)
		return c.Send([]byte(response))
	})

	type Video struct {
		VideoID string `query:"v"`
	}

	app.Get("/parameter", func(c *fiber.Ctx) error {
		video := Video{}
		if err := c.QueryParser(&video); err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		str := fmt.Sprintf("Dein Video ist %s", video.VideoID)
		return c.Send([]byte(str))
	})

	type Status struct {
		Message string `json:"message"`
		Online  bool   `json:"online"`
	}

	app.Post("/postdata", func(c *fiber.Ctx) error {
		status := Status{}
		if err := c.BodyParser(&status); err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}
		str := fmt.Sprintf("Online: %v; Nachricht: %v", status.Online, status.Message)
		return c.Send([]byte(str))
	})

	// Part 3 - Der Server antwortet
	app.Get("/forbidden", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusForbidden).
			Send([]byte("Hier darfst du nicht hin!"))
	})

	type Response struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	app.Get("/response/success", limit, func(c *fiber.Ctx) error {
		return c.JSON(Response{true, "Alles yut"})
	})

	app.Get("/response/fail", limit, func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusInternalServerError).
			JSON(Response{false, "Datenbank hat nicht geantwortet!"})
	})

	//Part 4 - Middleware

	// Part 5 - Cookies

	// Hinzufügen
	app.Get("/cookies/add", func(c *fiber.Ctx) error {
		c.Cookie(&fiber.Cookie{
			Name:     "MeinCookie",
			Value:    fmt.Sprintf("%d", time.Now().Unix()),
			HTTPOnly: true,
			Expires:  time.Now().Add(time.Hour * 10),
		})
		return c.Send([]byte("Du hast einen frisch gebackenen Keks erhalten!"))
	})

	// Prüfen
	app.Get("/cookies/check", func(c *fiber.Ctx) error {
		cookie := c.Request().Header.Cookie("MeinCookie")
		str := fmt.Sprintf("Dein Keks ist von %s", string(cookie))
		return c.Send([]byte(str))
	})

	// Entfernen
	app.Get("/cookies/remove", func(c *fiber.Ctx) error {
		c.Cookie(&fiber.Cookie{
			Name:     "MeinCookie",
			HTTPOnly: true,
			Expires:  time.Now().Add(-(time.Hour * 1000)),
		})
		return c.Send([]byte("Dein Keks wurde dir weggenommen!"))
	})

	// Part 6 - JWT

	// Erstellen
	app.Get("/jwt/add", func(c *fiber.Ctx) error {
		token := jwt.New(jwt.SigningMethodHS256)
		token.Claims.(jwt.MapClaims)["date"] = time.Now().Unix()
		str, err := token.SignedString([]byte("geheimniss"))
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		c.Cookie(&fiber.Cookie{
			Name:     "MeinJWT",
			Value:    str,
			HTTPOnly: true,
			Expires:  time.Now().Add(time.Hour * 10),
		})
		return c.SendStatus(fiber.StatusOK)
	})

	// Prüfen
	app.Get("/jwt/check", func(c *fiber.Ctx) error {
		cookie := c.Request().Header.Cookie("MeinJWT")
		token, err := jwt.Parse(string(cookie), func(t *jwt.Token) (interface{}, error) {
			return []byte("geheimniss"), nil
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).
				Send([]byte("Der Token konnte nicht Verifiziert werden!"))
		}
		str := fmt.Sprintf("Dein Token ist von %f", token.Claims.(jwt.MapClaims)["date"].(float64))
		return c.Status(fiber.StatusOK).
			Send([]byte(str))
	})
	app.Listen("localhost:3000")
}
