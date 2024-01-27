package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
)

var googleIDToken = func(ctx context.Context) (string, error) {
	endpoint := "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/identity?audience=api://AzureADTokenExchange"
	resp, err := http.Get(endpoint)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bytes, err := io.ReadAll(resp.Body)
	return string(bytes), err
}

type A struct {
	ctx    context.Context
	client *msgraphsdk.GraphServiceClient
}

func main() {
	ctx := context.Background()
	cred, err := azidentity.NewClientAssertionCredential("47f9d27d-8362-4034-a0bd-4bf7582904e1", "8fd0131b-bd08-449e-b3a1-ffcdf9f34d0d", googleIDToken, nil)
	if err != nil {
		log.Fatal(err)
	}
	client, err := msgraphsdk.NewGraphServiceClientWithCredentials(cred, []string{".default"})
	if err != nil {
		log.Fatal(err)
	}
	a := &A{
		ctx:    ctx,
		client: client,
	}
	log.Print("starting server...")
	http.HandleFunc("/", a.handler)

	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}

	// Start HTTP server.
	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func (a *A) handler(w http.ResponseWriter, r *http.Request) {
	user, err := a.client.Users().ByUserId("547fa9ec-60a9-4461-8d60-4b2b331e478e").Get(a.ctx, nil)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	fmt.Fprintln(w, user.GetDisplayName())
}
