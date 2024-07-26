package events

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/3atlab/netroub/pkg/model"
	"github.com/alexei-led/pumba/pkg/chaos"
	"github.com/alexei-led/pumba/pkg/chaos/docker"
	"github.com/alexei-led/pumba/pkg/chaos/netem"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func ExecPumbaCommand(index int) error {
	dur, err := time.ParseDuration(model.Scenar.Event[index].PumbaCommand.Options.Duration)
	if err != nil {
		return err
	}

	globalParams := chaos.GlobalParams{
		Random:     false,
		Labels:     nil,
		Pattern:    "",
		Names:      []string{model.Scenar.Event[index].Host},
		Interval:   0,
		DryRun:     false,
		SkipErrors: false,
	}
	netemParams := netem.Params{
		Iface:    model.Scenar.Event[index].PumbaCommand.Options.Interface,
		Ips:      nil,
		Sports:   nil,
		Dports:   nil,
		Duration: dur,
		Image:    "",
		Pull:     true,
		Limit:    0,
	}

	ctx := handleSignals()

	delayCmd, err := parseCommands(index, globalParams, netemParams)
	if err != nil {
		return errors.Wrap(err, "error creating netem delay command")
	}

	err = chaos.RunChaosCommand(ctx, delayCmd, &globalParams)
	if err != nil {
		return errors.Wrap(err, "error running netem delay command")

	}
	return nil
}

func handleSignals() context.Context {
	// Graceful shut-down on SIGINT/SIGTERM
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// create cancelable context
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer cancel()
		sid := <-sig
		logrus.Debugf("Received signal: %d\n", sid)
		logrus.Debug("Canceling running chaos commands ...")
		logrus.Debug("Gracefully exiting after some cleanup ...")
	}()

	return ctx
}

func parseCommands(index int, globalParams chaos.GlobalParams, netemParams netem.Params) (chaos.Command, error) {

	cmdOption := model.Scenar.Event[index].PumbaCommand.Options

	switch model.Scenar.Event[index].PumbaCommand.Name {
	case "delay":
		return netem.NewDelayCommand(chaos.DockerClient, &globalParams, &netemParams, cmdOption.Time, cmdOption.Jitter, cmdOption.Correlation, cmdOption.Distribution)
	case "corrupt":
		return netem.NewCorruptCommand(chaos.DockerClient, &globalParams, &netemParams, cmdOption.Percent, cmdOption.Correlation)
	case "duplicate":
		return netem.NewDuplicateCommand(chaos.DockerClient, &globalParams, &netemParams, cmdOption.Percent, cmdOption.Correlation)
	case "loss":
		return netem.NewLossCommand(chaos.DockerClient, &globalParams, &netemParams, cmdOption.Percent, cmdOption.Correlation)
	case "stop":
		return docker.NewPauseCommand(chaos.DockerClient, &globalParams, netemParams.Duration, cmdOption.Limit), nil
	case "pause":
		return docker.NewStopCommand(chaos.DockerClient, &globalParams, true, netemParams.Duration, 0, cmdOption.Limit), nil
	case "rate":
		return netem.NewRateCommand(chaos.DockerClient, &globalParams, &netemParams, cmdOption.Rate, cmdOption.PacketOverhead, cmdOption.CellSize, cmdOption.CellOverhead)
	default:
		return nil, nil
	}

}
