package models

// MetricDef - описание метрики
type MetricDef struct {
	Description string
}

// RuntimeMetrics - метрики из runtime, построенные из MetricDefs
type RuntimeMetrics struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

var (
	metricDefs   map[string]MetricDef
	gaugeDefs    map[string]MetricDef
	counterDefs  map[string]MetricDef
)

func init() {
	gaugeDefs = map[string]MetricDef{
		"Alloc":        {Description: "Bytes of allocated heap objects"},
		"BuckHashSys":  {Description: "Bytes in bucket hash tables"},
		"Frees":        {Description: "Cumulative count of freed heap objects"},
		"GCCPUFraction": {Description: "Fraction of CPU time used by GC"},
		"GCSys":        {Description: "Bytes of memory in GC metadata"},
		"HeapAlloc":    {Description: "Bytes of allocated heap objects"},
		"HeapIdle":     {Description: "Bytes in idle (unused) spans"},
		"HeapInuse":    {Description: "Bytes in in-use spans"},
		"HeapObjects":  {Description: "Number of allocated heap objects"},
		"HeapReleased": {Description: "Bytes of physical memory released to OS"},
		"HeapSys":      {Description: "Bytes of heap memory obtained from OS"},
		"LastGC":       {Description: "Time of last GC (nanoseconds since epoch)"},
		"Lookups":      {Description: "Cumulative number of pointer lookups"},
		"MCacheInuse":  {Description: "Bytes of mcache structures"},
		"MCacheSys":    {Description: "Bytes of memory obtained from OS for mcache"},
		"MSpanInuse":   {Description: "Bytes of mspan structures"},
		"MSpanSys":     {Description: "Bytes of memory obtained from OS for mspan"},
		"Mallocs":      {Description: "Cumulative count of allocated heap objects"},
		"NextGC":       {Description: "Target heap size for next GC cycle"},
		"NumForcedGC":  {Description: "Number of forced GC cycles"},
		"NumGC":        {Description: "Number of completed GC cycles"},
		"OtherSys":     {Description: "Bytes of memory in miscellaneous off-heap"},
		"PauseTotalNs": {Description: "Cumulative pause time in GC"},
		"StackInuse":   {Description: "Bytes in stack spans"},
		"StackSys":     {Description: "Bytes of stack memory obtained from OS"},
		"Sys":          {Description: "Bytes of memory obtained from OS"},
		"TotalAlloc":   {Description: "Cumulative bytes of allocated heap objects"},
		"RandomValue":  {Description: "Random value for demonstration"},
	}

	counterDefs = map[string]MetricDef{
		"PollCount": {Description: "Number of metric collection cycles"},
	}

	metricDefs = make(map[string]MetricDef, len(gaugeDefs)+len(counterDefs))
	for name, d := range gaugeDefs {
		metricDefs[name] = d
	}
	for name, d := range counterDefs {
		metricDefs[name] = d
	}
}

// MetricDefs возвращает определения всех метрик (единственный источник правды)
func MetricDefs() map[string]MetricDef {
	return metricDefs
}

// GaugeDefs возвращает gauge метрики
func GaugeDefs() map[string]MetricDef {
	return gaugeDefs
}

// CounterDefs возвращает counter метрики
func CounterDefs() map[string]MetricDef {
	return counterDefs
}
