package utils

import (
	"testing"
)

func TestDefaultInsecureClient(t *testing.T) {
	defaultClient := DefaultInsecureClient()
	if defaultClient == nil {
		t.Fatalf("the defaultClient cannot be nil")
	}
	if defaultClient.Transport == nil {
		t.Errorf("Transport cannot be nil")
	}
}
