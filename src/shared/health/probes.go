package health

import (
	"context"
	"fmt"

	dapr "github.com/dapr/go-sdk/client"
)

// DaprProbe checks the Dapr sidecar is reachable via gRPC.
type DaprProbe struct {
	Client dapr.Client
}

func (d DaprProbe) Name() string { return "dapr" }

func (d DaprProbe) Check(ctx context.Context) error {
	if d.Client == nil {
		return fmt.Errorf("client not initialized")
	}
	_, err := d.Client.GetMetadata(ctx)
	return err
}
