package interactive

import (
	"fmt"
	"net"
	"time"

	"github.com/Microsoft/go-winio"
	consts "github.com/mysteriumnetwork/myst-launcher/supervisor/const"
	"github.com/mysteriumnetwork/myst-launcher/supervisor/daemon"
	"github.com/mysteriumnetwork/myst-launcher/supervisor/daemon/client"
	"github.com/mysteriumnetwork/myst-launcher/supervisor/model"
	"github.com/mysteriumnetwork/myst-launcher/supervisor/util"
	"github.com/mysteriumnetwork/myst-launcher/utils"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func connect() (net.Conn, error) {
	var conn net.Conn
	err := utils.Retry(3, time.Second, func() error {
		var err error

		conn, err = winio.DialPipe(consts.Sock, nil)
		return err
	})
	if err != nil {
		log.Err(err).Msg("error listening")
		return nil, err
	}
	return conn, nil
}

func Handler() {

	conn, err := connect()
	if err != nil {
		log.Fatal().Err(err).Msg("Connect")
	}

	for {
		fmt.Print("\n> ")
		fmt.Println("Test mode. select an action")
		fmt.Println("----------------------------------------------")
		fmt.Println("1  Setup firewall")
		fmt.Println("7  Exit")
		fmt.Print("\n> ")
		k := util.ReadConsole()

		switch k {
		case "1":
			err = sendCmdSetupFirewall(conn)
			if err != nil {
				log.Info().Err(err).Msg("setupFirewall")
			}

		case "7":
			return
		}
	}
}

func sendCmdSetupFirewall(conn net.Conn) error {
	cmd := model.KVMap{
		"cmd": daemon.CommandSetupFW,
	}
	res := client.SendCommand(conn, cmd)
	if res["resp"] == "error" {
		errStr := res["err"].(string)
		return errors.New(errStr)
	}
	return nil
}
