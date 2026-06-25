package models

// MetricType - тип метрики
type MetricType string

const (
	TypeGauge   MetricType = "gauge"
	TypeCounter MetricType = "counter"
)

// Metric - структура метрики для передачи между компонентами
type Metric struct {
	Type  MetricType
	Name  string
	Value interface{} // float64 для gauge, int64 для counter
}

// GaugeMetric - метрика типа gauge
type GaugeMetric struct {
	Name  string
	Value float64
}

// CounterMetric - метрика типа counter
type CounterMetric struct {
	Name  string
	Value int64
}

// RuntimeMetrics - структура для сбора метрик из runtime
type RuntimeMetrics struct {
	// Gauge метрики
	Alloc         float64
	BuckHashSys   float64
	Frees         float64
	GCCPUFraction float64
	GCSys         float64
	HeapAlloc     float64
	HeapIdle      float64
	HeapInuse     float64
	HeapObjects   float64
	HeapReleased  float64
	HeapSys       float64
	LastGC        float64
	Lookups       float64
	MCacheInuse   float64
	MCacheSys     float64
	MSpanInuse    float64
	MSpanSys      float64
	Mallocs       float64
	NextGC        float64
	NumForcedGC   float64
	NumGC         float64
	OtherSys      float64
	PauseTotalNs  float64
	StackInuse    float64
	StackSys      float64
	Sys           float64
	TotalAlloc    float64
	RandomValue   float64
	PollCount     int64
}
