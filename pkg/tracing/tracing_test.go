package tracing

import "testing"

func TestInitialize(t *testing.T) {
	// Just testing to make sure this doesn't panic
	Initialize(Config{})
}
