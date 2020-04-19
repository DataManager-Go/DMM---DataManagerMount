package dmfs

import (
	"testing"

	"github.com/DataManager-Go/libdatamanager"
)

const (
	defaultNsFull   = "jojii_default"
	defaultNsTimmed = "default"
)

var (
	cfg = libdatamanager.RequestConfig{Username: "jojii"}
)

func TestAddNSName(t *testing.T) {
	newNsName := addNsName(defaultNsTimmed, &cfg)

	if newNsName != defaultNsFull {
		t.Errorf("Expected %s but got %s", defaultNsFull, newNsName)
	}
}

func TestRemoveNSName(t *testing.T) {
	newNsName := removeNsName(defaultNsFull)

	if newNsName != defaultNsTimmed {
		t.Errorf("Expected %s but got %s", defaultNsTimmed, newNsName)
	}
}
