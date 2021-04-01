package main

import (
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/worldOneo/fiber-tutorial/loginprojekt/util"
)

type Response struct {
	Success bool `json:"success"`
}

type MessageResponse struct {
	Response
	Message string `json:"message"`
}

type UserCreateRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type TokenRequest struct {
	Token string `json:"token"`
}

//https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0192384765"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func StartAPIServer(stop chan struct{}) chan error {
	rand.Seed(time.Now().UnixNano())
	errChan := make(chan error, 5)

	data, err := loadUserData(userDB)
	if err != nil {
		errChan <- err
		return errChan
	}

	tokens, err := loadUserData(tokenDB)
	if err != nil {
		errChan <- err
		return errChan
	}

	api := fiber.New()

	api.Use(
		logger.New(),
		cors.New(),
		limiter.New(limiter.Config{
			Max:        120,
			Expiration: time.Minute,
		}))

	authWare := func(c *fiber.Ctx) error {
		req := TokenRequest{}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).
				JSON(createMessageResponse(false, "Hierzu brauchst du ein token!"))
		}

		token, err := jwt.Parse(req.Token, func(t *jwt.Token) (interface{}, error) {
			claims, ok := t.Claims.(jwt.MapClaims)
			if !ok {
				log.Print("jwt err not a map")
				return "", errors.New("ungültiger token")
			}

			name, ok := claims["name"]
			if !ok {
				return "", errors.New("kein name im token definiert")
			}

			if _, ok := name.(string); !ok {
				return "", errors.New("kein name im token definiert")
			}

			secret, ok := tokens.Get(name.(string))
			if !ok {
				return "", errors.New("der user existiert nicht")
			}

			return []byte(secret), nil
		})

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).
				JSON(createMessageResponse(false, err.Error()))
		}

		date, ok := token.Claims.(jwt.MapClaims)["expr"]
		if !ok {
			return c.Status(fiber.StatusInternalServerError).
				JSON(createMessageResponse(false, "Kein gültiges ablauf datum"))
		}

		if fdate, ok := date.(float64); !ok || fdate < float64(time.Now().Unix()) {
			return c.Status(fiber.StatusInternalServerError).
				JSON(createMessageResponse(false, "Der token ist abgelaufen"))
		}

		secret, _ := tokens.Get(token.Claims.(jwt.MapClaims)["name"].(string))

		c.Request().URI().SetPath("/api/v1/time/cached/" + secret)
		c.Method(fiber.MethodGet)
		return c.Next()
	}

	gen := func(c *fiber.Ctx) string {
		return string(c.Request().URI().Path())
	}

	cacheWare := cache.New(cache.Config{
		KeyGenerator: gen,
		Key:          gen,
		Expiration:   10 * time.Second,
	})

	apigroup := api.Group("/api") // localhost:5000/api
	v1 := apigroup.Group("/v1")   // localhost:5000/api/v1

	v1.Post("/time", authWare, cacheWare, func(c *fiber.Ctx) error { // localhost:5000/api/v1/time
		time.Sleep(time.Second * 2)
		str := fmt.Sprintf("%d", time.Now().Unix())
		return c.Status(fiber.StatusOK).Send([]byte(str))
	})

	v1.Post("/createuser", func(c *fiber.Ctx) error {
		ucreq := UserCreateRequest{}
		if err := c.BodyParser(&ucreq); err != nil {
			return c.Status(fiber.StatusBadRequest).
				JSON(createMessageResponse(false, "Es wird ein username und password benötigt"))
		}

		if _, ok := data.Get(ucreq.Username); ok {
			return c.Status(fiber.StatusForbidden).
				JSON(createMessageResponse(false, "Dieser user existiert bereits."))
		}
		data.Put(ucreq.Username, ucreq.Password)
		return c.Status(fiber.StatusOK).
			JSON(createMessageResponse(true, "Der user wurde erstellt"))
	})

	v1.Post("/generatetoken", func(c *fiber.Ctx) error {
		ucreq := UserCreateRequest{}
		if err := c.BodyParser(&ucreq); err != nil {
			return c.Status(fiber.StatusBadRequest).
				JSON(createMessageResponse(false, "Es wird ein username und password benötigt"))
		}

		if pass, ok := data.Get(ucreq.Username); !ok || pass != ucreq.Password {
			return c.Status(fiber.StatusUnauthorized).
				JSON(createMessageResponse(false, "Ungültige nutzer daten."))
		}

		secret := RandStringBytes(1024)
		tokens.Put(ucreq.Username, secret)

		token := jwt.New(jwt.SigningMethodHS256)
		token.Claims.(jwt.MapClaims)["expr"] = time.Now().Add(time.Hour * 7 * 24).Unix()
		token.Claims.(jwt.MapClaims)["name"] = ucreq.Username

		str, err := token.SignedString([]byte(secret))
		if err != nil {
			log.Print("signing token: ", str)
			return c.Status(fiber.StatusInternalServerError).
				JSON(createMessageResponse(false, "Ein fehler ist aufgetreten!"))
		}

		return c.Status(fiber.StatusOK).
			JSON(createMessageResponse(true, str))
	})

	defer func() {
		if err := saveUserData(userDB, data); err != nil {
			log.Print("saving user data:", err)
		}
		if err := saveUserData(tokenDB, tokens); err != nil {
			log.Print("saving user data:", err)
		}
	}()

	go func() {
		errChan <- api.Listen("localhost:5000")
		close(errChan)
	}()

	go func() {
		<-stop
		errChan <- saveUserData(userDB, data)
		api.Shutdown()
	}()
	return errChan
}

const userDB = "users.gob"
const tokenDB = "tokens.gob"

func loadUserData(name string) (*util.LockedMap, error) {
	mp := util.New()
	if _, err := os.Stat(name); err != nil {
		return &mp, nil
	}

	file, err := os.Open(name)

	if err != nil {
		return &mp, err
	}

	gobdec := gob.NewDecoder(file)
	err = gobdec.Decode(&mp)
	return &mp, err
}

func saveUserData(name string, mp *util.LockedMap) error {

	file, err := os.OpenFile(name, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}

	gobdec := gob.NewEncoder(file)
	return gobdec.Encode(mp)
}

func createMessageResponse(success bool, msg string) MessageResponse {
	return MessageResponse{Response{success}, msg}
}
