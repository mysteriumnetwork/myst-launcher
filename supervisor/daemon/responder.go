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
	"encoding/json"
	"fmt"
	"io"

	"github.com/rs/zerolog/log"
)

type responder struct {
	io.Writer
}

type Result struct {
	Cmd  string      `json:"cmd,omitempty"`
	Resp string      `json:"resp,omitempty"`
	Err  string      `json:"err,omitempty"`
	Data interface{} `json:"data,omitempty"`
}

func (r *responder) ok(data interface{}) {
	m := Result{
		Resp: "ok",
		Data: data,
	}
	b, _ := json.Marshal(m)
	r._message(string(b))
}

func (r *responder) err(err error) {
	m := Result{
		Resp: "error",
		Err:  err.Error(),
	}
	b, _ := json.Marshal(m)
	r._message(string(b))
}

func (r *responder) pong() {
	m := Result{
		Resp: "pong",
	}
	b, _ := json.Marshal(m)
	r._message(string(b))
}

func (r *responder) _message(msg string) {
	log.Debug().Msgf("to client: %s", msg)
	if _, err := fmt.Fprintln(r, msg); err != nil {
		log.Err(err).Msgf("Could not send message: %q", msg)
	}
}
