package metrics

type MetricPlugin interface {
	Name() string
	Collect() map[string]interface{}
}
