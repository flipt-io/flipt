//go:build example
// +build example

package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"text/template"

	flipt "go.flipt.io/flipt-grpc"

	"google.golang.org/grpc"
)

type data struct {
	FlagKey     string
	FlagName    string
	FlagEnabled bool
}

var (
	fliptServer string
	flagKey     string
)

func init() {
	flag.StringVar(&fliptServer, "server", "flipt:9000", "address of Flipt backend server")
	flag.StringVar(&flagKey, "flag", "example", "flag key to query")
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	conn, err := grpc.Dial(fliptServer, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	log.Printf("connected to Flipt server at: %s", fliptServer)

	client := flipt.NewFliptClient(conn)

	t := template.Must(template.ParseFiles("./tmpl/basic.html"))

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flag, err := client.GetFlag(context.Background(), &flipt.GetFlagRequest{
			Key: flagKey,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if flag == nil {
			http.Error(w, "flag not found", http.StatusNotFound)
			return
		}

		log.Printf("got flag: %v", flag)

		data := data{
			FlagKey:     flag.Key,
			FlagName:    flag.Name,
			FlagEnabled: flag.Enabled,
		}

		if err := t.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	log.Println("Flipt UI available at http://localhost:8080")
	log.Println("Client UI available at http://localhost:8000")
	log.Printf("Flag Key: %q\n", flagKey)
	log.Fatal(http.ListenAndServe(":8000", h))
}
