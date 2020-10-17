/*
Process struct and methods.
*/

package utils

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
)

type process struct {
	Name    string
	Service service
	Logger  *logrus.Entry
	Cancel  context.CancelFunc
	Done    chan struct{}
}

func (proc *process) wait() {
	for len(proc.Service.Hooks["wait"]) != 0 {
		time.Sleep(time.Second)
	}
}

func (proc *process) update(status map[string][]string) {
	for _, procName := range status["stopped"] {
		for i, name := range proc.Service.Hooks["wait"] {
			if name == procName {
				proc.Service.Hooks["wait"] = remove(proc.Service.Hooks["wait"], i)
			}
		}
	}
}

func (proc *process) start() error {
	environs := proc.Service.withOsEnvirons()
	cwd := proc.Service.expandedEnv()

	command, err := proc.Service.parsedCommand()
	if err != nil {
		return err
	}

	logout := &logger{
		proc.Logger.WithField("prefix", proc.Name).WriterLevel(logrus.InfoLevel),
	}
	logerr := &logger{
		proc.Logger.WithField("prefix", proc.Name).WriterLevel(logrus.WarnLevel),
	}

	bash, err := interp.New(
		interp.Dir(cwd),
		interp.Env(expand.ListEnviron(environs...)),
		interp.OpenHandler(func(ctx context.Context, path string, flag int, permission os.FileMode) (io.ReadWriteCloser, error) {
			return interp.DefaultOpenHandler()(ctx, path, flag, permission)
		}),
		interp.StdIO(os.Stdin, logout, logerr),
	)
	if err != nil {
		return err
	}

	ctx := context.Background()
	ctx, proc.Cancel = context.WithCancel(ctx)
	return bash.Run(ctx, command)
}

func (proc *process) stop() {
	if proc.Cancel != nil {
		proc.Cancel()
		proc.Logger.WithField("prefix", proc.Name).Warn("stopped by GoPM")
	}
}

// Remove element by index from a slice.
func remove(slice []string, i int) []string {
	return append(slice[:i], slice[i+1:]...)
}
