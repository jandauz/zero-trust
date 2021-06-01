package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/service/common"
	"github.com/dapr/go-sdk/service/http"
	"github.com/golang-sql/civil"
	"github.com/ohler55/ojg/oj"
	"github.com/rs/xid"
)

func main() {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

	c, err := client.NewClient()
	if err != nil {
		logger.Fatal(err)
	}
	defer c.Close()

	secrets, err := c.GetSecret(context.Background(), "kubernetes", "mongo-mongodb", nil)
	if err != nil {
		logger.Fatal(err)
	}

	pwd, ok := secrets["mongodb-password"]
	if !ok {
		logger.Fatal("mongodb-password missing")
	}

	h := handler{
		logger: logger,
		pwd:    pwd,
		c:      c,
		parser: oj.Parser{},
	}

	s := http.NewService(":8080")
	err = s.AddServiceInvocationHandler("schedule-delivery", h.scheduleDelivery())
	if err != nil {
		logger.Fatal(err)
	}

	err = s.AddServiceInvocationHandler("find-courier", h.findCourier())
	if err != nil {
		logger.Fatal(err)
	}

	logger.Printf("listening on :8080")
	logger.Fatal(s.Start())
}

type handler struct {
	logger *log.Logger
	pwd    string
	c      client.Client
	parser oj.Parser
}

type serviceInvocationHandler func(context.Context, *common.InvocationEvent) (*common.Content, error)

func (h handler) scheduleDelivery() serviceInvocationHandler {
	return func(ctx context.Context, in *common.InvocationEvent) (*common.Content, error) {
		h.logger.Printf("Invocation (ContentType:%s, Verb:%s, QueryString:%s, Data:%s",
			in.ContentType, in.Verb, in.QueryString, string(in.Data))

		h.logger.Println("scheduling delivery")
		var d delivery
		err := h.parser.Unmarshal(in.Data, &d)
		if err != nil {
			h.logger.Println(err)
			return nil, err
		}

		id := xid.New().String()
		h.logger.Printf("assigning delivery to courier: %v", id)
		c := courier{
			ID:        id,
			Delivery:  d,
			Status:    Scheduled,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		b, err := json.Marshal(c)
		if err != nil {
			h.logger.Println(err)
			return nil, err
		}
		h.logger.Println("storing courier state")
		err = h.c.SaveState(ctx, "courier-store", id, b)
		if err != nil {
			h.logger.Println(err)
			return nil, err
		}

		h.logger.Printf("scheduled courier: %s", id)
		return &common.Content{
			ContentType: in.ContentType,
			Data:        []byte(id),
		}, nil
	}
}

type courier struct {
	ID        string    `json:"id"`
	Delivery  delivery  `json:"delivery"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type delivery struct {
	ID            string         `json:"id"`
	From          geolocation    `json:"from"`
	To            geolocation    `json:"to"`
	Route         route          `json:"route"`
	DeliveryDate  *civil.Date    `json:"delivery_date,omitempty"`
	TimeRemaining *time.Duration `json:"time_remaining,omitempty"`
}

type geolocation struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type route struct {
	Distance            int           `json:"distance"`
	EstimatedTravelTime time.Duration `json:"estimated_travel_time"`
}

type Status int

const (
	Scheduled Status = iota
	Enroute
	Delivered
)

func (s Status) String() string {
	return statusStrings[s]
}

var statusStrings = map[Status]string{
	Scheduled: "Scheduled",
	Enroute:   "Enroute",
	Delivered: "Delivered",
}

var statuses = map[string]Status{
	"Scheduled": Scheduled,
	"Enroute":   Enroute,
	"Delivered": Delivered,
}

func (s Status) MarshalJSON() ([]byte, error) {
	b := bytes.NewBufferString(`"`)
	b.WriteString(statusStrings[s])
	b.WriteString(`"`)
	return b.Bytes(), nil
}

func (s *Status) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}

	*s = statuses[str]
	return nil
}

func (h *handler) findCourier() serviceInvocationHandler {
	return func(ctx context.Context, in *common.InvocationEvent) (*common.Content, error) {
		h.logger.Printf("Invocation (ContentType:%s, Verb:%s, QueryString:%s, Data:%s",
			in.ContentType, in.Verb, in.QueryString, string(in.Data))

		params, err := url.ParseQuery(in.QueryString)
		if err != nil {
			h.logger.Println(err)
			return nil, err
		}

		id := params.Get("id")
		h.logger.Printf("finding courier: %v", id)
		data, err := h.c.GetState(ctx, "courier-store", id)
		if err != nil {
			h.logger.Println(err)
			return nil, err
		}

		return &common.Content{
			ContentType: in.ContentType,
			Data:        data.Value,
		}, nil
	}
}
