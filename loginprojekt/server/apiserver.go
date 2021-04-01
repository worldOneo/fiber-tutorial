package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"time"

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

func StartAPIServer(stop chan struct{}) chan error {
	data, err := loadUserData(userDB)
	errChan := make(chan error, 5)
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

	cacheWare := cache.New(cache.Config{
		Expiration: 10 * time.Second,
	})

	apigroup := api.Group("/api") // localhost:5000/api
	v1 := apigroup.Group("/v1")   // localhost:5000/api/v1

	v1.Get("/time", cacheWare, func(c *fiber.Ctx) error { // localhost:5000/api/v1/time
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
	defer func() {
		if err := saveUserData(userDB, data); err != nil {
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
