package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"

	"github.com/62teknologi/62dolphin/app/config"

	"gopkg.in/gomail.v2"
)

type EmailReceiver struct {
	Name    string
	Address string
}

func EmailSender(htmlTemplate string, params any, receiverList []EmailReceiver) {
	configs := config.Data

	d := gomail.NewDialer(configs.EmailSMTPHost, configs.EmailSMTPPort, configs.EmailAUTHUsername, configs.EmailAUTHPassword)
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
		m.SetHeader("From", configs.EmailSenderName)
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
