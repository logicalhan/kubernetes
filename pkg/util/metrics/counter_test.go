package metrics

import (
	"github.com/blang/semver"
	"testing"
)

func TestCounter(t *testing.T) {
	v115 := semver.MustParse("1.15.0")
	v114 := semver.MustParse("1.14.0")
	var tests = []struct {
		desc string
		CounterOpts
		registryVersion     *semver.Version
		expectedMetricCount int
		expectedHelp        string
	}{
		{
			desc: "Test non deprecated",
			CounterOpts: CounterOpts{
				Namespace: "namespace",
				Name:      "metric_test_name",
				Subsystem: "subsystem",
				Help:      "counter help",
			},
			registryVersion:     &v115,
			expectedMetricCount: 1,
			expectedHelp:        "counter help",
		},
		{
			desc: "Test deprecated",
			CounterOpts: CounterOpts{
				Namespace:         "namespace",
				Name:              "metric_test_name",
				Subsystem:         "subsystem",
				Help:              "counter help",
				DeprecatedVersion: &v115,
			},
			registryVersion:     &v115,
			expectedMetricCount: 1,
			expectedHelp:        "(Deprecated since 1.15.0) counter help",
		},
		{
			desc: "Test hidden",
			CounterOpts: CounterOpts{
				Namespace:         "namespace",
				Name:              "metric_test_name",
				Subsystem:         "subsystem",
				Help:              "counter help",
				DeprecatedVersion: &v114,
			},
			registryVersion:     &v115,
			expectedMetricCount: 0,
			expectedHelp:        "counter help",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			registry := NewKubeRegistry(*test.registryVersion)
			c := NewCounter(test.CounterOpts)
			registry.MustRegister(c)

			ms, err := registry.Gather()
			if len(ms) != test.expectedMetricCount {
				t.Errorf("Got %v metrics, Want: %v metrics", len(ms), test.expectedMetricCount)
			}
			if err != nil {
				t.Fatalf("Gather failed %v", err)
			}
			for _, metric := range ms {
				if metric.GetHelp() != test.expectedHelp {
					t.Errorf("Got %s as help message, want %s", metric.GetHelp(), test.expectedHelp)
				}
			}

			// let's increment the counter and verify that the metric still works
			c.Inc()
			c.Inc()
			ms, err = registry.Gather()
			if err != nil {
				t.Fatalf("Gather failed %v", err)
			}
			for _, mf := range ms {
				for _, m := range mf.GetMetric() {
					if m.GetCounter().GetValue() != 2 {
						t.Errorf("Got %v, wanted 2 as the count", m.GetCounter().GetValue())
					}
					t.Logf("%v\n", m.GetCounter().GetValue())
				}
			}
		})
	}
}

func TestCounterVec(t *testing.T) {
	v115 := semver.MustParse("1.15.0")
	v114 := semver.MustParse("1.14.0")
	var tests = []struct {
		desc string
		CounterOpts
		labels              []string
		registryVersion     *semver.Version
		expectedMetricCount int
		expectedHelp        string
	}{
		{
			desc: "Test non deprecated",
			CounterOpts: CounterOpts{
				Namespace: "namespace",
				Name:      "metric_test_name",
				Subsystem: "subsystem",
				Help:      "counter help",
			},
			labels:              []string{"label_a", "label_b"},
			registryVersion:     &v115,
			expectedMetricCount: 1,
			expectedHelp:        "counter help",
		},
		{
			desc: "Test deprecated",
			CounterOpts: CounterOpts{
				Namespace:         "namespace",
				Name:              "metric_test_name",
				Subsystem:         "subsystem",
				Help:              "counter help",
				DeprecatedVersion: &v115,
			},
			labels:              []string{"label_a", "label_b"},
			registryVersion:     &v115,
			expectedMetricCount: 1,
			expectedHelp:        "(Deprecated since 1.15.0) counter help",
		},
		{
			desc: "Test hidden",
			CounterOpts: CounterOpts{
				Namespace:         "namespace",
				Name:              "metric_test_name",
				Subsystem:         "subsystem",
				Help:              "counter help",
				DeprecatedVersion: &v114,
			},
			labels:              []string{"label_a", "label_b"},
			registryVersion:     &v115,
			expectedMetricCount: 0,
			expectedHelp:        "counter help",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			registry := NewKubeRegistry(*test.registryVersion)
			c := NewCounterVec(test.CounterOpts, test.labels)
			registry.MustRegister(c)
			c.WithLabelValues("1", "2").Inc()
			ms, err := registry.Gather()

			if len(ms) != test.expectedMetricCount {
				t.Errorf("Got %v metrics, Want: %v metrics", len(ms), test.expectedMetricCount)
			}
			if err != nil {
				t.Fatalf("Gather failed %v", err)
			}
			for _, metric := range ms {
				if metric.GetHelp() != test.expectedHelp {
					t.Errorf("Got %s as help message, want %s", metric.GetHelp(), test.expectedHelp)
				}
			}

			// let's increment the counter and verify that the metric still works
			c.WithLabelValues("1", "3").Inc()
			c.WithLabelValues("2", "3").Inc()
			ms, err = registry.Gather()
			if err != nil {
				t.Fatalf("Gather failed %v", err)
			}
			for _, mf := range ms {
				if len(mf.GetMetric()) != 3 {
					t.Errorf("Got %v metrics, wanted 2 as the count", len(mf.GetMetric()))
				}
			}
		})
	}
}