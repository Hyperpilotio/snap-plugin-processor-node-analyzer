package nodeanalyzer

import "time"

type AlertConfig struct {
	Metric    string
	Window    time.Duration
	Type      string
	Threshold float32
}

type ThresholdAlert struct {
}

type HitInterval struct {
	StartTime int64
	Interval  int64
}

type AlertState struct {
	AlertConfig            *AlertConfig
	HitIntervals           []HitInterval
	TotalHitInterval       int64
	CurrentStartHitTime    int64
	IntervalThresholdRatio float32

	PreviousStartTime int64
	PreviousValue     float32
}

func (as *AlertState) pruneIntervals(currentTime int64) {
	beginningWindowTime := currentTime - as.AlertConfig.Window.Nanoseconds()

	newBeginningIndex := -1
	// Prune the intervals to only include intervals within the current window time.
	// This will update TotalHitInterval to only count intervals that's within
	// this window.
	for i, hitInterval := range as.HitIntervals {
		if hitInterval.StartTime < beginningWindowTime {
			extraInterval := beginningWindowTime - hitInterval.StartTime
			if extraInterval > hitInterval.Interval {
				// This interval is no longer needed
				newBeginningIndex = i + 1
				as.TotalHitInterval -= hitInterval.Interval
			} else {
				hitInterval.StartTime = beginningWindowTime
				hitInterval.Interval -= extraInterval
				as.TotalHitInterval -= extraInterval
			}
		} else {
			break
		}
	}

	// Remove intervals that is no longer in window
	if newBeginningIndex != -1 {
		if newBeginningIndex == len(as.HitIntervals)-1 {
			as.HitIntervals = []HitInterval{}
		} else {
			as.HitIntervals = as.HitIntervals[newBeginningIndex:]
		}
	}
}

func (as *AlertState) ProcessValue(currentTime int64, value float32) *ThresholdAlert {
	isAboveThreshold := value >= as.AlertConfig.Threshold
	if isAboveThreshold {
		if as.CurrentStartHitTime == 0 {
			as.CurrentStartHitTime = currentTime
			as.PreviousValue = value
			as.PreviousStartTime = currentTime
			return nil
		}
	}

	if as.CurrentStartHitTime != 0 {
		if as.PreviousValue >= as.AlertConfig.Threshold && isAboveThreshold {
			as.TotalHitInterval += (currentTime - as.PreviousStartTime)
		}
		as.PreviousValue = value
		as.PreviousStartTime = currentTime
	}

	if as.TotalHitInterval >= as.AlertConfig.Window.Nanoseconds() {
		// TODO: Fill in details about the alert
		as.CurrentStartHitTime = 0
		return &ThresholdAlert{}
	}

	// if isAboveThreshold {
	// 	if as.CurrentStartHitTime == 0 {
	// 		as.CurrentStartHitTime = currentTime
	// 		return nil
	// 	}

	// 	// Prune intervals to update TotalHitInterval.
	// 	as.pruneIntervals(currentTime)

	// 	// Check if we're over the threshold now.
	// 	if as.TotalHitInterval+(currentTime-as.CurrentStartHitTime) >=
	// 		as.AlertConfig.Window.Nanoseconds() {
	// 		// TODO: Generate alert here
	// 		return &ThresholdAlert{}
	// 	}

	// 	return nil
	// }

	// if as.CurrentStartHitTime != 0 {
	// 	interval := currentTime - as.CurrentStartHitTime
	// 	hitInterval := HitInterval{
	// 		Interval:  interval,
	// 		StartTime: as.CurrentStartHitTime,
	// 	}
	// 	as.TotalHitInterval += interval
	// 	as.HitIntervals = append(as.HitIntervals, hitInterval)
	// 	as.CurrentStartHitTime = 0

	// 	as.pruneIntervals(currentTime)

	// 	// Check if interval is over the window * interval threshold
	// 	if as.TotalHitInterval >= int64(float32(as.AlertConfig.Window.Nanoseconds())*as.IntervalThresholdRatio) {
	// 		// TODO: Fill in details about the alert
	// 		return &ThresholdAlert{}
	// 	}
	// }

	return nil

}

type AlertEvaluator struct {
	AlertStates map[string]*AlertState
	Threshold   int
}

func NewAlertEvaluator(alertConfigs []*AlertConfig, globalThreshold int) *AlertEvaluator {
	alertStates := map[string]*AlertState{}
	for _, alertConfig := range alertConfigs {
		alertStates[alertConfig.Metric] = &AlertState{
			AlertConfig:  alertConfig,
			HitIntervals: make([]HitInterval, 0),
		}
	}

	return &AlertEvaluator{
		AlertStates: alertStates,
		Threshold:   globalThreshold,
	}
}

func (ae *AlertEvaluator) ProcessMetric(currentTime int64, metricName string, value float32) *ThresholdAlert {
	alertState, ok := ae.AlertStates[metricName]
	if !ok {
		return nil
	}

	return alertState.ProcessValue(currentTime, value)
}
