package utils

import (
	"bytes"
	"fmt"
	"gopkg.in/gomail.v2"
	"html/template"
	"io"
	"log"
	"os"
)

type EmailReceiver struct {
	Name    string
	Address string
}

func EmailSender(htmlTemplate string, params any, receiverList []EmailReceiver) {
	config, err := LoadConfig(".")
	if err != nil {
		fmt.Errorf("cannot load config: %w", err)
		return
	}

	d := gomail.NewDialer(config.EmailSMTPHost, config.EmailSMTPPort, config.EmailAUTHUsername, config.EmailAUTHPassword)
	s, err := d.Dial()
	if err != nil {
		panic(err)
	}

	templateFile, err := os.Open("public/" + htmlTemplate)
	if err != nil {
		log.Fatal(err)
	}
	defer templateFile.Close()

	// Baca isi file dan konversi ke string
	templateBytes, err := io.ReadAll(templateFile)
	if err != nil {
		log.Fatal(err)
	}
	templateString := string(templateBytes)

	// Parse the template string
	t, err := template.New("webpage").Parse(templateString)
	if err != nil {
		panic(err)
	}

	// Execute the template with the parameter values and capture the output as a string
	var buf bytes.Buffer
	err = t.Execute(&buf, params)
	if err != nil {
		panic(err)
	}

	// Convert the buffer to a string
	html := buf.String()

	m := gomail.NewMessage()
	for _, r := range receiverList {
		fmt.Println(r)
		m.SetHeader("From", config.EmailSenderName)
		m.SetAddressHeader("To", r.Address, r.Name)
		m.SetHeader("Subject", "Newsletter #1")
		m.SetBody("text/html", html)
		//m.Attach("public/verify_user.html") // attach whatever you want

		if err := gomail.Send(s, m); err != nil {
			fmt.Printf("Could not send email to %q: %v", r.Address, err)
		}
		m.Reset()
	}

	fmt.Println("Success sending email")
}
