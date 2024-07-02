package client

import (
	"encoding/json"
	"net"

	"github.com/mysteriumnetwork/myst-launcher/supervisor/model"
	"github.com/rs/zerolog/log"
)

func SendCommand(conn net.Conn, m model.KVMap) model.KVMap {
	b, _ := json.Marshal(m)
	log.Debug().Msgf("send: %v", string(b))
	conn.Write(b)
	conn.Write([]byte("\n"))

	out := make([]byte, 2000)

	// wait for response
	for {
		n, _ := conn.Read(out)
		if n > 0 {
			var res map[string]interface{}
			payload := out[:n-1]
			log.Debug().Msgf("rcv: %v", string(payload))

			json.Unmarshal(payload, &res)
			if res["resp"] == "error" || res["resp"] == "pong" {
				return res
			} else if res["resp"] == "ok" {
				return res
			}
		}
	}
	
}
