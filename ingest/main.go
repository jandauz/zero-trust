package main

import (
	"context"
	"log"
	"os"

	"github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/service/common"
	"github.com/dapr/go-sdk/service/http"
	"github.com/ohler55/ojg/oj"
	"github.com/rs/xid"
)

func main() {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

	s := http.NewService(":8080")

	c, err := client.NewClient()
	if err != nil {
		logger.Fatal(err)
	}
	defer c.Close()

	err = s.AddServiceInvocationHandler("delivery-requests", handler(c, logger))
	if err != nil {
		logger.Fatal(err)
	}

	logger.Printf("listening on :8080")
	logger.Fatal(s.Start())
}

func handler(c client.Client, logger *log.Logger) func(context.Context, *common.InvocationEvent) (*common.Content, error) {
	return func(ctx context.Context, in *common.InvocationEvent) (*common.Content, error) {
		logger.Printf("Invocation (ContentType:%s, Verb:%s, QueryString:%s, Data:%s",
			in.ContentType, in.Verb, in.QueryString, string(in.Data))
		var req deliveryRequest
		err := oj.Unmarshal(in.Data, &req)
		if err != nil {
			return nil, err
		}

		id := xid.New().String()
		req.ID = id

		b, err := oj.Marshal(req)
		if err != nil {
			return nil, err
		}
		err = c.PublishEvent(ctx, "pubsub", "delivery-requests", b)
		if err != nil {
			return nil, err
		}

		b, err = oj.Marshal(deliveryResponse{
			DeliveryID: id,
		})
		if err != nil {
			return nil, err
		}
		return &common.Content{
			ContentType: in.ContentType,
			Data:        b,
		}, nil
	}
}

type deliveryRequest struct {
	ID      string `json:"id,omitempty"`
	OwnerID string `json:"owner_id,omitempty"`
	From    string `json:"from,omitempty"`
	To      string `json:"to,omitempty"`
}

type deliveryResponse struct {
	DeliveryID string `json:"delivery_id,omitempty"`
}
