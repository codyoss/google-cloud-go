package main

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

var (
	project    = os.Getenv("GOOGLE_CLOUD_PROJECT")
	secretName = os.Getenv("SECRET_NAME")
)

// GOOGLE_CLOUD_PROJECT=some-project SECRET_NAME=secret-id go run cloud.google.com/go/internal/cloguserexamples/loggingoff
func main() {
	if err := run(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()
	client, err := secretmanager.NewRESTClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	resp, err := client.GetSecret(ctx, &secretmanagerpb.GetSecretRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s", project, secretName),
	})
	if err != nil {
		return err
	}
	fmt.Printf("got secret: %s\n", resp.Name)
	return nil
}
