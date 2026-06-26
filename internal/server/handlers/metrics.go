package handlers

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shukalov/go-ya/internal/server/storage"
	"github.com/shukalov/go-ya/pkg/models"
)

//go:embed templates/*.html
var templateFS embed.FS

var indexTemplate = template.Must(template.ParseFS(templateFS, "templates/index.html"))

type templateMetric struct {
	Name  string
	Kind  string
	Value string
	Desc  string
}

type indexData struct {
	Metrics []templateMetric
}

// MetricsHandler - обработчик для метрик
type MetricsHandler struct {
	storage storage.Storage
}

// NewMetricsHandler - создает новый обработчик
func NewMetricsHandler(storage storage.Storage) *MetricsHandler {
	return &MetricsHandler{
		storage: storage,
	}
}

// isValidMetricName - проверяет валидность имени метрики
func isValidMetricName(name string) bool {
	if name == "" {
		return false
	}
	for _, ch := range name {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '_' || ch == '.') {
			return false
		}
	}
	return true
}

// Update - обработчик для обновления метрик
func (h *MetricsHandler) Update(c *gin.Context) {
	// if c.Request.Header.Get("Content-Type") != "text/plain" {
	// 	c.String(http.StatusBadRequest, "Invalid Content-Type")
	// 	return
	// }

	metricType := c.Param("type")
	metricName := c.Param("name")
	valueStr := c.Param("value")

	if !isValidMetricName(metricName) {
		c.String(http.StatusNotFound, "Metric name is required")
		return
	}

	switch metricType {
	case "gauge":
		value, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("Invalid gauge value: '%s'", valueStr))
			return
		}
		if err := h.storage.UpdateGauge(metricName, value); err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Storage error: %v", err))
			return
		}

	case "counter":
		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("Invalid counter value: '%s'", valueStr))
			return
		}
		if err := h.storage.UpdateCounter(metricName, value); err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Storage error: %v", err))
			return
		}

	default:
		c.String(http.StatusBadRequest, "Invalid metric type. Allowed: gauge, counter")
		return
	}

	c.String(http.StatusOK, metricType)
}

// Get - обработчик для получения значений метрик
func (h *MetricsHandler) Get(c *gin.Context) {
	metricType := c.Param("type")
	metricName := c.Param("name")

	if !isValidMetricName(metricName) {
		c.Status(http.StatusNotFound)
		return
	}

	switch metricType {
	case "gauge":
		value, ok, err := h.storage.GetGauge(metricName)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Storage error: %v", err))
			return
		}
		if !ok {
			c.Status(http.StatusNotFound)
			return
		}
		c.String(http.StatusOK, "%s", strconv.FormatFloat(value, 'f', -1, 64))

	case "counter":
		value, ok, err := h.storage.GetCounter(metricName)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Storage error: %v", err))
			return
		}
		if !ok {
			c.Status(http.StatusNotFound)
			return
		}
		c.String(http.StatusOK, "%d", value)

	default:
		c.String(http.StatusBadRequest, "Invalid metric type")
	}
}

// Index - возвращает HTML-страницу со списком метрик
func (h *MetricsHandler) Index(c *gin.Context) {
	allMetrics, err := h.storage.GetAllMetrics()
	if err != nil {
		c.String(http.StatusInternalServerError, "Storage error")
		return
	}

	data := indexData{}
	gaugeDefs := models.GaugeDefs()
	counterDefs := models.CounterDefs()

	for name, v := range allMetrics.Gauges {
		desc := ""
		if def, ok := gaugeDefs[name]; ok {
			desc = def.Description
		}
		data.Metrics = append(data.Metrics, templateMetric{
			Name: name, Kind: "gauge", Value: strconv.FormatFloat(v, 'f', -1, 64), Desc: desc,
		})
	}
	for name, v := range allMetrics.Counters {
		desc := ""
		if def, ok := counterDefs[name]; ok {
			desc = def.Description
		}
		data.Metrics = append(data.Metrics, templateMetric{
			Name: name, Kind: "counter", Value: strconv.FormatInt(v, 10), Desc: desc,
		})
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := indexTemplate.Execute(c.Writer, data); err != nil {
		c.String(http.StatusInternalServerError, "Template error")
	}
}

// Metrics - возвращает метрики в формате Prometheus
func (h *MetricsHandler) Metrics(c *gin.Context) {
	allMetrics, err := h.storage.GetAllMetrics()
	if err != nil {
		c.String(http.StatusInternalServerError, "Storage error")
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Writer.WriteHeader(http.StatusOK)

	gaugeDefs := models.GaugeDefs()
	counterDefs := models.CounterDefs()

	for name, v := range allMetrics.Gauges {
		if def, ok := gaugeDefs[name]; ok {
			fmt.Fprintf(c.Writer, "# HELP %s %s\n# TYPE %s gauge\n%s %s\n\n",
				name, def.Description, name, name,
				strconv.FormatFloat(v, 'f', -1, 64))
		} else {
			fmt.Fprintf(c.Writer, "%s %s\n", name, strconv.FormatFloat(v, 'f', -1, 64))
		}
	}
	for name, v := range allMetrics.Counters {
		if def, ok := counterDefs[name]; ok {
			fmt.Fprintf(c.Writer, "# HELP %s %s\n# TYPE %s counter\n%s %d\n\n",
				name, def.Description, name, name, v)
		} else {
			fmt.Fprintf(c.Writer, "%s %d\n", name, v)
		}
	}
}
