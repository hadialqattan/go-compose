/*
Processor core implementation.
*/

package utils

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Processor is a struct that represents Go-compose single-core processor.
type Processor struct{ Core core }

// CreateProcessor is a function that creates a processor
// from the given services config.
func CreateProcessor(config *Config) *Processor {
	reg := createRegistry()
	for name, service := range config.services {
		reg.register(&process{name: name, service: service, logger: logrus.NewEntry(reg.logger)})
	}

	return &Processor{
		Core: core{
			reg:       reg,
			logger:    reg.logger,
			errors:    make(chan error, reg.len()),
			terminate: make(chan []string, reg.len()),
		},
	}
}

//===========================

type core struct {
	reg    *registry
	logger *logrus.Logger

	errors      chan error
	terminate   chan []string
	termination bool
	pool        sync.WaitGroup
}

// Run is a fuction that starts the core processes.
func (core *core) Run() {
	go core.procsTerminator()
	go core.errorsHandler()
	core.runProcesses(core.reg.listProcsMap(core.reg.ready))
	core.pool.Wait()
}

//==================================================================

func (core *core) runProcesses(procs []*process) {
	for _, proc := range procs {
		core.pool.Add(1)
		go core.start(proc)
	}
}

func (core *core) start(proc *process) {
	defer core.pool.Done()

	// Skip sub-services from autorun.
	if proc.service.SubService {
		return
	}

	proc.waitHook()
	core.reg.updateStatus(proc, "running")
	defer func() {
		core.reg.updateStatus(proc, "stopped")
		proc.startHook(core)
		proc.stopHook(core)
	}()

	err := proc.start()
	// Skip services that has `IgnoreFailures` flag.
	if err != nil && !proc.service.IgnoreFailures && !core.reg.isPermittedToBeKilled(proc.name) {
		core.errors <- err
		if proc.service.AutoRestart {
			core.logger.WithField("prefix", "core").Info("Rerun: ", proc.name)
			core.runProcesses([]*process{proc})
		}
	}
	time.Sleep(time.Second / 10)
}

func (core *core) errorsHandler() {
	for err := range core.errors {
		if err != nil {
			if core.termination {
				continue
			}
			core.termination = true

			core.logger.WithField("prefix", "core").Errorf("%s -> %#v", err.Error(), err)
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

func (core *core) shutdown() {
	core.logger.WithField("prefix", "core").Warn("Gracefully shutdown go-compose")
	core.termination = true
	for _, process := range core.reg.getProcesses() {
		core.stop(process)
	}
}

func (core *core) stop(proc *process) {
	proc.stop()
	core.reg.updateStatus(proc, "stopped")
}
