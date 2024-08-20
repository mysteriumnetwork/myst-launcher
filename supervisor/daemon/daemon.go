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

	"github.com/mysteriumnetwork/myst-launcher/controller/native"
	"github.com/mysteriumnetwork/myst-launcher/supervisor/daemon/transport"
	"github.com/mysteriumnetwork/myst-launcher/supervisor/model"
	"github.com/mysteriumnetwork/myst-launcher/supervisor/util"
	"github.com/mysteriumnetwork/myst-launcher/supervisor/util/winutil"

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
func (d *Daemon) Start(options transport.Options) error {
	defer util.PanicHandler("dialog_")

	log.Info().Msgf("Daemon !Start > %v", options)
	return transport.Start(d.dialog, options)
}

// dialog talks to the client via established connection.
func (d *Daemon) dialog(conn io.ReadWriteCloser) {
	log.Info().Msg("Daemon !dialog >")

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

	for l := range lines {
		switch line := l.(type) {
		case []byte:
			log.Info().Msgf("Daemon !dialog: %v", string(line))

			m := make(map[string]interface{})
			err := json.Unmarshal([]byte(line), &m)
			if err == nil {
				op := strings.ToLower(m["cmd"].(string))
				d.doOperation(op, answer, m, line)
			} else {
				log.Info().Msgf("Error %v", err)

				answer.err_("wrong string")
			}

		default:
			// no match;
		}
	}
}

type dtoCmdSetupFw struct {
	Sid     int    `json:"sid"`
	Exe     string `json:"exe"`
}

func (d *Daemon) doOperation(op string, answer responder, m map[string]interface{}, b []byte) {
	log.Info().Msg("Daemon !doOperation")

	switch op {
	case CommandVersion:
		answer.ok(nil)

	case CommandPing:
		answer.pong()

	case CommandSetupFW:
		dto := dtoCmdSetupFw{}
		err := json.Unmarshal(b, &dto)
		if err != nil || dto.Sid <= 0 || dto.Exe == "" {
			answer.err_("wrong json")
			return
		}

		closure := func() {
			ver := ""
			if d.cfg.V2Mode {
				ver = "2"
			}
			native.CheckAndInstallFirewallRules(ver, dto.Exe)
		}
		if !winutil.RunAsUserInThread(dto.Sid, closure) {
			answer.err_("setup firewall failed")
			return
		}
		answer.ok(nil)

	default:
		answer.err(errors.New("unknown command"))
	}
}
