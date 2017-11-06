package nodeanalyzer

import (
	"errors"
	"strings"
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

type NodeAnalyzer struct {
	*AlertConfig
	GlobalThreshold int
}

// NewProcessor generate processor
func NewAnalyzer() plugin.Processor {
	return &NodeAnalyzer{}
}

// Process test process function
func (p *NodeAnalyzer) Process(mts []plugin.Metric, cfg plugin.Config) ([]plugin.Metric, error) {
	if p.AlertConfig == nil {
		window := cfg["window"].(string)
		windowTime, err := time.ParseDuration(window)
		if err != nil {
			return nil, errors.New("Unable to parse window duration: " + err.Error())
		}

		p.AlertConfig = &AlertConfig{
			Metric:    cfg["metric"].(string),
			Window:    windowTime,
			Type:      cfg["type"].(string),
			Threshold: cfg["threshold"].(float32),
		}
	}

	alertEvaluator := NewAlertEvaluator([]*AlertConfig{p.AlertConfig}, p.GlobalThreshold)
	metrics := []plugin.Metric{}
	for _, mt := range mts {
		namespaces := mt.Namespace.Strings()
		metricName := "/" + strings.Join(namespaces[:len(namespaces)-2], "/")
		currentTime := mt.Timestamp.UnixNano()
		thresholdAlert := alertEvaluator.ProcessMetric(currentTime, metricName, mt.Data.(float32))
		if thresholdAlert != nil {
			metrics = append(metrics, mt)
		}
	}

	return metrics, nil
}

/*
	GetConfigPolicy() returns the configPolicy for your plugin.
	A config policy is how users can provide configuration info to
	plugin. Here you define what sorts of config info your plugin
	needs and/or requires.
*/
func (p *NodeAnalyzer) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()
	return *policy, nil
}
