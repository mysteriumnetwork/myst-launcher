package main

import (
	"os"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/mysteriumnetwork/myst-launcher/supervisor/daemon"
	"github.com/mysteriumnetwork/myst-launcher/supervisor/daemon/flags"
	"github.com/mysteriumnetwork/myst-launcher/supervisor/daemon/transport"
	"github.com/mysteriumnetwork/myst-launcher/supervisor/install"
	"github.com/mysteriumnetwork/myst-launcher/supervisor/interactive"
	"github.com/mysteriumnetwork/myst-launcher/supervisor/logconfig"
	"github.com/mysteriumnetwork/myst-launcher/supervisor/model"
	"github.com/mysteriumnetwork/myst-launcher/supervisor/util"
	"github.com/mysteriumnetwork/myst-launcher/supervisor/util/winutil"
)

func main() {
	defer util.PanicHandler("main")
	flags.Parse()

	if *flags.FlagInstall {
		initLogger()
		installSvc()

	} else if *flags.FlagUninstall {
		initLogger()
		uninstallSvc()

	} else if *flags.FlagWinService {
		initLogger()

		cfg := new(model.Config)
		// cfg.Read()
		svc := daemon.New(cfg)
		if err := svc.Start(transport.Options{WinService: *flags.FlagWinService}); err != nil {
			log.Fatal().Err(err).Msg("Error running service")
		}

	} else {
		interactive.Handler()
	}
}

func initLogger() {
	logOpts := logconfig.LogOptions{
		LogLevel: "info",
		Filepath: "",
	}
	if err := logconfig.Configure(logOpts); err != nil {
		log.Fatal().Err(err).Msg("Failed to configure logging")
	}

	workDir, err := winutil.AppDataDir("MystLauncherHelper")
	if err != nil {
		log.Fatal().Err(err).Msg("Error getting AppDataDir: " + err.Error())
	}
	err = os.Chdir(workDir)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to configure logging")
	}
}

func installSvc() error {
	path, err := util.ThisPath()
	if err != nil {
		return errors.Wrap(err, "Failed to determine service path")
	}
	options := install.Options{
		ExecuatblePath: path,
	}
	log.Info().Msgf("Installing dual-mode helper with options: %#v", options)
	if err = install.Install(options); err != nil {
		return errors.Wrap(err, "Failed to install service")
	}
	log.Info().Msg("Service installed")
	return nil
}

func uninstallSvc() error {
	log.Info().Msgf("Installing helper service")
	if err := install.Uninstall(); err != nil {
		return errors.Wrap(err, "Failed to uninstall helper service")
	}
	log.Info().Msg("Service uninstalled")
	return nil
}
