package tolliver_test

import (
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"fmt"
	"log"
	"log/slog"
	"math/rand/v2"
	"os"
	"os/exec"
	"strings"
	"testing"

	tolliver "github.com/tug-dev/tolliver/go"
)

// TODO: Check about mTLS
// TODO: Switch to buffer pool
// TODO: Add mutexes because currently not thread safe

var toClean = make([]string, 0)
var rSet = false
var certPool = x509.NewCertPool()
var globalCaName string

func registerCleanup(file string) {
	toClean = append(toClean, file)
}

func cleanup() {
	for _, f := range toClean {
		os.Remove(f)
	}
}

func makeCert(caName string) *tls.Certificate {
	n := rand.IntN(65536)

	makeCmd := exec.Command("./testData/mkcrt.sh", fmt.Sprintf("%d", n), caName)
	if _, err := makeCmd.Output(); err != nil {
		println(err.Error())
		log.Fatal("Could not make certificate")
	}
	registerCleanup(fmt.Sprintf("%d.crt", n))
	registerCleanup(fmt.Sprintf("%d.key", n))

	out, err := tls.LoadX509KeyPair(fmt.Sprintf("./%d.crt", n), fmt.Sprintf("./%d.key", n))

	if err != nil {
		log.Fatal("Could not read cert")
	}
	return &out
}

func makeCa() (name string, ca []byte) {
	n := rand.IntN(65536)
	makeCmd := exec.Command("./testData/mkca.sh", fmt.Sprintf("%d", n))
	if _, err := makeCmd.Output(); err != nil {
		println(err.Error())
		log.Fatal("Could not make CA")
	}

	registerCleanup(fmt.Sprintf("%d.crt", n))
	registerCleanup(fmt.Sprintf("%d.key", n))

	out, err := os.ReadFile(fmt.Sprintf("%d.crt", n))
	if err != nil {
		log.Fatal("Could not read CA")
	}

	return fmt.Sprintf("%d", n), out
}

func makeInstance(logger *slog.Logger, remotes []tolliver.RemoteAddr) *tolliver.Instance {
	n := rand.IntN(65536)
	config := tolliver.InstanceOptions{
		Remotes:      remotes,
		DatabasePath: fmt.Sprintf("%d.tolliver.db", n),
		CA:           certPool,
		InstanceCert: makeCert(globalCaName),
		Port:         min(uint16(n+1000), 1001),
		Logger:       logger,
	}

	out, err := tolliver.NewInstance(&config)
	if err != nil {
		log.Fatal("Could not make instance")
	}

	registerCleanup(fmt.Sprintf("./%d.tolliver.db", n))

	return out
}

func makeLogger(b *strings.Builder) *slog.Logger {
	out := slog.New(slog.NewTextHandler(b, nil))
	return out
}

func globalSetup() {
	name, ca := makeCa()
	certPool.AppendCertsFromPEM(ca)
	globalCaName = name
}

func TestHandshakeSuccess(t *testing.T) {
	globalSetup()

	var b1 strings.Builder
	i1 := makeInstance(makeLogger(&b1), nil)
	i1.Register("", "", func(mes []byte) bool {
		fmt.Printf("% x", mes)
		return true
	})

	print(b1.String())

	cleanup()
}

// Handshake and all failure paths
// Multiple connections subscribing to different key channel pairs and receiving the correct messages
// Ping for keepalive
// Retry connections until connected
// Retry messages until received
// Send messages while recipient not connected
