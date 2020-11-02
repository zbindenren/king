package king

import (
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// Register registers prometheus flag collectors that display configured flags as
// metrics. The genereated metrics are of the form:
//
//     kong_flag{program="progname", name="flagname", value="flagvalue"} 1
func (m Map) Register(program string, registerer prometheus.Registerer) Map {
	for _, c := range m.collectors(program) {
		registerer.MustRegister(c)
	}

	return m
}

func (m Map) collectors(program string) []prometheus.Collector {
	collectors := []prometheus.Collector{}

	for name, value := range m {
		if isRedacted(value) {
			continue
		}

		collectors = append(collectors, newFlagCollector(program, name, fmt.Sprintf("%v", value)))
	}

	return collectors
}

func newFlagCollector(program string, flags ...string) prometheus.Collector {
	return prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "kong_flag",
			Help: "A metric with a constant '1' value labeled by program, flag name and value",
			ConstLabels: prometheus.Labels{
				"program": program,
				"name":    flags[0],
				"value":   flags[1],
			},
		},
		func() float64 { return 1 },
	)
}

func isRedacted(value interface{}) bool {
	v, ok := value.(string)
	if !ok {
		return false
	}

	return v == strings.Repeat(string(redactChar), len(v))
}
