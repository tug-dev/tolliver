package tolliver_test

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	tolliver "github.com/tug-dev/tolliver/go"
)

// TODO: Check about mTLS
// TODO: Switch to buffer pool
// TODO: Add mutexes because currently not thread safe

func TestHandshake(t *testing.T) {
	cert1, err := tls.LoadX509KeyPair("./testData/instance1.crt", "./testData/instance1.key")
	if err != nil {
		t.Error(err)
	}
	cert2, err := tls.LoadX509KeyPair("./testData/instance2.crt", "./testData/instance2.key")
	if err != nil {
		t.Error(err)
	}
	caPEM, err := os.ReadFile("./testData/root.crt")
	if err != nil {
		t.Error(err)
	}
	caPool := x509.NewCertPool()
	if ok := caPool.AppendCertsFromPEM(caPEM); !ok {
		t.Error("Could not parse root cert")
	}

	println("Loaded certs")

	inst1, err := tolliver.NewInstance(&tolliver.InstanceOptions{
		Port:         8000,
		CA:           caPool,
		InstanceCert: &cert1,
		DatabasePath: "inst1.db",
	})
	if err != nil {
		t.Error(err)
	}

	inst2, err := tolliver.NewInstance(&tolliver.InstanceOptions{
		Port:         9000,
		CA:           caPool,
		InstanceCert: &cert2,
		DatabasePath: "inst2.db",
	})
	if err != nil {
		t.Error(err)
	}

	println("Created instances")

	err = inst1.NewConnection(&net.TCPAddr{IP: []byte{127, 0, 0, 1}, Port: 9000}, "")
	if err != nil {
		t.Error(err)
	}

	println("Connected instances")

	inst2.Subscribe("test", "key")

	println("Sent subscription")

	inst2.Register("test", "key", func(m []byte) {
		fmt.Printf("Received message: %s\n", string(m))
	})
	time.Sleep(1 * time.Millisecond)

	inst1.Send("test", "key", []byte("Hello World!"))

	println("Sent message")
	time.Sleep(50 * time.Millisecond)
}
