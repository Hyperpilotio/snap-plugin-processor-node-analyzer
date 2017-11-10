package nodeanalyzer

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
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

func (p *NodeAnalyzer) ProcessMetrics(mts []plugin.Metric) ([]plugin.Metric, error) {
	newMetrics := []plugin.Metric{}
	for _, mt := range mts {
		if mt.Data == nil {
			continue
		}

		currentTime := mt.Timestamp.UnixNano()
		metricName := "/" + strings.Join(mt.Namespace.Strings(), "/")
		metricData := convertFloat64(mt.Data)
		derivedMetric, err := p.DerivedMetrics.ProcessMetric(currentTime, metricName, metricData)
		if err != nil {
			return nil, errors.New("Unable to process metric: " + err.Error())
		}

		if derivedMetric != nil {
			namespaces := mt.Namespace.Strings()
			namespaces = append(namespaces, derivedMetric.Name)
			mt.Tags["derived_metrics_process"] = "true"
			mt.Tags["average_data"] = strconv.FormatFloat(metricData, 'f', -1, 64)

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

	return p.ProcessMetrics(mts)
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

func convertFloat64(data interface{}) float64 {
	switch data.(type) {
	case int:
		return float64(data.(int))
	case int8:
		return float64(data.(int8))
	case int16:
		return float64(data.(int16))
	case int32:
		return float64(data.(int32))
	case int64:
		return float64(data.(int64))
	case uint64:
		return float64(data.(uint64))
	case float32:
		return float64(data.(float32))
	case float64:
		return float64(data.(float64))
	default:
		return float64(0)
	}
}
