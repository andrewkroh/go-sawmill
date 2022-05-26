// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var reg *prometheus.Registry

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
