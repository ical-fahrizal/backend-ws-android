package utils

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	conf "api-gateway-android/config"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const (
	// TIME_LAYOUT                 = "2006-01-02 15:04:05-07"
	// TIME_LAYOUT_DB              = "2006-01-02T15:04:05-07:00"
	TIME_LAYOUT_DB_WO_TZ = "2006-01-02 15:04:05"
	// TIME_LAYOUT_DATE            = "2006-01-02"
	// TIME_LAYOUT_DATE_SHORT      = "06-01-02"
	// TIME_LAYOUT_DATE_MONTH_AND_YEAR = "2006-01"
	// TIME_LAYOUT_DATE_EXCEL      = "2-Jan-06"
	// TIME_LAYOUT_DATETIME_EXCEL  = "1/2/06 15:04"
	// ERROR_HANDLING_PERMISSION_1 = "Failed 001, you don't have permission to access this module"
	// ERROR_HANDLING_PERMISSION_2 = "Failed 002, you don't have permission to access this module"
	// ERROR_HANDLING_PERMISSION_3 = "Failed 003, you don't have permission to access this module"
	// ERROR_HANDLING_PERMISSION_4 = "Failed 004, you don't have permission to access this module"
	// ERROR_HANDLING_PERMISSION_5 = "Failed 005, you don't have permission to access this module"
	ERROR_HANDLING_CREATE = "Failed 006, cannot insert into database"
	ERROR_HANDLING_UPDATE = "Failed 006, cannot update data to database"
	ERROR_HANDLING_DELETE = "Failed 005, cannot delete data from database"
	ERROR_HANDLING_READ   = "Failed 005, cannot read data from database"
	// LATITUDE_REGEX_STRING       = "^[-+]?([1-8]?\\d(\\.\\d+)?|90(\\.0+)?)$"
	// LONGITUDE_REGEX_STRING      = "^[-+]?(180(\\.0+)?|((1[0-7]\\d)|([1-9]?\\d))(\\.\\d+)?)$"
	RE_LEADCLOSE_WHTSP = `^[\s\p{Zs}]+|[\s\p{Zs}]+$`
	RE_INSIDE_WHTSP    = `[\s\p{Zs}]{2,}`
	// RE_POST_CODE                = `\d{5}`
	LENGTH_DATA = 20
	// MODULE_USER                 = "user"
	// MODULE_AREA                 = "master-area"
	// MODULE_CHECK_IN_OUT         = "checkinout"
	// MODULE_COMPANY              = "company"
	// MODULE_COUNTRY              = "country"
	// MODULE_CUSTOMER             = "customer"
	// MODULE_DELIVERY_ORDER       = "shipment-order"
	// MODULE_DRIVER               = "driver"
	// MODULE_DRIVER_ASSIGNMENT    = "company"
	// MODULE_DRIVER_LICENSE       = "master-driver-license"
	// MODULE_DROP_POINT           = "shipment-order"
	// MODULE_INSPECT              = "master-inspection"
	// MODULE_POOL                 = "pool"
	// MODULE_POOL_GROUP           = "pool-group"
	// MODULE_PRODUCT              = "product"
	// MODULE_SALES                = "sales"
	// MODULE_SALES_ORDER          = "sales-order"
	// MODULE_SECURITY_PERSON      = "security-person"
	// MODULE_SHIPMENT_ORDER       = "shipment-order"
	// MODULE_TRANSIT_POINT        = "customer-point"
	// MODULE_TRANSIT_POINT_GROUP  = "group-customer-point"
	// MODULE_VEHICLE              = "vehicle"
	// MODULE_VEHICLE_LICENSE      = "master-vehicle-license"
	// MODULE_VEHICLE_TYPE         = "vehicle-type"
	// MODULE_VENDOR               = "vendor"
	// MODULE_VTDS                 = "gps"
	// MODULE_REPORT_CALCULATE     = "company"
	// MODULE_RESHARE_LOG          = "reshare-log"
	// MODULE_GEOFENCE             = "geofence"
	// MODULE_ROUTE                = "company"
	// MODULE_ROUTE_DETAIL         = "company"
	// MODULE_REPORT_SPATIAL       = "report-spatial"
	// MODULE_REPORT_SENSOR        = "report-sensor"
	// MODULE_PRICING              = "transportation-cost"
	// MODULE_TRANSPORTATION_FEE   = "delivery-cost"
	// MODULE_UNIT_TYPE            = "company"
	// MODULE_REPORT_SUMMARY       = "report-summary"
	// MODULE_ROLE                 = "role"
	// MODULE_PACKAGE              = "package-menu"
	// MODULE_MENU                 = "master-menu"
	// MODULE_NOTIFICATION         = "company"
	// MODULE_VEHICLE_DISPOSAL     = "company"
	// MODULE_VEHICLE_WRITE_OFF    = "company"
	// MODULE_VEHICLE_TRANSFER     = "company"
	// MODULE_VEHICLE_ACCIDENT     = "company"
	// MODULE_POINT                = "company"
	// MODULE_TOP_UP_POINT         = "company"
	// MODULE_HISTORY_POINT        = "company"
	// MODULE_SERVICE_TYPE         = "company"
	// MODULE_SEAL                 = "company"
	// MODULE_TARGET_OTA           = "company"
	// MODULE_TARGET_OTLOTD        = "company"
	// MODULE_TYPE_LOCATION        = "type-location"
	// MODULE_VEHICLE_MAINTENANCE_LITE   = "company"
	// MODULE_HEAVY_EQUIPMENT_ACTIVITIES = "company"
	// MODULE_BUM_VEHICLE_ACTIVITY = "company"
	// MODULE_COMPANY_ACCESS       = "company"
	// API_MASKING_GMS_OLD         = "http://103.16.199.187"
	// API_MASKING_GMS             = "http://sms.mysmsmasking.com"
	// API_OSMAP                   = "http://192.168.70.210:81"
	// API_EPOOOL                  = "https://app.epoool.id"
	// API_ADDRESS                 = "http://192.168.70.200/nominatim"
	// API_AMARIS                  = "https://vrp.amarisai.org"
	// API_NOMINATIM_OPEN_STREET   = "https://nominatim.openstreetmap.org"
	// API_KRAKATAU                = "http://114.4.26.114"
	// API_PAXEL_ADDRESS           = "https://paxel.oslogdev.com"
	// API_QUINCUS_ADDRESS         = "https://api.dataxchange.poc.quincus.com"

)

func Message(code uint, status bool, message string) map[string]interface{} {
	if strings.TrimSpace(strings.ToLower(message)) == "success" {
		return map[string]interface{}{"code": code, "status": status, "message": http.StatusText(http.StatusOK)}
	} else {
		return map[string]interface{}{"code": code, "status": status, "message": message}
	}
}

func Respond(w http.ResponseWriter, data map[string]interface{}) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func SendToBroker(data conf.RequestWs, key string, valHeaderBroker string) error {

	// timeProcess := time.Now()

	procdr, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":   conf.GetBroker(),
		"delivery.timeout.ms": 1000,
	})
	if err != nil {
		log.Printf("Failed to create producer: %s\n", err)
		return err
		// os.Exit(1)
	}
	log.Printf("Created Producer %v\n", procdr)

	if procdr != nil {
		dataJson, err := json.Marshal(data)
		if err != nil {
			log.Printf("Error parse to JSON : %v\n", err)
			return err
		}

		//producer channel
		doneChan := make(chan bool)
		go func() {
			defer close(doneChan)
			for e := range procdr.Events() {
				switch ev := e.(type) {
				case *kafka.Message:
					m := ev
					if m.TopicPartition.Error != nil {
						log.Printf("Sent failed: %v\n", m.TopicPartition.Error)
					} else {
						log.Printf("Sent message to topic %s [%d] at offset %v\n",
							*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
					}
					return

				default:
					log.Printf("Ignored event: %s\n", ev)
				}
			}
		}()
		value := string(dataJson)
		//log.Printf("dataJson: %v\n", value)
		topicRequest := conf.GetTopicRequest()
		procdr.ProduceChannel() <- &kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topicRequest, Partition: kafka.PartitionAny},
			Value:          []byte(value),
			Headers:        []kafka.Header{{Key: key, Value: []byte(valHeaderBroker)}},
		}
		// wait for delivery report goroutine to finish
		_ = <-doneChan

	} else {
		//if database not defined, display packet in log
		log.Printf("Message Broker not defined\n")
	}

	// duration := time.Since(timeProcess)
	log.Printf("Sent To Broker\n")

	return nil
}
