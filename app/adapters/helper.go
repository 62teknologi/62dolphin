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
	}

	return adapter
}
