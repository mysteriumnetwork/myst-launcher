/*
 * Copyright (C) 2021 The "MysteriumNetwork/node" Authors.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package daemon

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"strings"

	transport2 "github.com/mysteriumnetwork/myst-launcher/supervisor/daemon/transport"
	"github.com/mysteriumnetwork/myst-launcher/supervisor/model"
	"github.com/mysteriumnetwork/myst-launcher/supervisor/util"

	"github.com/rs/zerolog/log"
)

// Daemon - vm helper process.
type Daemon struct {
	cfg *model.Config
}

// New creates a new daemon.
func New(cfg *model.Config) Daemon {
	d := Daemon{}
	d.cfg = cfg

	return d
}

// Start the daemon. Blocks.
func (d *Daemon) Start(options transport2.Options) error {
	defer util.PanicHandler("dialog_")

	log.Info().Msgf("Daemon !Start > %v", options)
	return transport2.Start(d.dialog, options)
}

// dialog talks to the client via established connection.
func (d *Daemon) dialog(conn io.ReadWriteCloser) {
	log.Info().Msg("Daemon !dialog >>>")

	answer := responder{conn}
	lines := make(chan interface{})

	go func() {
		scan := bufio.NewScanner(conn)
		for scan.Scan() {
			b := scan.Bytes()
			lines <- b
		}
		lines <- scan.Err()
	}()

	log.Info().Msg("Daemon !dialog")

	for l := range lines {
		switch line := l.(type) {
		case []byte:
			log.Info().Msgf("Daemon !dialog: %v", string(line))

			m := make(map[string]interface{})
			_ = json.Unmarshal([]byte(line), &m)
			op := strings.ToLower(m["cmd"].(string))
			d.doOperation(op, answer, m)

		default:
			// no match;
		}
	}
}

func (d *Daemon) doOperation(op string, answer responder, m map[string]interface{}) {
	log.Info().Msg("Daemon !doOperation")

	switch op {
	case CommandVersion:
		answer.ok(nil)

	case CommandPing:
		answer.pong()

	case CommandSetupFW:
		answer.ok(nil)

	default:
		answer.err(errors.New("unknown command"))
	}
}
