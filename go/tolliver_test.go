package tolliver_test

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"fmt"
	"log/slog"
	"net"
	"os"
	"slices"
	"testing"

	tolliver "github.com/tug-dev/tolliver/go"
)

// TODO: Check about mTLS
// TODO: Switch to buffer pool
// TODO: Add mutexes because currently not thread safe

var i1 *tolliver.Instance
var i2 *tolliver.Instance

var log1 *slog.Logger
var log2 *slog.Logger

var c1 tls.Certificate
var c2 tls.Certificate

// var i3 *tolliver.Instance
// var i4 *tolliver.Instance

//go:embed testData/root.crt
var ca string

var rSet = false

var certPool = x509.NewCertPool()

func globalSetup() {
	certPool.AppendCertsFromPEM([]byte(ca))
	c1, _ = tls.LoadX509KeyPair("./testData/i1.crt", "./testData/i1.key")
	c2, _ = tls.LoadX509KeyPair("./testData/i2.crt", "./testData/i2.key")

	f, _ := os.Create("i1.log")
	log1 = slog.New(slog.NewTextHandler(bufio.NewWriter(f), nil))

	f2, _ := os.Create("i2.log")
	log2 = slog.New(slog.NewTextHandler(bufio.NewWriter(f2), nil))
}

func setup() {
	if !rSet {
		globalSetup()
		rSet = true
	}

	cfg1 := tolliver.InstanceOptions{
		Remotes:      nil,
		DatabasePath: "i1.sqlite",
		CA:           certPool,
		InstanceCert: &c1,
		Interface:    "127.0.0.1",
		Port:         2000,
		Logger:       log1,
	}

	cfg2 := tolliver.InstanceOptions{
		Remotes:      nil,
		DatabasePath: "i2.sqlite",
		CA:           certPool,
		InstanceCert: &c2,
		Interface:    "127.0.0.1",
		Port:         3000,
		Logger:       log2,
	}
	var err error

	i1, err = tolliver.NewInstance(&cfg1)

	if err != nil {
		println(err)
	}

	i2, err = tolliver.NewInstance(&cfg2)
}

func TestHandshake(t *testing.T) {
	setup()

	i1.NewConnection(tolliver.RemoteAddr{
		&net.TCPAddr{
			IP:   []byte{127, 0, 0, 1},
			Port: 3000,
			Zone: "",
		},
		"tug.dev",
	})

	i2.Register("test", "key", func(x []byte) bool {
		if slices.Compare(x, []byte{8, 8, 8, 8}) != 0 {
			fmt.Printf("% x != % x\n", x, []byte{8, 8, 8, 8})
			t.Error("Incorrect message")
		}

		return true
	})
	i2.Subscribe("test", "key")
	i1.Send("test", "key", []byte{8, 8, 8, 8})
}
