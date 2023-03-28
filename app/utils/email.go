package utils

import (
	"bytes"
	"fmt"
	"gopkg.in/gomail.v2"
	"html/template"
	"io"
	"os"
)

type EmailReceiver struct {
	Name    string
	Address string
}

func EmailSender(htmlTemplate string, params any, receiverList []EmailReceiver) {
	config, err := LoadConfig(".")
	if err != nil {
		fmt.Errorf("error while load config: %w", err)
		return
	}

	d := gomail.NewDialer(config.EmailSMTPHost, config.EmailSMTPPort, config.EmailAUTHUsername, config.EmailAUTHPassword)
	s, err := d.Dial()
	if err != nil {
		fmt.Errorf("error while setup email config: %w", err)
		return
	}

	templateFile, err := os.Open("public/" + htmlTemplate)
	if err != nil {
		fmt.Errorf("error while load template: %w", err)
		return
	}
	defer templateFile.Close()

	// Baca isi file dan konversi ke string
	templateBytes, err := io.ReadAll(templateFile)
	if err != nil {
		fmt.Errorf("error while convert template to string: %w", err)
		return
	}
	templateString := string(templateBytes)

	// Parse the template string
	t, err := template.New("webpage").Parse(templateString)
	if err != nil {
		fmt.Errorf("error while parse template string: %w", err)
		return
	}

	// Execute the template with the parameter values and capture the output as a string
	var buf bytes.Buffer
	err = t.Execute(&buf, params)
	if err != nil {
		fmt.Errorf("error execute template: %w", err)
		return
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
			return
		}
		m.Reset()
	}

	fmt.Println("Success sending email")
}
