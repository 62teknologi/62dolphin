package adapters

import (
	"fmt"
	"github.com/62teknologi/62dolphin/app/interfaces"
)

func GetAdapter(name string) (interfaces.AuthInterface, error) {
	var adapter interfaces.AuthInterface

	switch name {
	case "google":
		adapter = &GoogleAdapter{}
	case "facebook":
		adapter = &FacebookAdapter{}
	case "microsoft":
		adapter = &MicrosoftAdapter{}
	case "local":
		adapter = &LocalAdapter{}
	case "apple":
		adapter = &AppleAdapter{}
	default:
		return nil, fmt.Errorf("adapter %s not found", name)
	}

	return adapter, nil
}

type Profile struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Photo     string `json:"photo"`
	Gender    string `json:"gender"`
	Birthdate string `json:"birtdate"`
	AgeMin    int    `json:"age_min"`
	AgeMax    int    `json:"age_max"`
	AgeRange  string `json:"age_range"`
}
