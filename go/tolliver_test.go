package tolliver_test

import (
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"fmt"
	"log"
	"log/slog"
	"math/rand/v2"
	"net"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

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
		Port:         max(uint16(n), 1025),
		Logger:       logger,
		Interface:    "127.0.0.1",
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

func matches(pattern, literal string) bool {
	if strings.Contains(pattern, "*") {
		a := strings.Split(pattern, "*")
		if len(a) > 2 {
			return false
		}

		if len(literal) >= len(a[0]) && a[0] == literal[0:len(a[0])] {
			return true
		} else {
			return false
		}
	}

	return pattern == literal
}

func expectLogs(log string, lines ...string) bool {
	pointer := 0
	for _, l := range strings.Split(log, "\n") {
		a := strings.Split(l, "msg=\"")
		if len(a) == 1 {
			continue
		}

		msg := a[1][0 : len(a[1])-1]
		if matches(lines[pointer], msg) {
			pointer++
			if pointer == len(lines) {
				return true
			}
		}
	}

	return false
}

func globalSetup() {
	name, ca := makeCa()
	certPool.AppendCertsFromPEM(ca)
	globalCaName = name
}

func TestMain(m *testing.M) {
	certPool = x509.NewCertPool()
	globalSetup()
	m.Run()
	cleanup()
	toClean = make([]string, 0)
}

func TestHandshakeSuccess(t *testing.T) {
	var b1 strings.Builder
	var b2 strings.Builder
	i1 := makeInstance(makeLogger(&b1), nil)

	i2 := makeInstance(makeLogger(&b2), nil)
	i2.NewConnection(tolliver.RemoteAddr{
		&net.TCPAddr{
			IP:   []byte{127, 0, 0, 1},
			Port: i1.Port,
			Zone: "",
		},
		"tug.dev",
	})

	time.Sleep(2 * time.Second)

	f1 := expectLogs(b1.String(), "Started listening on *", "Connection established *")
	f2 := expectLogs(b2.String(), "Started listening on *", "Connection established *")
	if !f1 || !f2 {
		println(b1.String())
		println(b2.String())
		t.Fail()
	}
}

// Handshake and all failure paths
// Multiple connections subscribing to different key channel pairs and receiving the correct messages
// Ping for keepalive
// Retry connections until connected
// Retry messages until received
// Send messages while recipient not connected
