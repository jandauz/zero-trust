package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/http"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
	"github.com/rs/xid"
)

func main() {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

	c, err := dapr.NewClient()
	if err != nil {
		logger.Fatal(err)
	}
	defer c.Close()

	secrets, err := c.GetSecret(context.Background(), "kubernetes", "azure-maps", nil)
	if err != nil {
		logger.Fatal(err)
	}

	key, ok := secrets["subscription-key"]
	if !ok {
		logger.Fatal("subscription-key missing")
	}

	h := handler{
		logger: logger,
		key:    key,
		u: url.URL{
			Scheme: "https",
			Host:   "atlas.microsoft.com",
		},
		p: url.Values{
			"api-version":      []string{"1.0"},
			"language":         []string{"en-US"},
			"query":            []string{},
			"subscription-key": []string{key},
		},
		parser: oj.Parser{},
		c:      c,
	}

	s := daprd.NewService(":8080")
	sub := &common.Subscription{
		PubsubName: "pubsub",
		Topic:      "delivery-requests",
		Route:      "/delivery-requests",
	}
	err = s.AddTopicEventHandler(sub, h.handle())
	if err != nil {
		logger.Fatal(err)
	}

	logger.Printf("listening on :8080")
	logger.Fatal(s.Start())
}

type handler struct {
	logger *log.Logger
	key    string
	u      url.URL
	p      url.Values
	parser oj.Parser
	c      dapr.Client
}

type topicEventHandlerFunc func(context.Context, *common.TopicEvent) (bool, error)

func (h handler) handle() topicEventHandlerFunc {
	return func(ctx context.Context, e *common.TopicEvent) (bool, error) {
		h.logger.Printf(
			"event - PubsubName:%s, Topic:%s, ID:%s, Data: %v",
			e.PubsubName, e.Topic, e.ID, e.Data,
		)

		var req deliveryRequest
		err := oj.Unmarshal([]byte(e.Data.(string)), &req)
		if err != nil {
			return true, nil
		}

		from, err := h.findGeolocation(req.From)
		if err != nil {
			return true, nil
		}

		to, err := h.findGeolocation(req.To)
		if err != nil {
			return true, nil
		}

		r, err := h.findRoute(from, to)
		if err != nil {
			return true, nil
		}

		id := xid.New().String()
		d := delivery{
			ID:   id,
			From: from,
			To:   to,
			Route: route{
				Distance:            r.LengthInMeters,
				EstimatedTravelTime: r.TravelTimeInSeconds,
			},
		}

		_, err = h.scheduleDelivery(ctx, d)
		if err != nil {
			return true, nil
		}

		return false, nil
	}
}

type deliveryRequest struct {
	ID      string `json:"id,omitempty"`
	OwnerID string `json:"owner_id,omitempty"`
	From    string `json:"from,omitempty"`
	To      string `json:"to,omitempty"`
}

func (h *handler) findGeolocation(address string) (geolocation, error) {
	h.logger.Printf("finding geolocation for address: %s", address)
	h.p.Set("query", address)
	h.u.Path = "search/address/json"
	h.u.RawQuery = h.p.Encode()
	resp, err := http.Get(h.u.String())
	if err != nil {
		return geolocation{}, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return geolocation{}, err
	}

	data, err := h.parser.Parse(b)
	if err != nil {
		return geolocation{}, err
	}

	x := jp.MustParseString("$.results[?(@.type == 'Point Address' || @.type == 'Address Range')].position")
	result := x.Get(data)[0]

	var g geolocation
	err = h.parser.Unmarshal([]byte(oj.JSON(result)), &g)
	if err != nil {
		return geolocation{}, nil
	}

	h.logger.Printf("geolocation: %v", g.String())
	return g, nil
}

type geolocation struct {
	Lat float64 `json:"lat,omitempty"`
	Lon float64 `json:"lon,omitempty"`
}

func (g geolocation) String() string {
	return fmt.Sprintf("%v,%v", g.Lat, g.Lon)
}

func (h *handler) findRoute(from, to geolocation) (azureRoute, error) {
	h.logger.Printf("finding route between %v and %v", from, to)
	h.p.Set("query", fmt.Sprintf("%s:%s", from.String(), to.String()))
	h.u.Path = "route/directions/json"
	h.u.RawQuery = h.p.Encode()
	resp, err := http.Get(h.u.String())
	if err != nil {
		return azureRoute{}, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return azureRoute{}, err
	}

	data, err := h.parser.Parse(b)
	if err != nil {
		return azureRoute{}, err
	}

	x := jp.MustParseString("$.routes[0].summary")
	result := x.Get(data)[0]

	var r azureRoute
	err = h.parser.Unmarshal([]byte(oj.JSON(result)), &r)
	if err != nil {
		return azureRoute{}, err
	}

	h.logger.Printf("distance: %v", r.LengthInMeters)
	h.logger.Printf("estimated travel time: %v", r.TravelTimeInSeconds)
	return r, nil
}

type azureRoute struct {
	LengthInMeters      int           `json:"lengthInMeters,omitempty"`
	TravelTimeInSeconds time.Duration `json:"travelTimeInSeconds,omitempty"`
}

type route struct {
	Distance            int           `json:"distance,omitempty"`
	EstimatedTravelTime time.Duration `json:"estimated_travel_time,omitempty"`
}

func (h *handler) scheduleDelivery(ctx context.Context, d delivery) (string, error) {
	h.logger.Printf("scheduling delivery: %#v", d)
	b, err := oj.Marshal(d)
	if err != nil {
		return "", err
	}
	b, err = h.c.InvokeMethodWithContent(ctx, "courier", "schedule-delivery", http.MethodPost, &dapr.DataContent{
		ContentType: "application/json",
		Data:        b,
	})
	if err != nil {
		return "", err
	}
	h.logger.Printf("scheduled delivery with courier: %v", string(b))
	return string(b), err
}

type delivery struct {
	ID    string      `json:"id,omitempty"`
	From  geolocation `json:"from,omitempty"`
	To    geolocation `json:"to,omitempty"`
	Route route       `json:"route,omitempty"`
}
