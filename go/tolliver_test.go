package tolliver

import (
	"testing"
	"time"
)

func TestBasicTest(t *testing.T) {
	NewInstance(InstanceOptions{
		nil,
		"tolliver.sqlite",
		time.Now(),
		5555,
	})
}
