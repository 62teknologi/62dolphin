package utils

import (
	"crypto/rand"
	"fmt"
	"github.com/62teknologi/62dolphin/62golib/utils"
	"time"
)

const otpChars = "1234567890"

func GenerateOTP(length int) (string, error) {
	buffer := make([]byte, length)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}

	otpCharsLength := len(otpChars)
	for i := 0; i < length; i++ {
		buffer[i] = otpChars[int(buffer[i])%otpCharsLength]
	}

	return string(buffer), nil
}

func VerifyOTP(result map[string]any, otpMethod string, otpReceiver string, otpCode string) error {
	utils.DB.Table("otps").Where("type = ?", otpMethod).Where("receiver = ?", otpReceiver).Where("code = ?", otpCode).Take(&result)

	if result["id"] == nil {
		return fmt.Errorf("invalid otp")
	} else if result["expires_at"].(time.Time).Unix() < time.Now().Unix() {
		return fmt.Errorf("expired otp")
	}

	utils.DB.Table("otps").Where("id = ?", result["id"]).Updates(map[string]any{
		"expires_at": time.Now(),
	})

	return nil
}
