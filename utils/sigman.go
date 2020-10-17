/*
Signals management utility (ready, running, stopped).
*/

package utils

type sensor interface {
	notify(status map[string][]string)
}

type sensorFunc func(status map[string][]string)

func (function sensorFunc) notify(status map[string][]string) {
	function(status)
}

//===========================================

type sensors []sensor

func (sensors *sensors) appendSensor(sensor sensor) {
	*sensors = append(*sensors, sensor)
}

func (sensors sensors) notifyAll(status map[string][]string) {
	for _, sensor := range sensors {
		sensor.notify(status)
	}
}
