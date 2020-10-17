/*
Processor core implementation.
*/

package utils

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type core struct {
	reg    *registry
	logger *logrus.Logger

	errors      chan error
	terminate   chan []string
	termination bool
}

func (core *core) run() error {
	go core.procsTerminator()
	go core.errorsHandler()

	var queue sync.WaitGroup
	for _, proc := range core.reg.listProcsMap(core.reg.ready) {
		if !proc.Service.SubService {
			queue.Add(1)
			go func(proc *process) {
				defer queue.Done()
				proc.wait()
				core.reg.updateStatus(proc, "running")
				err := proc.start()
				if err != nil && !proc.Service.IgnoreFailures {
					if !core.reg.isPermittedToBeKilled(proc.Name) {
						core.errors <- err
					}
				}
			}(proc)
		}
	}
	queue.Wait()

	return nil
}

func (core *core) errorsHandler() {
	for err := range core.errors {
		if err != nil {
			if core.termination {
				continue
			}
			core.termination = true

			core.logger.WithField("prefix", "core").Errorf("%s ; %#v", err.Error(), err)
			for _, proc := range core.reg.listProcsMap(core.reg.running) {
				core.stop(proc)
			}
		}
	}
}

func (core *core) procsTerminator() {
	for chanNames := range core.terminate {
		core.terminateNames(chanNames)
	}
}

func (core *core) terminateNames(names []string) {
	for _, name := range names {
		core.logger.WithField("prefix", "core").Warn(name)
		proc, status := core.reg.getProcess(name)
		switch status {
		case "ready":
			core.reg.updateStatus(proc, "stopped")
		case "running":
			core.stop(proc)
		case "stopped":
			continue
		}
	}
}

func (core *core) stop(proc *process) {
	proc.stop()
	core.reg.updateStatus(proc, "stopped")
}
