package adapters

import (
	"github.com/62teknologi/62dolphin/app/interfaces"
)

func GetAdapter(name string) interfaces.AuthInterface {
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
	}

	return adapter
}

type Profile struct {
	Gid       string `json:"google_id"`
	Fbid      string `json:"facebook_id"`
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
