package metrics

var plugins []MetricPlugin

func Register(p MetricPlugin) {
	plugins = append(plugins, p)
}

func CollectAll() map[string]interface{} {
	out := make(map[string]interface{})

	for _, p := range plugins {
		data := p.Collect()
		for k, v := range data {
			out[k] = v
		}
	}

	return out
}
