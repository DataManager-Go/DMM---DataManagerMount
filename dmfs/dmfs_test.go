package dmfs

import (
	"testing"

	"github.com/DataManager-Go/libdatamanager"
)

const (
	defaultNsFull   = "jojii_default"
	defaultNsTimmed = "default"
)

func TestFullNS(t *testing.T) {
	// Build pseudo config
	data := dataStruct{
		libdm: &libdatamanager.LibDM{
			Config: &libdatamanager.RequestConfig{
				Username: "jojii",
			},
		},
	}

	newNsName := data.fullNS(defaultNsFull)

	if newNsName != defaultNsFull {
		t.Errorf("Expected %s but got %s", defaultNsFull, newNsName)
	}
}

func TestTrimmedNS(t *testing.T) {
	// Build pseudo config
	data := dataStruct{
		libdm: &libdatamanager.LibDM{
			Config: &libdatamanager.RequestConfig{
				Username: "jojii",
			},
		},
	}

	newNsName := data.trimmedNS(defaultNsFull)

	if newNsName != defaultNsTimmed {
		t.Errorf("Expected %s but got %s", defaultNsTimmed, newNsName)
	}
}
