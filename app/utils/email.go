package utils

import (
	"fmt"
	"gopkg.in/gomail.v2"
)

type receiver struct {
	Name    string
	Address string
}

const (
	CONFIG_SMTP_HOST     = "smtp.gmail.com"
	CONFIG_SMTP_PORT     = 587
	CONFIG_SENDER_NAME   = "PT. Makmur Subur Jaya <emailanda@gmail.com>"
	CONFIG_AUTH_EMAIL    = "dimas.bagus@62teknologi.com"
	CONFIG_AUTH_PASSWORD = "kdvdwoabjontwclu"
)

func EmailSender() {
	receiverList := []receiver{
		{
			Name:    "Dimas",
			Address: "dimasbagussusilo@gmail.com",
		},
	}
	d := gomail.NewDialer(CONFIG_SMTP_HOST, CONFIG_SMTP_PORT, CONFIG_AUTH_EMAIL, CONFIG_AUTH_PASSWORD)
	s, err := d.Dial()
	if err != nil {
		panic(err)
	}

	m := gomail.NewMessage()
	for _, r := range receiverList {
		fmt.Println(r)
		m.SetHeader("From", "no-reply@example.com")
		m.SetAddressHeader("To", r.Address, r.Name)
		m.SetHeader("Subject", "Newsletter #1")
		m.SetBody("text/html", fmt.Sprintf("Hello %s!", r.Name))

		if err := gomail.Send(s, m); err != nil {
			fmt.Printf("Could not send email to %q: %v", r.Address, err)
		}
		m.Reset()
	}

	fmt.Println("Success sending email")
}
