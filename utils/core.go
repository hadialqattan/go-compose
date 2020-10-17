/*
Processor core implementation.
*/

package utils

import (
	"sync"

	"github.com/sirupsen/logrus"
)

// Processor is a struct that represents GoPM single-core processor.
type Processor struct{ Core core }

// CreateProcessor is a function that creates a processor
// from the given services config.
func CreateProcessor(config *Config) *Processor {
	reg := createRegistry()
	for name, service := range config.services {
		reg.register(&process{name: name, service: service})
	}
	return &Processor{
		Core: core{
			reg:       reg,
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
}

// Run is a fuction that starts the core processes.
func (core *core) Run() {
	go core.procsTerminator()
	go core.errorsHandler()

	var pool sync.WaitGroup
	for _, proc := range core.reg.listProcsMap(core.reg.ready) {
		// Skip sub-services from autorun.
		if proc.service.subService {
			continue
		}
		pool.Add(1)
		go core.start(proc, &pool)
	}
	pool.Wait()
}

//==================================================================

func (core *core) start(proc *process, pool *sync.WaitGroup) {
	proc.wait()
	defer pool.Done()
	core.reg.updateStatus(proc, "running")

	err := proc.start()
	// Skip services that has `ignoreFailures` flag.
	if err != nil && !proc.service.ignoreFailures && !core.reg.isPermittedToBeKilled(proc.name) {
		core.errors <- err
	}
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

func (core *core) shutdown() {
	core.logger.Warn("[Gracefully shutdown GoPM]")
	core.termination = true
	for _, process := range core.reg.getProcesses() {
		core.stop(process)
	}
}

func (core *core) stop(proc *process) {
	proc.stop()
	core.reg.updateStatus(proc, "stopped")
}
