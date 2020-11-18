package otp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/mrjones/oauth"
	"github.com/parnurzeal/gorequest"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

const (
	APIGateway       = "https://gatewayapi.com/rest"
	GatewayAPIToken  = "qYixes_ETXuffKRGag_o5MnbtAIJF7WKw_osIyoJwpxzHrjmaYklBc19M8RwFIy3"
	GatewayAPIKey    = "OCCpiLF6z8RNe598pEZqD-jk"
	GatewayAPISecret = "BP!mpAXzCsz7qbGGIudvvCDmqI.j95-hPsxzGl3i"
)

func printBody(resp gorequest.Response, body string, errs []error) {
	// fmt.Println(resp.Status)
	fmt.Println(resp)
}

func SendSMS(phoneNumber string, message string) error {
	// Authentication
	consumer := oauth.NewConsumer(GatewayAPIKey, GatewayAPISecret, oauth.ServiceProvider{})
	client, err := consumer.MakeHttpClient(&oauth.AccessToken{})
	if err != nil {
		return err
	}

	// Request
	request := &GatewayAPIRequest{
		Sender:  "OnlineCity",
		Message: message,
		Recipients: []GatewayAPIRecipient{
			{
				Msisdn: cast.ToUint64(phoneNumber),
			},
		},
	}

	// Send it
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	res, err := client.Post(
		fmt.Sprintf(`%s/mtsms`, APIGateway),
		"application/json",
		&buf,
	)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		body, _ := ioutil.ReadAll(res.Body)
		return errors.New(fmt.Sprintf("http error reply, status: %q, body: %q", res.Status, body))
	}
	// Parse the response
	response := &GatewayAPIResponse{}
	if err := json.NewDecoder(res.Body).Decode(response); err != nil {
		return err
	}
	log.Println("ids", response.Ids)
	return nil
}

// https://gowalker.org/github.com/parnurzeal/gorequest
// https://gatewayapi.com/docs/rest.html#sending-sms-es
// https://mholt.github.io/json-to-go/
func SendSMS2(phoneNumber string, message string) error {
	request := gorequest.New()

	var res GatewayAPIRes
	//printBody

	request.Get(fmt.Sprintf(`%s/mtsms?token=%s&message=%s&recipients.0.msisdn=%s`, APIGateway, GatewayAPIToken, message, phoneNumber)).
		Retry(3, 2*time.Second, http.StatusBadRequest, http.StatusInternalServerError).
		Set("Accept", "application/json").
		Set("Content-Type", "application/json").
		// Send(fmt.Sprintf(`{"recipients.0.msisdn":"%s", "message":"%s"}`, , )).
		EndStruct(&res) //&res

	fmt.Println(res)
	// if res. != true {
	// 	return errors.New("SMS_SERVICE_UNAVAILABLE")
	// }

	return nil
}

func SendSMS_amazon(phoneNumber string, message string) error {
	// Create Session and assign AccessKeyID and SecretAccessKey
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(viper.GetString("aws_region_sns")),
		Credentials: credentials.NewStaticCredentials(viper.GetString("aws_key_sns"), viper.GetString("aws_secret_sns"), ""),
	},
	)

	// Create SNS service
	svc := sns.New(sess)

	// Pass the phone number and message.
	params := &sns.PublishInput{
		PhoneNumber: aws.String(phoneNumber),
		Message:     aws.String(message),
	}

	// sends a text message (SMS message) directly to a phone number.
	resp, err := svc.Publish(params)

	if err != nil {
		return errors.New("SMS_SERVICE_UNAVAILABLE")
	}

	fmt.Println(resp) // print the response data.

	return nil
}
