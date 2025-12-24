package tolliver_test

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"os"
	"testing"

	tolliver "github.com/tug-dev/tolliver/go"
)

// TODO: Check about mTLS
// TODO: Switch to buffer pool

func TestHandshake(t *testing.T) {
	t.Log("Started test")
	cert1, err := tls.LoadX509KeyPair("./exampleCerts/instance1.crt", "./exampleCerts/instance1.key")
	if err != nil {
		t.Error(err)
	}
	cert2, err := tls.LoadX509KeyPair("./exampleCerts/instance2.crt", "./exampleCerts/instance2.key")
	if err != nil {
		t.Error(err)
	}
	caPEM, err := os.ReadFile("./exampleCerts/root.crt")
	if err != nil {
		t.Error(err)
	}
	caPool := x509.NewCertPool()
	if ok := caPool.AppendCertsFromPEM(caPEM); !ok {
		t.Error("Could not parse root cert")
	}

	inst1, err := tolliver.NewInstance(&tolliver.InstanceOptions{
		Port:         8000,
		CA:           caPool,
		InstanceCert: &cert1,
	})

	if err != nil {
		t.Error(err)
	}

	inst2, err := tolliver.NewInstance(&tolliver.InstanceOptions{
		Port:         9000,
		CA:           caPool,
		InstanceCert: &cert2,
	})

	if err != nil {
		t.Error(err)
	}

	err = inst1.NewConnection(net.TCPAddr{IP: []byte{127, 0, 0, 1}, Port: 9000}, "")
	if err != nil {
		t.Error(err)
	}
	inst2.Subscribe("test", "key")
	for {

	}
}
