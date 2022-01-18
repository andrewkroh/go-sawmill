package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	reg *prometheus.Registry
)

func init() {
	reg = prometheus.NewRegistry()
	reg.Register(collectors.NewGoCollector())
	reg.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{
		Namespace: "process",
	}))
}

func Register(c ...prometheus.Collector) {
	for _, coll := range c {
		reg.Register(coll)
	}
}

func UnregisterPipeline() {

}

func Listen() {
	httpHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		ErrorLog:            nil,
		ErrorHandling:       0,
		Registry:            nil,
		DisableCompression:  false,
		MaxRequestsInFlight: 0,
		Timeout:             0,
		EnableOpenMetrics:   false,
	})
	mux := http.NewServeMux()
	mux.Handle("/metrics", httpHandler)
	go http.ListenAndServe("localhost:9003", mux)
}
