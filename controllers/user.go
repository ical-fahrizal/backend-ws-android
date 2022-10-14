package controllers

import (
	u "api-gateway-android/utils"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	// co "api-gateway-android/controllers"

	conf "api-gateway-android/config"

	"api-gateway-android/app"

	"api-gateway-android/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	uuid "github.com/satori/go.uuid"
)

// User contains the login user information
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type PayloadToken struct {
	Code  string `json:"code"`
	Token string `json:"token"`
}

type UserData struct {
	Nama  string
	Nohp  string
	Uid   string
	Uuid  string
	Token string
	Time  int64
	Rand  int
}

var ValidLogins []User

var CreateUser = func(payload map[string]interface{}) map[string]interface{} {

	resp := u.Message(http.StatusOK, true, http.StatusText(http.StatusOK))

	log.Printf("Payload : %v", payload)
	resp["data"] = payload

	return resp

}

var UpdateUser = func(payload interface{}) map[string]interface{} {

	resp := u.Message(http.StatusOK, true, http.StatusText(http.StatusOK))

	log.Printf("Payload : %v", payload)
	resp["data"] = payload

	return resp

}

var SearchUser = func(payload interface{}) map[string]interface{} {

	resp := u.Message(http.StatusOK, true, http.StatusText(http.StatusOK))

	log.Printf("Payload : %v", payload)
	data := make(map[string]interface{})
	data["username"] = "ical"
	data["role"] = 23
	data["company"] = 12

	resp["data"] = data

	return resp

}

var LoginUser2 = func(ctx *fiber.Ctx) error {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// check for the incoming request body
	body := new(request)
	if err := ctx.BodyParser(&body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse JSON",
		})
	}

	log.Printf("%v", body.Username)
	log.Printf("%v", body.Password)

	if body.Username == "admin" && body.Password == "admin123" {

	} else {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid username & password",
		})
	}

	//Worked! Logged In
	tokenExp := time.Duration(conf.GetTokenExpiration()) * time.Minute

	//jwt dgrijalva
	// create a variable claims for adding expiration token
	claims := conf.Token{
		StandardClaims: jwt.StandardClaims{
			Issuer:    conf.GetAppName(),
			ExpiresAt: time.Now().Add(tokenExp).Unix(),
		},
		// get user id
		Id:          1,
		Name:        "ical",
		Root:        1,
		Office:      12,
		Departement: 3,
		Company:     24,
	}

	//Create JWT token
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), claims)
	tokenString, _ := token.SignedString(conf.GetJwtSecretKey()) // get environmet tokenPassword
	log.Println(tokenString)
	var err error
	keyId := uuid.Must(uuid.NewV4(), err).String()

	room := PayloadToken{
		Code:  keyId,
		Token: tokenString,
	}

	// rooms[room.UUID] = &room
	return ctx.JSON(room)
}

var LoginUser = func(ctx *fiber.Ctx) error {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// check for the incoming request body
	body := new(request)
	if err := ctx.BodyParser(&body); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot parse JSON",
		})
	}

	log.Printf("%v", body.Username)
	log.Printf("%v", body.Password)

	//Database Redis
	user, err := models.GetUser(body.Username)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "username not exist",
		})
	}
	log.Printf("user : %v", user)

	if body.Password == user.Password {

	} else {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid username & password",
		})
	}

	m := UserData{user.Username, user.Hp, "1", "", "", time.Now().Add(20 * time.Minute).Unix(), rand.Intn(10000)}

	b, err := json.Marshal(m)
	if err != nil {
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	t, err := Encrypt(string(b), conf.MySecret)
	if err != nil {
		log.Println("error encrypting your classified text: ", err)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}
	log.Println(t)

	// rooms[room.UUID] = &room
	return ctx.JSON(fiber.Map{"token": t})
}

var VadidationUser = func(ctx *fiber.Ctx) error {
	ctx.GetRespHeader("Content-Type", "application/json")
	tk, err := app.JwtProtected(ctx)
	if err != nil {
		return ctx.JSON(err.Error())
	}

	//Worked! Logged In
	tokenExp := time.Duration(conf.GetTokenExpiration()) * time.Minute

	//jwt dgrijalva
	// create a variable claims for adding expiration token
	claims := conf.Token{
		StandardClaims: jwt.StandardClaims{
			Issuer:    conf.GetAppName(),
			ExpiresAt: time.Now().Add(tokenExp).Unix(),
		},
		// get user id
		Id:          tk.Id,
		Name:        tk.Name,
		Root:        tk.Root,
		Office:      tk.Office,
		Departement: tk.Departement,
		Company:     tk.Company,
	}

	//Create JWT token
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), claims)
	tokenString, _ := token.SignedString(conf.GetJwtSecretKey()) // get environmet tokenPassword
	log.Println(tokenString)

	room := PayloadToken{
		Token: tokenString,
	}

	return ctx.JSON(room)
}

// func findUser(list []User, compareUser *User) bool {
// 	for _, item := range list {
// 		if item.Username == compareUser.Username && item.Password == compareUser.Password {
// 			return true
// 		}
// 	}
// 	return false
// }

func DataUser() {
	ValidLogins = []User{
		{Username: "bob", Password: "test"},
		{Username: "alice", Password: "test"},
		{Username: "ical", Password: "12345"},
		{Username: "salma", Password: "12345"},
		{Username: "zaid", Password: "12345"},
	}
}

func Decrypt(text, MySecret string) (string, error) {
	block, err := aes.NewCipher([]byte(MySecret))
	if err != nil {
		return "", err
	}
	cipherText, errr := Decode(text)
	if errr != nil {
		return "", err
	}
	cfb := cipher.NewCFBDecrypter(block, conf.Bytes)
	plainText := make([]byte, len(cipherText))
	cfb.XORKeyStream(plainText, cipherText)
	return string(plainText), nil
}

func Decode(s string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	// if err != nil {
	// 	panic(err)
	// }
	return data, err
}

func Encrypt(text, MySecret string) (string, error) {
	block, err := aes.NewCipher([]byte(MySecret))
	if err != nil {
		return "", err
	}
	plainText := []byte(text)
	cfb := cipher.NewCFBEncrypter(block, conf.Bytes)
	cipherText := make([]byte, len(plainText))
	cfb.XORKeyStream(cipherText, plainText)
	return Encode(cipherText), nil
}

func Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
