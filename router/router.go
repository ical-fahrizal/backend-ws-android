package router

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	conf "api-gateway-android/config"
	cont "api-gateway-android/controllers"

	"github.com/antoniodipinto/ikisocket"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

var roomClient = make(map[string][]string)
var clients map[string]UserData

// var requestClient = make(map[string][]string)

const MySecret string = "abc&1*~#^2^#s0^=)^^7%b34"

type MessageKafka struct {
	Topic string      `json:"topic"`
	Proc  string      `json:"proc"`
	Data  interface{} `json:"data"`
	From  string      `json:"-"`
	Exp   int64       `json:"exp"`
}

type MessageWebsocket struct {
	Type  int         `json:"type"`
	Topic string      `json:"topic"`
	Proc  string      `json:"proc"`
	Data  interface{} `json:"data"`
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

// SetupRoutes setup router api
func SetupRoutes() {

	// Start a new Fiber application
	app := fiber.New()

	//send info device
	app.Post("/device-info", cont.LoginDevice)

	//Login
	app.Post("/login", cont.LoginUser)

	// Setup the middleware to retrieve the data sent in first GET request
	app.Use(func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	setupSocketListeners()

	app.Get("/ws", ikisocket.New(func(kws *ikisocket.Websocket) {

	}))

	go kafkaConsumer()

	log.Fatal(app.Listen(":3002"))

}

func setupSocketListeners() {
	clients = make(map[string]UserData)

	ikisocket.On(ikisocket.EventConnect, func(ep *ikisocket.EventPayload) {
		fmt.Printf("Connection event - User: %s with UUID: %s\n", ep.Kws.GetStringAttribute("user_id"), ep.Kws.UUID)
	})

	ikisocket.On(ikisocket.EventMessage, func(ep *ikisocket.EventPayload) {

		if string(ep.Data) == "9" {
			ep.Kws.Fire("ping", ep.Data)
		} else {
			req := MessageKafka{}

			err := json.Unmarshal(ep.Data, &req)
			if err != nil {
				log.Printf("error 1 : %v", err)

				m := MessageWebsocket{3, req.Topic, req.Proc, "JSON Invalid"}

				msge, err := json.Marshal(m)
				if err != nil {
					fmt.Printf("Error : %s\n", err)
				}
				ep.Kws.Emit(msge)
				ep.Kws.Close()
				return
			}

			log.Printf("req.Topic : %v", req.Topic)
			if req.Topic == "login" {
				isLoggedIn := checkLogin(req.Data, ep)
				if !isLoggedIn {
					// ep.Kws.Emit([]byte("disconnected 2"))
					ep.Kws.Close()
					return
				}
			} else if req.Topic == "deviceInfo" {
				isLoggedIn := checkDevice(req.Data, ep)
				if !isLoggedIn {
					// ep.Kws.Emit([]byte("disconnected 2"))
					ep.Kws.Close()
					return
				}
			} else if req.Topic == "non-login" {
				vKeyTime := int64(ep.Kws.GetIntAttribute("keyTime"))

				if vKeyTime != 0 {
					req.Exp = time.Now().Add(20 * time.Minute).Unix()
					req.From = ep.Kws.UUID

					m := MessageWebsocket{2, req.Topic, req.Proc, "OK"}

					msge, err := json.Marshal(m)
					if err != nil {
						fmt.Printf("Error : %s\n", err)
					}

					ep.Kws.Emit(msge)
				} else {
					// ep.Kws.Emit([]byte("disconnected"))
					m := MessageWebsocket{3, req.Topic, req.Proc, "Please Insert info device "}

					msge, err := json.Marshal(m)
					if err != nil {
						fmt.Printf("Error : %s\n", err)
					}

					ep.Kws.Emit(msge)
					ep.Kws.Close()
					return
				}
			} else {
				if ep.Kws.GetStringAttribute("isLoggedIn") == "true" {
					req.Exp = time.Now().Add(20 * time.Minute).Unix()
					req.From = ep.Kws.UUID

					m := MessageWebsocket{2, req.Topic, req.Proc, "OK"}

					msge, err := json.Marshal(m)
					if err != nil {
						fmt.Printf("Error : %s\n", err)
					}

					ep.Kws.Emit(msge)

					kafkaProducer(req)
				} else {
					// ep.Kws.Emit([]byte("disconnected"))
					m := MessageWebsocket{3, req.Topic, req.Proc, "Please Login"}

					msge, err := json.Marshal(m)
					if err != nil {
						fmt.Printf("Error : %s\n", err)
					}

					ep.Kws.Emit(msge)
					ep.Kws.Close()
					return
				}
			}
		}
	})

	ikisocket.On(ikisocket.EventDisconnect, func(ep *ikisocket.EventPayload) {
		delete(clients, ep.Kws.GetStringAttribute("token"))
		fmt.Printf("Disconnection event - User: %s with UUID: %s\n", ep.Kws.GetStringAttribute("user_id"), ep.Kws.UUID)
	})

	ikisocket.On(ikisocket.EventClose, func(ep *ikisocket.EventPayload) {
		delete(clients, ep.Kws.GetStringAttribute("user_id"))
		fmt.Println("Close event - User: ", ep.Kws.GetStringAttribute("user_id"))
	})

	ikisocket.On(ikisocket.EventError, func(ep *ikisocket.EventPayload) {
		fmt.Printf("Error event - User: %s with UUID: %s Error : %s\n", ep.Kws.GetStringAttribute("user_id"), ep.Kws.UUID, ep.Error)
	})
}

func SendWs(data *conf.ResponKafka) {

	log.Printf("SendWs : %v", data.Token)
	messageWs := conf.MessageWs{}
	messageWs.Message = "SUCCESS"
	messageWs.Status = true
	messageWs.Data = data
	sRespon, _ := json.Marshal(data)
	// ikisocket.EmitTo(clients[data.Token], sRespon)
	ikisocket.EmitToList(roomClient[data.Token], sRespon)

}

func checkLogin(str interface{}, ep *ikisocket.EventPayload) bool {

	encText := fmt.Sprintf("%v", str)
	decText, err := Decrypt(encText, MySecret)
	if err != nil {
		fmt.Println("error decrypting your encrypted text: ", err)
		fmt.Println("disini user unknown")
		m := MessageWebsocket{3, "", "", "System Error"}

		msge, err := json.Marshal(m)
		if err != nil {
			fmt.Printf("Error : %s\n", err)
		}

		ep.Kws.Emit(msge)
		return false
	} else {
		if decText != "" {
			var m UserData
			errr := json.Unmarshal([]byte(decText), &m)
			if errr != nil {
				fmt.Println("error rubah ke obj dr json: ", err)
			} else {
				if m.Nama != "John Doe" && m.Uid != "1" {
					fmt.Println("disini user unknown")
					m := MessageWebsocket{3, "", "", "User unknown"}

					msge, err := json.Marshal(m)
					if err != nil {
						fmt.Printf("Error : %s\n", err)
					}

					ep.Kws.Emit(msge)

					return false
				}
				if m.Time < time.Now().Unix() {
					m := MessageWebsocket{3, "", "", "Token is Expired"}

					msge, err := json.Marshal(m)
					if err != nil {
						fmt.Printf("Error : %s\n", err)
					}

					ep.Kws.Emit(msge)
					return false
				}

				ep.Kws.SetAttribute("user_id", m.Uid)
				ep.Kws.SetAttribute("token", encText)
				ep.Kws.SetAttribute("isLoggedIn", "true")

				m.Token = encText
				m.Uuid = ep.Kws.UUID
				log.Printf("m : %v", m)
				log.Printf("encText : %v", encText)
				clients[encText] = m

				log.Printf("clients : %v", clients)
				m := MessageWebsocket{4, "", "", "OK"}

				msge, err := json.Marshal(m)
				if err != nil {
					fmt.Printf("Error : %s\n", err)
				}

				ep.Kws.Emit(msge)

				return true
			}
		}
	}
	return false
}

func checkDevice(str interface{}, ep *ikisocket.EventPayload) bool {

	encText := fmt.Sprintf("%v", str)
	decText, err := Decrypt(encText, MySecret)
	log.Printf("decText : %v", decText)

	if err != nil {
		fmt.Println("error decrypting your encrypted text: ", err)
		fmt.Println("disini user unknown")
		m := MessageWebsocket{3, "", "", "System Error"}

		msge, err := json.Marshal(m)
		if err != nil {
			fmt.Printf("Error : %s\n", err)
		}

		ep.Kws.Emit(msge)
		return false
	} else {
		if decText != "" {
			var m cont.DeviceInfo
			errr := json.Unmarshal([]byte(decText), &m)
			if errr != nil {
				fmt.Println("error rubah ke obj dr json: ", err)
			} else {
				log.Printf("time : %v", m.Time)
				if m.Time < time.Now().Unix() {
					m := MessageWebsocket{3, "", "", "Token is Expired"}

					msge, err := json.Marshal(m)
					if err != nil {
						fmt.Printf("Error : %s\n", err)
					}

					ep.Kws.Emit(msge)
					return false
				}

				ep.Kws.SetAttribute("phoneBrand", m.PhoneBrand)
				ep.Kws.SetAttribute("phoneType", m.PhoneType)
				ep.Kws.SetAttribute("phoneOs", m.PhoneOsVersion)
				ep.Kws.SetAttribute("phoneOsVersion", m.PhoneOsVersion)
				ep.Kws.SetAttribute("keyTime", int(m.Time))
				ep.Kws.SetAttribute("token", encText)
				ep.Kws.SetAttribute("isLoggedIn", "false")

				log.Printf("clients : %v", clients)
				m := MessageWebsocket{4, "", "", "OK"}

				msge, err := json.Marshal(m)
				if err != nil {
					fmt.Printf("Error : %s\n", err)
				}

				ep.Kws.Emit(msge)

				return true
			}
		}
	}
	return false
}

func Decrypt(text interface{}, MySecret string) (string, error) {
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

func Decode(str interface{}) ([]byte, error) {
	s := fmt.Sprintf("%v", str)
	data, err := base64.StdEncoding.DecodeString(s)
	// if err != nil {
	// 	panic(err)
	// }
	return data, err
}

func kafkaProducer(mg MessageKafka) {

	// p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "192.168.10.156"})
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": conf.GetBroker()})
	if err != nil {
		panic(err)
	}

	defer p.Close()

	// Delivery report handler for produced messages
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	word, err := json.Marshal(mg)
	if err != nil {
		fmt.Printf("Error : %s\n", err)
	}

	// Produce messages to topic (asynchronously)
	topic := mg.Topic + "-request"
	p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: 0},
		Value:          word,
	}, nil)

	// for _, word := range []string{"Welcome", "to", "the", "Confluent", "Kafka", "Golang", "client"} {
	// 	p.Produce(&kafka.Message{
	// 		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: 0},
	// 		Value:          []byte(word),
	// 	}, nil)
	// }

	// Wait for message deliveries before shutting down
	p.Flush(1 * 1000)
}

func kafkaConsumer() {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": conf.GetBroker(),
		"group.id":          conf.GetGroup(),
		"auto.offset.reset": "latest",
	})

	if err != nil {
		panic(err)
	}

	run := true
	topics := []string{conf.GetTopicRequest()}
	err = c.SubscribeTopics(topics, nil)

	for run {
		ev := c.Poll(0)
		switch e := ev.(type) {
		case *kafka.Message:
			// fmt.Printf("%% Message on %s:\n%s\n", e.TopicPartition, string(e.Value))
			var msgDrKafka MessageKafka
			errr := json.Unmarshal([]byte(string(e.Value)), &msgDrKafka)
			if errr != nil {
				fmt.Printf("Error : %s\n", errr)
			} else {
				if msgDrKafka.Exp < time.Now().Unix() {
					fmt.Printf("Message Expired: %v", msgDrKafka)
				} else {
					log.Printf("msgDrKafka.From : %v", msgDrKafka.From)
					m := MessageWebsocket{10, msgDrKafka.Topic, msgDrKafka.Proc, msgDrKafka.Data}
					msge, errrr := json.Marshal(m)
					if errrr != nil {
						fmt.Printf("Error : %s\n", err)
					} else {
						// fmt.Printf("Sent : %s\n", msge)
						ikisocket.EmitTo(msgDrKafka.From, msge)
					}
				}
			}
		case kafka.PartitionEOF:
			fmt.Printf("%% Reached %v\n", e)
		case kafka.Error:
			fmt.Fprintf(os.Stderr, "%% Error: %v\n", e)
			run = false
		default:
			if e != nil {
				fmt.Printf("Ignored %v\n", e)
			}
		}
	}
	c.Close()
}
