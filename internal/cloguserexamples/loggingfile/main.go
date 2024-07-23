package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"cloud.google.com/go/clog"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

var (
	project    = os.Getenv("GOOGLE_CLOUD_PROJECT")
	secretName = os.Getenv("SECRET_NAME")
)

// GOOGLE_CLOUD_PROJECT=some-project SECRET_NAME=secret-id go run cloud.google.com/go/internal/cloguserexamples/loggingfile
func main() {
	oldEnv := os.Getenv("GOOGLE_SDK_DEBUG_LOGGING")
	defer os.Setenv("GOOGLE_SDK_DEBUG_LOGGING", oldEnv)
	os.Setenv("GOOGLE_SDK_DEBUG_LOGGING", "true")

	if err := run(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	// Create a file to log to
	f, err := os.Create("temp.txt")
	if err != nil {
		return err
	}
	defer f.Close()
	// Example user configuration
	clog.SetDefaults(&clog.DefaultOptions{
		Writer: f,
		Level:  slog.LevelDebug,
	})

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
