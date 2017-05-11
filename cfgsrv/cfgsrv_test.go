package main_test

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/onsi/gomega/gbytes"
	"github.com/tonjun/cfgsrv"
	"github.com/tonjun/wsclient"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func getListenAddress() string {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	addr := l.Addr().String()
	l.Close()
	return addr
}

var _ = Describe("ConfigServer", func() {

	var (
		server *cfgsrv.ConfigServer

		client1 *wsclient.WSClient
		client2 *wsclient.WSClient

		buffer1 *gbytes.Buffer
		buffer2 *gbytes.Buffer
	)

	BeforeEach(func() {
		buffer1 = gbytes.NewBuffer()
		buffer2 = gbytes.NewBuffer()

		connected := make(chan bool)

		addr := getListenAddress()

		// run server
		server = cfgsrv.NewConfigServer(&cfgsrv.Options{
			ListenAddr: addr,
			ConfigFile: "./test_config.json",
		})
		go server.Start()

		// connect a client1 and write all incoming message to gbytes.Buffer
		time.Sleep(10 * time.Millisecond)
		client1 = wsclient.NewWSClient(fmt.Sprintf("ws://%s", addr))
		client1.OnMessage(func(data []byte) {
			log.Printf("client1 recv: %s", string(data))
			buffer1.Write(data)
		})
		client1.OnOpen(func() {
			connected <- true
		})
		client1.OnError(func(err error) {
			log.Printf("reconnecting..")
			time.Sleep(10 * time.Millisecond)
			client1.Connect()
		})
		client1.Connect()
		<-connected

		// connect a client2 and write all incoming message to gbytes.Buffer
		client2 = wsclient.NewWSClient(fmt.Sprintf("ws://%s", addr))
		client2.OnMessage(func(data []byte) {
			log.Printf("client2 recv: %s", string(data))
			buffer2.Write(data)
		})
		client2.OnOpen(func() {
			connected <- true
		})
		client2.OnError(func(err error) {
			log.Printf("reconnecting..")
			time.Sleep(10 * time.Millisecond)
			client2.Connect()
		})
		client2.Connect()
		<-connected

	})

	AfterEach(func() {
		server.Stop()
	})

	It("op \"get\" should return the config and the empty list of servers", func() {

		client1.SendJSON(wsclient.M{
			"op":   "get",
			"type": "request",
			"id":   "get1",
		})
		Eventually(buffer1).Should(gbytes.Say(
			`{"op":"get","type":"response","id":"get1","config":\{"feature1":\{"enable":false\},"feature2":\{"enable":true\}\}}`,
		))

	})

})
