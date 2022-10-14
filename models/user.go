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
	// Id               uint   `json:"id"`
	// Status           bool   `json:"status"`
	Locked           bool   `json:"locked" default:"false"`
	FailedLoginCount int    `json:"failedLoginCount" default:"0"`
	Fullname         string `json:"fullname"`
	Email            string `json:"email"`
	Hp               string `json:"hp"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	// FcmToken         string `json:"fcmToken"`
	// ApiKey           string `json:"apiKey"`
	// CompanyId        int    `json:"companyId"`
	// ParentUserId     int    `json:"parentUserId"`
	// RootId           int    `json:"rootId"`
	// OfficeId         int    `json:"officeId"`
	// DepartementId    int    `json:"departementId"`
	// LastLoginTime    string `json:"lastLoginTime"`
	// Token            string `json:"token"`
	// Total            uint   `json:"-"`
	// MsgError         string `json:"-"`
}

func GetUser(username string) (*User, error) {
	user := &User{}
	keyRedis := "user:" + username
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
	keyRedis := "user:" + user.Username
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
