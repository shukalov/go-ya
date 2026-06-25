package models

// MetricType - тип метрики
type MetricType string

const (
	TypeGauge   MetricType = "gauge"
	TypeCounter MetricType = "counter"
)

// MetricDef - описание метрики
type MetricDef struct {
	Name        string
	Description string
	Type        MetricType
}

// RuntimeMetrics - метрики из runtime, построенные из MetricDefs
type RuntimeMetrics struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

var (
	gaugeDefs  []MetricDef
	counterDefs []MetricDef
)

func init() {
	all := MetricDefs()
	gaugeDefs = make([]MetricDef, 0, len(all))
	counterDefs = make([]MetricDef, 0, len(all))
	for _, d := range all {
		switch d.Type {
		case TypeGauge:
			gaugeDefs = append(gaugeDefs, d)
		case TypeCounter:
			counterDefs = append(counterDefs, d)
		}
	}
}

// MetricDefs возвращает определения всех метрик (единственный источник правды)
func MetricDefs() []MetricDef {
	return []MetricDef{
		{Name: "Alloc", Description: "Bytes of allocated heap objects", Type: TypeGauge},
		{Name: "BuckHashSys", Description: "Bytes in bucket hash tables", Type: TypeGauge},
		{Name: "Frees", Description: "Cumulative count of freed heap objects", Type: TypeGauge},
		{Name: "GCCPUFraction", Description: "Fraction of CPU time used by GC", Type: TypeGauge},
		{Name: "GCSys", Description: "Bytes of memory in GC metadata", Type: TypeGauge},
		{Name: "HeapAlloc", Description: "Bytes of allocated heap objects", Type: TypeGauge},
		{Name: "HeapIdle", Description: "Bytes in idle (unused) spans", Type: TypeGauge},
		{Name: "HeapInuse", Description: "Bytes in in-use spans", Type: TypeGauge},
		{Name: "HeapObjects", Description: "Number of allocated heap objects", Type: TypeGauge},
		{Name: "HeapReleased", Description: "Bytes of physical memory released to OS", Type: TypeGauge},
		{Name: "HeapSys", Description: "Bytes of heap memory obtained from OS", Type: TypeGauge},
		{Name: "LastGC", Description: "Time of last GC (nanoseconds since epoch)", Type: TypeGauge},
		{Name: "Lookups", Description: "Cumulative number of pointer lookups", Type: TypeGauge},
		{Name: "MCacheInuse", Description: "Bytes of mcache structures", Type: TypeGauge},
		{Name: "MCacheSys", Description: "Bytes of memory obtained from OS for mcache", Type: TypeGauge},
		{Name: "MSpanInuse", Description: "Bytes of mspan structures", Type: TypeGauge},
		{Name: "MSpanSys", Description: "Bytes of memory obtained from OS for mspan", Type: TypeGauge},
		{Name: "Mallocs", Description: "Cumulative count of allocated heap objects", Type: TypeGauge},
		{Name: "NextGC", Description: "Target heap size for next GC cycle", Type: TypeGauge},
		{Name: "NumForcedGC", Description: "Number of forced GC cycles", Type: TypeGauge},
		{Name: "NumGC", Description: "Number of completed GC cycles", Type: TypeGauge},
		{Name: "OtherSys", Description: "Bytes of memory in miscellaneous off-heap", Type: TypeGauge},
		{Name: "PauseTotalNs", Description: "Cumulative pause time in GC", Type: TypeGauge},
		{Name: "StackInuse", Description: "Bytes in stack spans", Type: TypeGauge},
		{Name: "StackSys", Description: "Bytes of stack memory obtained from OS", Type: TypeGauge},
		{Name: "Sys", Description: "Bytes of memory obtained from OS", Type: TypeGauge},
		{Name: "TotalAlloc", Description: "Cumulative bytes of allocated heap objects", Type: TypeGauge},
		{Name: "RandomValue", Description: "Random value for demonstration", Type: TypeGauge},
		{Name: "PollCount", Description: "Number of metric collection cycles", Type: TypeCounter},
	}
}

// GaugeDefs возвращает только gauge метрики
func GaugeDefs() []MetricDef {
	return gaugeDefs
}

// CounterDefs возвращает только counter метрики
func CounterDefs() []MetricDef {
	return counterDefs
}
