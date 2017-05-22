package main_test

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/onsi/gomega/gbytes"
	"github.com/tonjun/cfgsrv"
	"github.com/tonjun/pubsub"
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

func connectClient(addr string, buff *gbytes.Buffer, name string) *wsclient.WSClient {
	connected := make(chan bool)
	c := wsclient.NewWSClient(fmt.Sprintf("ws://%s", addr))
	c.OnMessage(func(data []byte) {
		log.Printf("%s recv: %s", name, string(data))
		buff.Write(data)
	})
	c.OnOpen(func() {
		connected <- true
	})
	c.OnError(func(err error) {
		log.Printf("%s reconnecting..", name)
		time.Sleep(10 * time.Millisecond)
		c.Connect()
	})
	c.Connect()
	<-connected
	return c
}

var _ = Describe("ConfigServer", func() {

	var (
		server *cfgsrv.ConfigServer

		client1 *wsclient.WSClient
		client2 *wsclient.WSClient
		client3 *wsclient.WSClient

		buffer1 *gbytes.Buffer
		buffer2 *gbytes.Buffer
		buffer3 *gbytes.Buffer
	)

	BeforeEach(func() {
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
		buffer1 = gbytes.NewBuffer()
		buffer2 = gbytes.NewBuffer()
		buffer3 = gbytes.NewBuffer()

		addr := getListenAddress()

		// run server
		server = cfgsrv.NewConfigServer(&cfgsrv.Options{
			ListenAddr: addr,
			ConfigFile: "./test_config.json",
		})
		go server.Start()

		time.Sleep(10 * time.Millisecond)

		// create client connections and send incomding data to buffer
		client1 = connectClient(addr, buffer1, "client1")
		client2 = connectClient(addr, buffer2, "client2")
		client3 = connectClient(addr, buffer3, "client3")

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

	It("op \"connect\" should return the config and the list of peers", func() {

		client1.SendJSON(wsclient.M{
			"op":   "connect",
			"type": "request",
			"id":   "2",
			"addr": "127.0.0.1:7171",
		})
		Eventually(buffer1).Should(gbytes.Say(
			`{"op":"connect","type":"response","id":"2","peers":\["127\.0\.0\.1:7171"\],"config":\{"feature1":\{"enable":false\},"feature2":\{"enable":true\}\}}`,
		))

	})

	It("op \"connect\" should add the given addr to the list of peers", func() {

		client1.SendJSON(wsclient.M{
			"op":   "connect",
			"type": "request",
			"id":   "2",
			"addr": "127.0.0.1:7171",
		})
		Eventually(buffer1).Should(gbytes.Say(
			`{"op":"connect","type":"response","id":"2","peers":\["127\.0\.0\.1:7171"\],"config":\{"feature1":\{"enable":false\},"feature2":\{"enable":true\}\}}`,
		))

		client2.SendJSON(wsclient.M{
			"op":   "connect",
			"type": "request",
			"id":   "2",
			"addr": "192.168.0.100:7171",
		})
		Eventually(buffer2).Should(gbytes.Say(
			`{"op":"connect","type":"response","id":"2","peers":\["127\.0\.0\.1:7171"\,"192\.168\.0\.100:7171"],"config":\{"feature1":\{"enable":false\},"feature2":\{"enable":true\}\}}`,
		))

	})

	It("should inform all the peers on connect", func() {

		client1.SendJSON(wsclient.M{
			"op":   "connect",
			"type": "request",
			"id":   "2",
			"addr": "127.0.0.1:7171",
		})
		Eventually(buffer1).Should(gbytes.Say(
			`{"op":"connect","type":"response","id":"2","peers":\["127\.0\.0\.1:7171"\],"config":\{"feature1":\{"enable":false\},"feature2":\{"enable":true\}\}}`,
		))

		client2.SendJSON(wsclient.M{
			"op":   "connect",
			"type": "request",
			"id":   "2",
			"addr": "192.168.0.100:7171",
		})
		Eventually(buffer2).Should(gbytes.Say(
			`{"op":"connect","type":"response","id":"2","peers":\["127\.0\.0\.1:7171"\,"192\.168\.0\.100:7171"],"config":\{"feature1":\{"enable":false\},"feature2":\{"enable":true\}\}}`,
		))

		Eventually(buffer1).Should(gbytes.Say(
			`{"op":"peers_changed","type":"push","id":".","peers":\["127\.0\.0\.1:7171"\,"192\.168\.0\.100:7171"]}`,
		))
		Consistently(buffer2).ShouldNot(gbytes.Say(
			`{"op":"peers_changed","type":"push","id":".","peers":\["127\.0\.0\.1:7171"\,"192\.168\.0\.100:7171"]}`,
		))

		client3.SendJSON(wsclient.M{
			"op":   "connect",
			"type": "request",
			"id":   "req-client-3",
			"addr": "192.168.0.101:7171",
		})
		Eventually(buffer3).Should(gbytes.Say(
			`{"op":"connect","type":"response","id":"req-client-3","peers":\["127\.0\.0\.1:7171","192\.168\.0\.100:7171","192\.168\.0\.101:7171"],"config":\{"feature1":\{"enable":false\},"feature2":\{"enable":true\}\}}`,
		))

		Eventually(buffer1).Should(gbytes.Say(
			`{"op":"peers_changed","type":"push","id":"3","peers":\["127\.0\.0\.1:7171","192\.168\.0\.100:7171","192\.168\.0\.101:7171"\]}`,
		))
		Eventually(buffer2).Should(gbytes.Say(
			`{"op":"peers_changed","type":"push","id":"3","peers":\["127\.0\.0\.1:7171","192\.168\.0\.100:7171","192\.168\.0\.101:7171"\]}`,
		))

	})

	It("should remove disconnected clients from the memory store", func(done Done) {
		client1.SendJSON(wsclient.M{
			"op":   "connect",
			"type": "request",
			"id":   "2",
			"addr": "127.0.0.1:7171",
		})
		Eventually(buffer1).Should(gbytes.Say(
			`{"op":"connect","type":"response","id":"2","peers":\["127\.0\.0\.1:7171"\],"config":\{"feature1":\{"enable":false\},"feature2":\{"enable":true\}\}}`,
		))

		store := server.GetStore()
		i, found, _ := store.Get("127.0.0.1:7171")
		Expect(found).To(Equal(true))
		Expect(i).ShouldNot(BeNil())
		c := i.Value.(pubsub.Conn)
		connID := c.ID()
		log.Printf("client1 connection id: %d", connID)

		i, found, _ = store.Get(fmt.Sprintf("%d", connID))
		Expect(found).To(Equal(true))
		Expect(i).ShouldNot(BeNil())

		client2.SendJSON(wsclient.M{
			"op":   "connect",
			"type": "request",
			"id":   "2",
			"addr": "192.168.0.100:7171",
		})
		Eventually(buffer2).Should(gbytes.Say(
			`{"op":"connect","type":"response","id":"2","peers":\["127\.0\.0\.1:7171"\,"192\.168\.0\.100:7171"],"config":\{"feature1":\{"enable":false\},"feature2":\{"enable":true\}\}}`,
		))

		client1.Close()
		client1.OnClose(func() {
			defer GinkgoRecover()

			log.Printf("OnClose")

			time.Sleep(10 * time.Millisecond)

			store := server.GetStore()
			i, found, _ := store.Get("127.0.0.1:7171")
			Expect(found).To(Equal(false))
			Expect(i).Should(BeNil())

			i, found, _ = store.Get(fmt.Sprintf("%d", connID))
			Expect(found).To(Equal(false))
			Expect(i).Should(BeNil())

			close(done)
		})

	})

})
