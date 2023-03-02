package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Axway/agent-sdk/pkg/amplify/agent/correlation"
	"github.com/google/uuid"
	"github.com/vivekschauhan/correlation-service/pkg/config"
	"github.com/vivekschauhan/correlation-service/pkg/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/structpb"
)

type Client interface {
	GetResource(path string) (service.Resource, error)
}

type client struct {
	Client
	svcClient correlation.CorrelationServiceClient
}

func NewClient(ctx context.Context, cfg *config.Config) (Client, error) {
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	conn, err := grpc.DialContext(ctx, fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		grpc.WithReturnConnectionError(),
		grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	svcClient := correlation.NewCorrelationServiceClient(conn)

	return &client{
		svcClient: svcClient,
	}, nil
}

type Metadata struct {
	FilterMetadata map[string]interface{}
}

func (c *client) GetResource(path string) (service.Resource, error) {
	res := service.Resource{}
	reqCtx := &correlation.TransactionContext{
		TransactionId: uuid.New().String(),
		Request: &correlation.Request{
			Path: path,
		},
	}

	reqCtx.Metadata = make(map[string]*structpb.Value)
	c.appendMetadata(reqCtx.Metadata, "custom_tags", map[string]string{
		"test": "value",
	})
	c.appendMetadata(reqCtx.Metadata, "filter_metadata", Metadata{
		FilterMetadata: map[string]interface{}{
			"test": "value",
		},
	})
	c.appendMetadata(reqCtx.Metadata, "sample_rate", "100")

	resCtx, err := c.svcClient.GetResourceContext(context.Background(), reqCtx)
	if err != nil {
		return res, err
	}
	res.APIID = resCtx.ApiId
	res.Version = resCtx.Version
	res.Stage = resCtx.Stage
	res.ClientID = resCtx.ConsumerId

	return res, nil
}

func (c *client) appendMetadata(metadataMap map[string]*structpb.Value, key string, value interface{}) error {
	v, err := structpb.NewValue(value)
	if err != nil {
		buf, _ := json.Marshal(value)
		m := make(map[string]interface{})
		json.Unmarshal(buf, &m)
		v, err = structpb.NewValue(m)
	}
	if err != nil {
		return err
	}
	metadataMap[key] = v
	return nil
}
