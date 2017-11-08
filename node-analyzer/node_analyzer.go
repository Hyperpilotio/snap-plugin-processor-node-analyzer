package nodeanalyzer

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

type NodeAnalyzer struct {
	DerivedMetrics *DerivedMetrics
}

type MetricConfigs struct {
	Configs []DerivedMetricConfig `json:"configs" binding:"required"`
}

func downloadConfigFile(url string) (*MetricConfigs, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, errors.New("Unable to download config file: " + err.Error())
	}
	defer response.Body.Close()

	decoder := json.NewDecoder(response.Body)
	configs := MetricConfigs{}
	if err := decoder.Decode(&configs); err != nil {
		return nil, errors.New("Unable to decode body: " + err.Error())
	}

	return &configs, nil
}

// NewProcessor generate processor
func NewAnalyzer() plugin.Processor {
	return &NodeAnalyzer{}
}

// Process test process function
func (p *NodeAnalyzer) Process(mts []plugin.Metric, cfg plugin.Config) ([]plugin.Metric, error) {
	if p.DerivedMetrics == nil {
		configUrl, err := cfg.GetString("configUrl")
		if err != nil {
			return nil, errors.New("Unable to parse sampleInterval duration: " + err.Error())
		}

		configs, err := downloadConfigFile(configUrl)
		if err != nil {
			return nil, errors.New("Unable to download and deserialize configs: " + err.Error())
		}

		derivedMetrics, err := NewDerivedMetrics(configs.Configs)
		if err != nil {
			return nil, errors.New("Unable to create derived metrics: " + err.Error())
		}

		p.DerivedMetrics = derivedMetrics
	}

	newMetrics := []plugin.Metric{}
	for _, mt := range mts {
		currentTime := mt.Timestamp.UnixNano()
		metricName := "/" + strings.Join(mt.Namespace.Strings(), "/")
		derivedMetric := p.DerivedMetrics.ProcessMetric(currentTime, metricName, float64(mt.Data.(float32)))
		if derivedMetric != nil {
			namespaces := mt.Namespace.Strings()
			namespaces = append(namespaces, derivedMetric.Name)
			newMetric := plugin.Metric{
				Namespace: plugin.NewNamespace(namespaces...),
				Version:   mt.Version,
				Tags:      mt.Tags,
				Timestamp: mt.Timestamp,
				Data:      derivedMetric.Value,
			}

			newMetrics = append(newMetrics, newMetric)
		}
	}

	for _, newMetric := range newMetrics {
		mts = append(mts, newMetric)
	}

	return mts, nil
}

/*
	GetConfigPolicy() returns the configPolicy for your plugin.
	A config policy is how users can provide configuration info to
	plugin. Here you define what sorts of config info your plugin
	needs and/or requires.
*/
func (p *NodeAnalyzer) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()
	policy.AddNewStringRule([]string{"hyperpilot", "node-analyzer"},
		"configUrl",
		false,
		plugin.SetDefaultString(""))

	return *policy, nil
}
