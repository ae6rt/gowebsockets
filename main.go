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

	originURL, err := url.Parse("https://" + *ipAddress + "/api/v1/watch/namespaces/decap/pods?watch=true&labelSelector=type=decap-build")
	if err != nil {
		log.Fatal(err)
	}
	serviceURL, err := url.Parse("wss://" + *ipAddress + "/api/v1/watch/namespaces/decap/pods?watch=true&labelSelector=type=decap-build")
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

	// returns a Pod Object
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
		var pod Pod
		if err := json.Unmarshal([]byte(msg), &pod); err != nil {
			log.Println(err)
			continue
		}
		var deletePod bool
		for _, status := range pod.Object.Status.Statuses {
			if status.Name == "build-server" && status.State.Terminated.ContainerID != "" {
				deletePod = true
				break
			}
		}
		if deletePod {
			log.Printf("Would delete:  %+v\n", pod)
		}

	}
	os.Exit(0)
}

type Pod struct {
	Object Object `json:"object"`
}

type Object struct {
	Meta   Metadata `json:"metadata"`
	Status Status   `json:"status"`
}

type Metadata struct {
	Name string `json:"name"`
}

type Status struct {
	Statuses []ContainerStatus `json:"containerStatuses"`
}

type ContainerStatus struct {
	Name  string `json:"name"`
	Ready bool   `json:"ready"`
	State State  `json:"state"`
}

type State struct {
	Terminated Terminated `json:"terminated"`
}

type Terminated struct {
	ContainerID string `json:"containerID"`
	ExitCode    int    `json:"exitCode"`
}
