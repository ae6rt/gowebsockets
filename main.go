package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"io"
	"os"

	"crypto/tls"
	"log"

	"net/url"

	"golang.org/x/net/websocket"
)

var (
	user      = flag.String("user", "", "Kubernetes master username")
	password  = flag.String("password", "", "Kubernetes master password")
	ipAddress = flag.String("ip-address", "", "Kubernetes master IP address")
)

func init() {
	flag.Parse()

}
func main() {
	if *user == "" {
		log.Fatalf("No user/pass\n")
	}

	originURL, err := url.Parse("https://" + *ipAddress + "/api/v1/watch/namespaces/default/events?watch=true&pretty=true")
	if err != nil {
		log.Fatal(err)
	}
	serviceURL, err := url.Parse("wss://" + *ipAddress + "/api/v1/watch/namespaces/default/events?watch=true&pretty=true")
	if err != nil {
		log.Fatal(err)
	}

	cfg := websocket.Config{
		Location:  serviceURL,
		Origin:    originURL,
		TlsConfig: &tls.Config{InsecureSkipVerify: true},
		Header:    map[string][]string{"Authorization": []string{"Basic " + base64.StdEncoding.EncodeToString([]byte(*user+":"+*password))}},
		Version:   websocket.ProtocolVersionHybi13,
	}

	conn, err := websocket.DialConfig(&cfg)
	if err != nil {
		log.Fatalf("Error opening connection: %v\n", err)
	}

	var msg string
	for {
		err := websocket.Message.Receive(conn, &msg)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Println("Couldn't receive msg " + err.Error())
			break
		}
		var event Event
		if err := json.Unmarshal([]byte(msg), &event); err != nil {
			log.Println(err)
			continue
		}
		if event.Object.Reason != "creating loadbalancer failed" {
			log.Printf("%+v\n", event)
		}
	}
	os.Exit(0)
}

type Event struct {
	Object Object `json:"object"`
	Type   string `json:"type"`
}

type Object struct {
	InvolvedObject InvolvedObject `json:"involvedObject"`
	Reason         string         `json:"reason"`
}

type InvolvedObject struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}
