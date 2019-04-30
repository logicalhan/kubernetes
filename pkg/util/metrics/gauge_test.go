package metrics

import (
	"github.com/blang/semver"
	"testing"
)

func TestGauge(t *testing.T) {
	v115 := semver.MustParse("1.15.0")
	v114 := semver.MustParse("1.14.0")
	var tests = []struct {
		desc string
		*GaugeOpts
		registryVersion     *semver.Version
		expectedMetricCount int
		expectedHelp        string
	}{
		{
			desc: "Test non deprecated",
			GaugeOpts: &GaugeOpts{
				Namespace: "namespace",
				Name:      "metric_test_name",
				Subsystem: "subsystem",
				Help:      "gauge help",
			},
			registryVersion:     &v115,
			expectedMetricCount: 1,
			expectedHelp:        "[ALPHA] gauge help",
		},
		{
			desc: "Test deprecated",
			GaugeOpts: &GaugeOpts{
				Namespace:         "namespace",
				Name:              "metric_test_name",
				Subsystem:         "subsystem",
				Help:              "gauge help",
				DeprecatedVersion: &v115,
			},
			registryVersion:     &v115,
			expectedMetricCount: 1,
			expectedHelp:        "[ALPHA] (Deprecated since 1.15.0) gauge help",
		},
		{
			desc: "Test hidden",
			GaugeOpts: &GaugeOpts{
				Namespace:         "namespace",
				Name:              "metric_test_name",
				Subsystem:         "subsystem",
				Help:              "gauge help",
				DeprecatedVersion: &v114,
			},
			registryVersion:     &v115,
			expectedMetricCount: 0,
			expectedHelp:        "gauge help",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			registry := newKubeRegistry(*test.registryVersion)
			c := NewGauge(test.GaugeOpts)
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
			c.Set(100)
			c.Set(101)
			expected := 101
			ms, err = registry.Gather()
			if err != nil {
				t.Fatalf("Gather failed %v", err)
			}
			for _, mf := range ms {
				for _, m := range mf.GetMetric() {
					if int(m.GetGauge().GetValue()) != expected {
						t.Errorf("Got %v, wanted %v as the count", m.GetGauge().GetValue(), expected)
					}
					t.Logf("%v\n", m.GetGauge().GetValue())
				}
			}
		})
	}
}

func TestGaugeVec(t *testing.T) {
	v115 := semver.MustParse("1.15.0")
	v114 := semver.MustParse("1.14.0")
	var tests = []struct {
		desc string
		*GaugeOpts
		labels              []string
		registryVersion     *semver.Version
		expectedMetricCount int
		expectedHelp        string
	}{
		{
			desc: "Test non deprecated",
			GaugeOpts: &GaugeOpts{
				Namespace: "namespace",
				Name:      "metric_test_name",
				Subsystem: "subsystem",
				Help:      "gauge help",
			},
			labels:              []string{"label_a", "label_b"},
			registryVersion:     &v115,
			expectedMetricCount: 1,
			expectedHelp:        "[ALPHA] gauge help",
		},
		{
			desc: "Test deprecated",
			GaugeOpts: &GaugeOpts{
				Namespace:         "namespace",
				Name:              "metric_test_name",
				Subsystem:         "subsystem",
				Help:              "gauge help",
				DeprecatedVersion: &v115,
			},
			labels:              []string{"label_a", "label_b"},
			registryVersion:     &v115,
			expectedMetricCount: 1,
			expectedHelp:        "[ALPHA] (Deprecated since 1.15.0) gauge help",
		},
		{
			desc: "Test hidden",
			GaugeOpts: &GaugeOpts{
				Namespace:         "namespace",
				Name:              "metric_test_name",
				Subsystem:         "subsystem",
				Help:              "gauge help",
				DeprecatedVersion: &v114,
			},
			labels:              []string{"label_a", "label_b"},
			registryVersion:     &v115,
			expectedMetricCount: 0,
			expectedHelp:        "gauge help",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			registry := newKubeRegistry(*test.registryVersion)
			c := NewGaugeVec(test.GaugeOpts, test.labels)
			registry.MustRegister(c)
			c.WithLabelValues("1", "2").Set(1.0)
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
			c.WithLabelValues("1", "3").Set(1.0)
			c.WithLabelValues("2", "3").Set(1.0)
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
