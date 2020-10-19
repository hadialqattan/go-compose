/*
Processes registering utility.
*/

package utils

import (
	"sync"

	goset "github.com/deckarep/golang-set"
	"github.com/sirupsen/logrus"
)

type registry struct {
	sensors
	sync.RWMutex

	ready        map[string]*process
	running      map[string]*process
	stopped      map[string]*process
	permitToKill goset.Set

	logger *logrus.Logger
}

func createRegistry() *registry {
	return &registry{
		ready:        map[string]*process{},
		running:      map[string]*process{},
		stopped:      map[string]*process{},
		permitToKill: goset.NewSet(),
		logger:       newCustomLogger(),
	}
}

func (reg *registry) register(proc *process) {
	reg.appendSensor(sensorFunc(proc.update))

	reg.Lock()
	defer reg.Unlock()

	reg.ready[proc.name] = proc
	for _, procName := range proc.service.Hooks["kill"] {
		reg.permitToKill.Add(procName)
	}
}

func (reg *registry) updateStatus(proc *process, status string) {
	reg.Lock()
	switch status {
	case "running":
		if _, find := reg.ready[proc.name]; find {
			delete(reg.ready, proc.name)
			reg.running[proc.name] = proc
			reg.logger.WithField("prefix", proc.name).Info("running")
		}
	case "stopped":
		if _, find := reg.running[proc.name]; find {
			delete(reg.running, proc.name)
			reg.stopped[proc.name] = proc
			reg.logger.WithField("prefix", proc.name).Info("stopped")
		}
	default:
		panic("WTF are you doing?! Unknown process!")
	}
	reg.Unlock()

	regStatus := reg.getStatus()
	reg.notifyAll(regStatus)
	reg.logger.WithField("prefix", "registry").Debug(regStatus)
}

func (reg *registry) getStatus() map[string][]string {
	status := make(map[string][]string)

	reg.RLock()
	defer reg.RUnlock()

	for name := range reg.ready {
		status["ready"] = append(status["ready"], name)
	}
	for name := range reg.running {
		status["running"] = append(status["running"], name)
	}
	for name := range reg.stopped {
		status["stopped"] = append(status["stopped"], name)
	}
	for _, name := range reg.permitToKill.ToSlice() {
		val := name.(string)
		status["permit_to_kill"] = append(status["permit_to_kill"], val)
	}
	return status
}

func (reg *registry) isPermittedToBeKilled(name string) bool {
	reg.RLock()
	defer reg.RUnlock()
	return reg.permitToKill.Contains(name)
}

func (reg *registry) getProcess(name string) (*process, string) {
	reg.RLock()
	defer reg.RUnlock()

	if proc, found := reg.ready[name]; found {
		return proc, "ready"
	} else if proc, found := reg.running[name]; found {
		return proc, "running"
	} else if proc, found := reg.stopped[name]; found {
		return proc, "stopped"
	}
	panic("WTF are you doing?! Unknown process!")
}

func (reg *registry) len() int {
	return len(reg.getProcesses())
}

func (reg *registry) getProcesses() []*process {
	return append(
		append(
			reg.listProcsMap(reg.ready),
			reg.listProcsMap(reg.running)...,
		), reg.listProcsMap(reg.stopped)...,
	)
}

func (reg *registry) listProcsMap(procsMap map[string]*process) []*process {
	reg.RLock()
	defer reg.RUnlock()

	procs := []*process{}
	for _, proc := range procsMap {
		procs = append(procs, proc)
	}
	return procs
}
