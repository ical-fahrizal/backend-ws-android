package controllers

import (
	conf "api-gateway-android/config"
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2"
)

type DeviceInfo struct {
	PhoneBrand     string `json:"phoneBrand"`
	PhoneType      string `json:"phoneType"`
	PhoneOs        string `json:"phoneOs"`
	PhoneOsVersion string `json:"phoneOsVersion"`
	Time           int64  `json:"time"`
	Rand           int    `json:"rand"`
}

var LoginDevice = func(ctx *fiber.Ctx) error {
	// type request struct {
	// 	Username string `json:"username"`
	// 	Password string `json:"password"`
	// }

	// check for the incoming request body
	body := new(DeviceInfo)
	if err := ctx.BodyParser(&body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse JSON",
		})
	}

	log.Printf("%v", body.PhoneBrand)
	log.Printf("%v", body.PhoneType)

	// m := UserData{user.Username, user.Hp, "1", "", "", time.Now().Add(20 * time.Minute).Unix(), rand.Intn(10000)}
	m := DeviceInfo{body.PhoneBrand,
		body.PhoneType,
		body.PhoneOs,
		body.PhoneOsVersion,
		time.Now().Add(20 * time.Minute).Unix(),
		rand.Intn(10000),
	}

	log.Printf("m.Time : %v", m.Time)
	log.Printf("m.Rand : %v", m.Rand)
	b, err := json.Marshal(m)
	if err != nil {
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	t, err := Encrypt(string(b), conf.MySecret)
	if err != nil {
		log.Println("error encrypting your classified text: ", err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}
	// log.Println(t)

	decText, _ := Decrypt(t, conf.MySecret)
	log.Printf("decText : %v", decText)

	// rooms[room.UUID] = &room
	return ctx.JSON(fiber.Map{"token": t})
}
