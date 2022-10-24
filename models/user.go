package models

import (
	"encoding/json"
	"log"

	conf "api-gateway-android/config"

	"github.com/gomodule/redigo/redis"
)

// User struct
// type User struct {
// 	gorm.Model
// 	Username string `gorm:"unique_index;not null" json:"username"`
// 	Email    string `gorm:"unique_index;not null" json:"email"`
// 	Password string `gorm:"not null" json:"password"`
// 	Names    string `json:"names"`
// }

type User struct {
	Id               string `json:"id" default:"0"`
	Status           string `json:"status" default:"1"`
	Locked           string `json:"locked" default:"0"`
	FailedLoginCount string `json:"failedLoginCount" default:"0"`
	Fullname         string `json:"fullname"`
	Email            string `json:"email"`
	Hp               string `json:"hp"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	CompanyId        string `json:"companyId" default:"0"`
	RoleId           string `json:"roleId" default:"0"`
	OfficeId         string `json:"officeId" default:"0"`
	DepartementId    string `json:"departementId" default:"0"`
	// Total            uint   `json:"-"`
	// MsgError         string `json:"-"`
}

func GetUser(username string) (*User, error) {
	user := &User{}
	keyRedis := "user:profile:" + username
	log.Printf("keyRedis : %v", keyRedis)
	userRedis, err := redis.Bytes(conf.GetReJson().JSONGet(keyRedis, "."))
	if err != nil {
		log.Printf("err : %v", err)
		return user, err
	}

	err = json.Unmarshal(userRedis, &user)
	if err != nil {
		log.Printf("err : %v", err)
		return user, err
	}

	return user, nil
}

// CreateUser new user
func CreateUser(user *User) (*User, error) {
	userData := &User{}
	keyRedis := "user:profile:" + user.Username
	res, err := conf.GetReJson().JSONSet(keyRedis, ".", user)
	if err != nil {
		log.Fatalf("Failed to JSONSet")
		return userData, err
	}

	var err1 error
	if res.(string) == "OK" {

	} else {
		return userData, err1
	}

	return user, nil

}
