package nodeanalyzer

import (
	"errors"
	"time"
)

type AlertConfig struct {
	Metric    string
	Window    time.Duration
	Type      string
	Threshold float32
}

type ThresholdAlert struct {
	Average float32
}

type AlertState struct {
	AlertConfig    *AlertConfig
	Hits           []int64
	TargetCount    int
	SampleInterval int64
}

func (as *AlertState) ProcessValue(currentTime int64, value float32) *ThresholdAlert {
	isAboveThreshold := value >= as.AlertConfig.Threshold
	if isAboveThreshold {
		hitsLength := len(as.Hits)
		if hitsLength > 0 {
			// Fill in missing metric points
			for currentTime-as.Hits[len(as.Hits)-1] >= 2*as.SampleInterval {
				filledHitTime := as.Hits[len(as.Hits)-1] + as.SampleInterval
				as.Hits = append(as.Hits, filledHitTime)
			}
		}

		// Prune values outside of window
		windowBeginTime := currentTime - as.AlertConfig.Window.Nanoseconds()
		lastGoodIndex := -1
		for i, hit := range as.Hits {
			if hit >= windowBeginTime {
				lastGoodIndex = i
				break
			}
		}

		if lastGoodIndex == -1 {
			// All values are outside of window, clear all values
			as.Hits = []int64{}
		} else {
			as.Hits = as.Hits[lastGoodIndex:]
		}

		as.Hits = append(as.Hits, currentTime)

		// Check if we should create new alert
		if len(as.Hits) >= as.TargetCount {
			totalInterval := as.Hits[len(as.Hits)-1] - as.Hits[0] + as.SampleInterval
			as.Hits = []int64{}
			return &ThresholdAlert{
				Average: float32(time.Duration(as.SampleInterval).Seconds() / time.Duration(totalInterval).Seconds()),
			}
		}
	}

	return nil
}

type AlertEvaluator struct {
	AlertStates map[string]*AlertState
}

func NewAlertEvaluator(alertConfigs []*AlertConfig, alertRatio float64, sampleInterval int64) (*AlertEvaluator, error) {
	if alertRatio == 0 {
		return nil, errors.New("Alert ratio has to be larger than zero")
	}

	if sampleInterval <= 0 {
		return nil, errors.New("Sample interval has to be larger than zero")
	}

	alertStates := map[string]*AlertState{}
	for _, alertConfig := range alertConfigs {
		targetCount := int(float64(alertConfig.Window.Nanoseconds()/sampleInterval) * alertRatio)
		alertStates[alertConfig.Metric] = &AlertState{
			AlertConfig:    alertConfig,
			Hits:           make([]int64, 0),
			TargetCount:    targetCount,
			SampleInterval: sampleInterval,
		}
	}

	return &AlertEvaluator{
		AlertStates: alertStates,
	}, nil
}

func (ae *AlertEvaluator) ProcessMetric(currentTime int64, metricName string, value float32) *ThresholdAlert {
	alertState, ok := ae.AlertStates[metricName]
	if !ok {
		return nil
	}

	return alertState.ProcessValue(currentTime, value)
}
