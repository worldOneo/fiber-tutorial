package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type UserCreateRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthRequest struct {
	Token    string `json:"token"`
	Username string `json:"username"`
}

const userDB = "users.gob.db"
const tokenDB = "tokens.gob.db"

func generateToken(size int) string {
	chars := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRS")
	res := make([]byte, size)
	for i := 0; i < size; i++ {
		res[i] = chars[rand.Intn(len(chars))]
	}
	return string(res)
}

func StartAPIServer(closer chan struct{}) chan struct{} {
	rand.Seed(time.Now().UnixNano())

	api := fiber.New()

	userMap, err := loadDB(userDB)
	if err != nil {
		log.Fatal("load user db:", err)
	}

	tokenMap, err := loadDB(userDB)
	if err != nil {
		log.Fatal("load token db:", err)
	}

	auth := func(c *fiber.Ctx) error {
		if c.Method() != fiber.MethodPost {
			return c.Next()
		}
		authreq := AuthRequest{}
		if err := c.BodyParser(&authreq); err != nil {
			return c.Status(fiber.StatusForbidden).
				JSON(Response{false, "Ein token wird benötigt."})
		}
		token, ok := tokenMap.Get(authreq.Username)

		if !ok || token != authreq.Token {
			return c.Status(fiber.StatusForbidden).
				JSON(Response{false, "Ein token wird benötigt."})
		}

		c.Request().URI().SetPath("/time/cache/t/" + token)
		c.Method(fiber.MethodGet)
		return c.Next()
	}

	cache := cache.New(cache.Config{
		Expiration: 10 * time.Second,
	})

	api.Use(logger.New(),
		cors.New(),
		limiter.New(limiter.Config{
			Max:        120,
			Expiration: time.Minute,
		}))

	apigroup := api.Group("/api") // localhost:6060/api
	v1 := apigroup.Group("/v1")   // localhost:6060/api/v1

	v1.Post("/time", auth, cache, func(c *fiber.Ctx) error {
		time.Sleep(2 * time.Second)
		str := fmt.Sprintf("%d", time.Now().Unix())
		return c.JSON(Response{true, str})
	})

	v1.Post("/createuser", func(c *fiber.Ctx) error {
		ucreq := UserCreateRequest{}
		if err := c.BodyParser(&ucreq); err != nil {
			return c.Status(fiber.StatusBadRequest).
				JSON(Response{false, "Es wird ein username und ein password benötigt."})
		}
		_, ok := userMap.Get(ucreq.Username)
		if ok {
			return c.Status(fiber.StatusUnauthorized).
				JSON(Response{false, "Der user existiert bereits."})
		}
		userMap.Put(ucreq.Username, ucreq.Password)
		return c.Status(fiber.StatusOK).
			JSON(Response{true, "Der user wurde erstell."})
	})

	v1.Post("/generatetoken", func(c *fiber.Ctx) error {
		ucreq := UserCreateRequest{}
		if err := c.BodyParser(&ucreq); err != nil {
			return c.Status(fiber.StatusBadRequest).
				JSON(Response{false, "Es wird ein username und ein password benötigt."})
		}
		pass, ok := userMap.Get(ucreq.Username)
		if !ok || pass != ucreq.Password {
			return c.Status(fiber.StatusForbidden).
				JSON(Response{false, "Der user wurde nicht gefunden oder das passwort stimmt nicht überein!"})
		}
		token := generateToken(64)
		tokenMap.Put(ucreq.Username, token)

		return c.Status(fiber.StatusOK).
			JSON(Response{true, token})
	})

	l, _ := net.Listen("tcp", "localhost:6060")
	go api.Listener(l)
	finisch := make(chan struct{}, 5)
	go func() {
		<-closer
		log.Printf("shutting down the API")
		l.Close()
		err = saveDB(userDB, userMap)
		if err != nil {
			log.Print("saving user db:", err)
		}

		err = saveDB(tokenDB, tokenMap)
		if err != nil {
			log.Print("saving tokens db:", err)
		}

		log.Printf("stopped the api")
		close(finisch)
	}()
	return finisch
}

func loadDB(filename string) (*LockedMap, error) {
	lMap := NewLockedMap()

	if _, err := os.Stat(filename); err != nil {
		return &lMap, nil
	}

	reader, err := os.Open(filename)
	if err != nil {
		return &lMap, err
	}
	dec := gob.NewDecoder(reader)
	err = dec.Decode(&lMap)
	return &lMap, err
}

func saveDB(filename string, lMap *LockedMap) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := gob.NewEncoder(file)
	return enc.Encode(lMap)
}
