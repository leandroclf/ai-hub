package httpserver

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"sync"
)

// Registry e um contador de metricas minimalista, deliberadamente sem
// dependencia externa de client Prometheus, expondo o mesmo formato de
// texto que o Prometheus sabe fazer scrape (OPE-06). Dimensoes
// permitidas: ambiente, componente, servico, provedor — nunca
// protocol_id/event_id/dado pessoal (essa regra e responsabilidade de
// quem chama Inc/Add, nao deste registry).
type Registry struct {
	mu       sync.Mutex
	counters map[string]float64
}

// Observe exports cumulative buckets and count/sum for bounded route dimensions.
func (r *Registry) Observe(name string, labels map[string]string, value float64) {
	bounds := []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30}
	r.mu.Lock()
	defer r.mu.Unlock()
	bucketLabels := make(map[string]string, len(labels)+1)
	for k, v := range labels {
		bucketLabels[k] = v
	}
	for _, bound := range bounds {
		bucketLabels["le"] = strconv.FormatFloat(bound, 'g', -1, 64)
		bucketKey := key(name+"_bucket", bucketLabels)
		if value <= bound {
			r.counters[bucketKey]++
		} else {
			r.counters[bucketKey] += 0
		}
	}
	bucketLabels["le"] = "+Inf"
	r.counters[key(name+"_bucket", bucketLabels)]++
	r.counters[key(name+"_count", labels)]++
	r.counters[key(name+"_sum", labels)] += value
}

// NewRegistry cria um registry vazio.
func NewRegistry() *Registry {
	return &Registry{counters: make(map[string]float64)}
}

func key(name string, labels map[string]string) string {
	if len(labels) == 0 {
		return name
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	s := name + "{"
	for i, k := range keys {
		if i > 0 {
			s += ","
		}
		s += fmt.Sprintf("%s=%q", k, labels[k])
	}
	return s + "}"
}

// Inc incrementa um contador de baixa cardinalidade em 1.
func (r *Registry) Inc(name string, labels map[string]string) {
	r.Add(name, labels, 1)
}

// Add incrementa um contador de baixa cardinalidade em delta.
func (r *Registry) Add(name string, labels map[string]string, delta float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.counters[key(name, labels)] += delta
}

// ServeHTTP expoe os contadores no formato texto do Prometheus.
func (r *Registry) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	r.mu.Lock()
	defer r.mu.Unlock()
	keys := make([]string, 0, len(r.counters))
	for k := range r.counters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	for _, k := range keys {
		fmt.Fprintf(w, "%s %g\n", k, r.counters[k])
	}
}
