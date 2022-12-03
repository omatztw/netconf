package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/nemith/netconf"
	ncssh "github.com/nemith/netconf/transport/ssh"
)

const sshAddr = "localhost:12345"

type DataNotif struct {
	netconf.NotificationMsg
	DataNotification struct {
		XMLName  xml.Name `xml:"yang:lighty:test:notifications dataNotification"`
		Ordinal  int      `xml:"Ordinal"`
		Payload  string   `xml:"Payload"`
		ClientId int      `xml:"ClientId"`
	}
}
type DataNotificationListener struct{}

func (d *DataNotificationListener) Do(msg netconf.NotificationMsg) {
	raw, err := xml.Marshal(msg)
	if err != nil {
		return
	}
	var data DataNotif
	if err := xml.Unmarshal(raw, &data); err != nil {
		return
	}
	fmt.Printf("%v\n", data.DataNotification.Payload)
}

type DataNotificationListener2 struct {
	count int
}

func (d *DataNotificationListener2) Do(msg netconf.NotificationMsg) {
	raw, err := xml.Marshal(msg)
	if err != nil {
		return
	}
	var data DataNotif
	if err := xml.Unmarshal(raw, &data); err != nil {
		return
	}
	d.count++
	fmt.Printf("%v\n", d.count)
}

func main() {
	config := &ssh.ClientConfig{
		User: "admin",
		Auth: []ssh.AuthMethod{
			ssh.Password("admin"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	transport, err := ncssh.Dial(ctx, "tcp", sshAddr, config)
	if err != nil {
		panic(err)
	}
	defer transport.Close()

	session, err := netconf.Open(transport)
	if err != nil {
		panic(err)
	}
	defer session.Close(context.Background())

	listener := &DataNotificationListener{}
	listener2 := &DataNotificationListener2{}
	session.AddNotificationListener("l1", listener)
	session.AddNotificationListener("l2", listener2)

	err = session.CreateSubscription(ctx, "aa:dataNotification")
	if err != nil {
		log.Fatalf("failed to subscribe: %v", err)
	}

	filter := netconf.Filter{
		Type:     netconf.Subtree,
		InnerXML: []byte("<netconf-state xmlns=\"urn:ietf:params:xml:ns:yang:ietf-netconf-monitoring\" />"),
	}
	data, err := session.Get(ctx, filter)
	if err != nil {
		log.Fatalf("failed to get: %v", err)
	}
	fmt.Printf("GET DATA:: %v\n", string(data))

	loop := make(chan struct{})
	<-loop
}
