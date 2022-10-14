package app

import (
	"errors"
	"log"
	"strings"

	conf "api-gateway-android/config"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

func JwtProtected(c *fiber.Ctx) (*conf.Token, error) {
	tokenHeader := c.Get("Authorization")
	// resp := u.Message(http.StatusOK, true, "success")
	tk := &conf.Token{}
	//resp := u.Message(http.StatusOK, true, "success")
	if tokenHeader == "" { //Token is missing, returns with error code 403 Unauthorized
		// resp := u.Message(http.StatusForbidden, false, "Missing auth token")
		return tk, errors.New("Missing auth token")
	}

	splitted := strings.Split(tokenHeader, " ") //The token normally comes in format `Bearer {token-body}`, we check if the retrieved token matched this requirement
	tokenPart := splitted[1]                    //Grab the token part, what we are truly interested in

	log.Printf("tokenPart : %v", tokenPart)
	token, err := jwt.ParseWithClaims(tokenPart, tk, func(token *jwt.Token) (interface{}, error) {
		return conf.GetJwtSecretKey(), nil
	})

	if len(splitted) != 2 {
		// resp := u.Message(http.StatusForbidden, false, "Invalid/Malformed auth token")
		return tk, errors.New("Invalid/Malformed auth token")
	}
	log.Printf("Toke Valid %v", token.Valid)
	if err != nil { //Malformed token, returns with http code 403 as usual
		// resp := u.Message(http.StatusForbidden, false, "Your token has been expired, please logout first and re login to access this menu!")
		return tk, errors.New("Your token has been expired, please logout first and re login to access this menu!")
	}

	if !token.Valid { //Token is invalid, maybe not signed on this server
		// resp := u.Message(http.StatusForbidden, false, "Token is not valid.")
		return tk, errors.New("Your token has been expired, please logout first and re login to access this menu!")
	}

	//Everything went well, proceed with the request and set the caller to the user retrieved from the parsed token
	userContext := &conf.UserContext{}
	userContext.Id = tk.Id
	userContext.Name = tk.Name
	userContext.Root = tk.Root
	userContext.Office = tk.Office
	userContext.Departement = tk.Departement
	userContext.Company = tk.Company

	log.Printf("Company : %v", userContext.Company)
	// ctx := context.WithValue(c.Context(), "user_context", userContext)
	// c.SetUserContext(ctx)
	// resp["data"] = userContext

	return tk, nil
}
