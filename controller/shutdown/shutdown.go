package shutdown

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/advbet/sseclient"
	"github.com/asaskevich/EventBus"
)

type x struct {
	Type    string `json:"type"`
	Payload struct {
		SessionsStats struct {
			CountConsumers int `json:"count_consumers"`
		} `json:"sessions_stats"`
	} `json:"payload"`
}

type ShutdownController struct {
	sse    *sseclient.Client
	cancel context.CancelFunc
	bus    EventBus.Bus
}

func NewShutdownController(bus EventBus.Bus) *ShutdownController {
	return &ShutdownController{
		bus: bus,
	}
}

func (s *ShutdownController) eventHandler(event *sseclient.Event) error {
	// log.Printf("event : %s : %s : %s", event.ID, event.Event, event.Data)
	x := x{}

	// set default not-zero value
	x.Payload.SessionsStats.CountConsumers = -1
	json.Unmarshal(event.Data, &x)
	if x.Type == "state-change" {
		fmt.Println(x)

		if x.Payload.SessionsStats.CountConsumers == 0 {
			s.bus.Publish("ready-to-shutdown")
		}
	}

	return nil
}

func (s *ShutdownController) Start() {
	addr := "http://localhost:4050/events/state"

	if s.sse == nil {
		c := sseclient.New(addr, "")
		ctx, cancel := context.WithCancel(context.Background())
		s.cancel = cancel
		s.sse = c

		go c.Start(ctx, s.eventHandler, sseclient.ReconnectOnError)
	}
}

func (s *ShutdownController) Stop() {
	if s.cancel != nil {
		s.cancel()
		s.sse = nil
	}
}
