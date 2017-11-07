package nodeanalyzer

import (
	"testing"
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	INTERVAL       = 10
	THRESHOLD_TYPE = "active_percentage"
)

func (p *NodeAnalyzer) getTestMetrics(testData []float32) []plugin.Metric {
	metrics := []plugin.Metric{}
	currentStartHitTime := time.Now()
	for i, data := range testData {
		currentTime := currentStartHitTime.Add(time.Second * time.Duration(i*INTERVAL))
		metrics = append(metrics, plugin.Metric{
			Namespace: plugin.Namespace{
				plugin.NamespaceElement{
					Value:       "intel",
					Description: "",
					Name:        "",
				},
				plugin.NamespaceElement{
					Value:       "procfs",
					Description: "",
					Name:        "",
				},
				plugin.NamespaceElement{
					Value:       "cpu",
					Description: "",
					Name:        "",
				},
				plugin.NamespaceElement{
					Value:       "all",
					Description: "ID of CPU ('all' for aggregate)",
					Name:        "cpuID",
				},
				plugin.NamespaceElement{
					Value:       THRESHOLD_TYPE,
					Description: "",
					Name:        "",
				},
			},
			Version: 6,
			Config:  plugin.Config{},
			Data:    data,
			Tags: map[string]string{
				"deploymentId":      "",
				"nodename":          "minikube",
				"plugin_running_on": "snap-goddd1-6dd5d7d847-sdqlx",
			},
			Timestamp:   currentTime,
			Unit:        "",
			Description: "",
		})
	}

	return metrics
}

func (p *NodeAnalyzer) getTestConfig() plugin.Config {
	cfg := plugin.Config{}
	cfg["metric"] = "/intel/procfs/cpu"
	cfg["window"] = "30s"
	cfg["threshold"] = float32(50)
	cfg["type"] = "active_percentage"
	cfg["sampleInterval"] = "5s"
	return cfg
}

func TestProcessor(t *testing.T) {
	Convey("Create Node Analyzer", t, func() {
		nodeAnalyzer := NewAnalyzer()
		Convey("So node Analyzer should not be nil", func() {
			So(nodeAnalyzer, ShouldNotBeNil)
		})
		Convey("nodeAnalyzer.GetConfigPolicy() should return a config policy", func() {
			configPolicy, err := nodeAnalyzer.GetConfigPolicy()
			Convey("So config policy should not be nil", func() {
				So(err, ShouldBeNil)
				So(configPolicy, ShouldNotBeNil)
				t.Log(configPolicy)
			})
			Convey("So config policy should be a policy.ConfigPolicy", func() {
				So(configPolicy, ShouldHaveSameTypeAs, plugin.ConfigPolicy{})
			})
		})
	})

	Convey("Test parsing metrics", t, func() {
		nodeAnalyzer := &NodeAnalyzer{}
		Convey("Node Analyzer metrics should succesfully parse test metrics", func() {
			testData1 := []float32{50, 60, 0, 70, 80, 0, 90, 100}
			metrics := nodeAnalyzer.getTestMetrics(testData1)
			cfg := nodeAnalyzer.getTestConfig()

			processMetrics, err := nodeAnalyzer.Process(metrics, cfg)
			So(err, ShouldBeNil)
			So(len(processMetrics), ShouldEqual, 1)
		})
	})
}
