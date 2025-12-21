package tolliver_test

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"testing"
)

// TODO: Check about mTLS
// TODO: Switch to buffer pool

func TestHandshake(t *testing.T) {
	cert1, err := tls.LoadX509KeyPair("./exampleCerts/instance1.crt", "./exampleCerts/instance1.key")
	if err != nil {
		panic(err)
	}
	cert2, err := tls.LoadX509KeyPair("./exampleCerts/instance2.crt", "./exampleCerts/instance2.key")
	if err != nil {
		panic(err)
	}
	caPEM, err := os.ReadFile("./exampleCerts/root.crt")
	if err != nil {
		panic(err)
	}
	caPool := x509.NewCertPool()
	if ok := caPool.AppendCertsFromPEM(caPEM); !ok {
		panic("Couldn't parse CA'")
	}

}
