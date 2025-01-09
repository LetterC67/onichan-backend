package services

import (
	"os"
	"strconv"

	gomail "gopkg.in/mail.v2"
)

var EMAIL string
var PORT int
var USERNAME string
var PASSWORD string
var HOST string

func LoadEnv() {
	EMAIL = os.Getenv("EMAIL")
	PORT, _ = strconv.Atoi(os.Getenv("EMAIL_PORT"))
	USERNAME = os.Getenv("EMAIL_USERNAME")
	PASSWORD = os.Getenv("EMAIL_PASSWORD")
	HOST = os.Getenv("EMAIL_HOST")
}

func SendEmail(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", EMAIL)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(HOST, PORT, USERNAME, PASSWORD)

	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
