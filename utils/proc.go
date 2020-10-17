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
	name    string
	service *service
	logger  *logrus.Entry
	cancel  context.CancelFunc
	done    chan struct{}
}

func (proc *process) wait() {
	for len(proc.service.hooks["wait"]) != 0 {
		time.Sleep(time.Second)
	}
}

func (proc *process) update(status map[string][]string) {
	for _, procName := range status["stopped"] {
		for i, name := range proc.service.hooks["wait"] {
			if name == procName {
				proc.service.hooks["wait"] = remove(proc.service.hooks["wait"], i)
			}
		}
	}
}

func (proc *process) start() error {
	environs := proc.service.withOsEnvirons()
	cwd := proc.service.expandedEnv()

	command, err := proc.service.parsedCommand()
	if err != nil {
		return err
	}

	logout := &logger{
		proc.logger.WithField("prefix", proc.name).WriterLevel(logrus.InfoLevel),
	}
	logerr := &logger{
		proc.logger.WithField("prefix", proc.name).WriterLevel(logrus.WarnLevel),
	}

	shell, err := interp.New(
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
	ctx, proc.cancel = context.WithCancel(ctx)
	return shell.Run(ctx, command)
}

func (proc *process) stop() {
	if proc.cancel != nil {
		proc.cancel()
		proc.logger.WithField("prefix", proc.name).Warn("stopped by GoPM")
	}
}

// Remove element by index from a slice.
func remove(slice []string, i int) []string {
	return append(slice[:i], slice[i+1:]...)
}
