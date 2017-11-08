package nodeanalyzer

import (
	"errors"
	"strings"
	"time"

	"github.com/gobwas/glob"
)

type DerivedMetric struct {
	Name  string
	Value float64
}

type DerivedMetricCalculator interface {
	GetDerivedMetric(currentTime int64, value float64) *DerivedMetric
	GetConfig() *DerivedMetricConfig
}

type DerivedMetricConfig struct {
	Metric string
	Name   string

	ThresholdBased *ThresholdBasedConfig
}

type ThresholdBasedConfig struct {
	Window         time.Duration
	Threshold      float64
	SampleInterval int64
}

type ThresholdBasedState struct {
	Config     *DerivedMetricConfig
	Hits       []int64
	TotalCount int64
}

func NewThresholdBasedState(config *DerivedMetricConfig) *ThresholdBasedState {
	totalCount := config.ThresholdBased.Window.Nanoseconds() / config.ThresholdBased.SampleInterval
	return &ThresholdBasedState{
		Config:     config,
		Hits:       make([]int64, 0),
		TotalCount: totalCount,
	}
}

func NewDerivedMetricCalculator(config *DerivedMetricConfig) (DerivedMetricCalculator, error) {
	if config.ThresholdBased != nil {
		return NewThresholdBasedState(config), nil
	}

	return nil, errors.New("No metric config found")
}

func (tbs *ThresholdBasedState) GetConfig() *DerivedMetricConfig {
	return tbs.Config
}

func (tbs *ThresholdBasedState) GetDerivedMetric(currentTime int64, value float64) *DerivedMetric {
	isAboveThreshold := value >= tbs.Config.ThresholdBased.Threshold
	if isAboveThreshold {
		hitsLength := len(tbs.Hits)
		if hitsLength > 0 {
			// Fill in missing metric points
			for currentTime-tbs.Hits[len(tbs.Hits)-1] >= 2*tbs.Config.ThresholdBased.SampleInterval {
				filledHitTime := tbs.Hits[len(tbs.Hits)-1] + tbs.Config.ThresholdBased.SampleInterval
				tbs.Hits = append(tbs.Hits, filledHitTime)
			}
		}

		// Prune values outside of window
		windowBeginTime := currentTime - tbs.Config.ThresholdBased.Window.Nanoseconds()
		lastGoodIndex := -1
		for i, hit := range tbs.Hits {
			if hit >= windowBeginTime {
				lastGoodIndex = i
				break
			}
		}

		if lastGoodIndex == -1 {
			// All values are outside of window, clear all values
			tbs.Hits = []int64{}
		} else {
			tbs.Hits = tbs.Hits[lastGoodIndex:]
		}

		tbs.Hits = append(tbs.Hits, currentTime)
	}

	return &DerivedMetric{
		Name:  tbs.Config.Name,
		Value: float64(len(tbs.Hits)) / float64(tbs.TotalCount),
	}
}

type DerivedMetrics struct {
	States map[string]DerivedMetricCalculator
}

func NewDerivedMetrics(configs []DerivedMetricConfig) (*DerivedMetrics, error) {
	states := map[string]DerivedMetricCalculator{}
	for _, config := range configs {
		calculator, err := NewDerivedMetricCalculator(&config)
		if err != nil {
			return nil, errors.New("Unable to create derived metric calculator: " + err.Error())
		}
		states[config.Metric] = calculator
	}

	return &DerivedMetrics{
		States: states,
	}, nil
}

func (ae *DerivedMetrics) ProcessMetric(currentTime int64, metricName string, value float64) *DerivedMetric {
	state, ok := ae.States[metricName]
	if !ok {
		for statesName, calculator := range ae.States {
			// Assumes the /* wildcard that accepts the last metric
			if !strings.HasSuffix(statesName, "/*") {
				continue
			}

			if isKeywordMatch(metricName, statesName) {
				calculator, err := NewDerivedMetricCalculator(calculator.GetConfig())
				if err != nil {
					return nil
				}
				ae.States[metricName] = calculator
				state = calculator
			}
		}

		if state == nil {
			return nil
		}
	}

	return state.GetDerivedMetric(currentTime, value)
}

func isKeywordMatch(keyword string, pattern string) bool {
	return glob.MustCompile(pattern).Match(keyword)
}
