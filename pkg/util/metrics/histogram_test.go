package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"testing"
)

func TestHistogram(t *testing.T) {
	var tests = []struct {
		desc string
		HistogramOpts
		registryVersion     *Version
		expectedMetricCount int
		expectedHelp        string
	}{
		{
			desc: "Test non deprecated",
			HistogramOpts: HistogramOpts{
				Namespace: "namespace",
				Name:      "metric_test_name",
				Subsystem: "subsystem",
				Help:      "counter help",
				Buckets:   prometheus.DefBuckets,
			},
			registryVersion:     MustParseGeneric("1.15.0"),
			expectedMetricCount: 1,
			expectedHelp:        "counter help",
		},
		{
			desc: "Test deprecated",
			HistogramOpts: HistogramOpts{
				Namespace:         "namespace",
				Name:              "metric_test_name",
				Subsystem:         "subsystem",
				Help:              "counter help",
				DeprecatedVersion: MustParseGeneric("1.15.0"),
				Buckets:           prometheus.DefBuckets,
			},
			registryVersion:     MustParseGeneric("1.15.0"),
			expectedMetricCount: 1,
			expectedHelp:        "(Deprecated since 1.15.0) counter help",
		},
		{
			desc: "Test hidden",
			HistogramOpts: HistogramOpts{
				Namespace:         "namespace",
				Name:              "metric_test_name",
				Subsystem:         "subsystem",
				Help:              "counter help",
				DeprecatedVersion: MustParseGeneric("1.14.0"),
				Buckets:           prometheus.DefBuckets,
			},
			registryVersion:     MustParseGeneric("1.15.0"),
			expectedMetricCount: 0,
			expectedHelp:        "counter help",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			registry := NewKubeRegistry(test.registryVersion)
			c := NewHistogram(test.HistogramOpts)
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
			c.Observe(1)
			c.Observe(2)
			c.Observe(3)
			c.Observe(1.5)
			expected := 4
			ms, err = registry.Gather()
			if err != nil {
				t.Fatalf("Gather failed %v", err)
			}
			for _, mf := range ms {
				for _, m := range mf.GetMetric() {
					if int(m.GetHistogram().GetSampleCount()) != expected {
						t.Errorf("Got %v, want %v as the sample count", m.GetHistogram().GetSampleCount(), expected)
					}
				}
			}
		})
	}
}

func TestHistogramVec(t *testing.T) {
	var tests = []struct {
		desc string
		HistogramOpts
		labels              []string
		registryVersion     *Version
		expectedMetricCount int
		expectedHelp        string
	}{
		{
			desc: "Test non deprecated",
			HistogramOpts: HistogramOpts{
				Namespace: "namespace",
				Name:      "metric_test_name",
				Subsystem: "subsystem",
				Help:      "counter help",
				Buckets:   prometheus.DefBuckets,
			},
			labels:              []string{"label_a", "label_b"},
			registryVersion:     MustParseGeneric("1.15.0"),
			expectedMetricCount: 1,
			expectedHelp:        "counter help",
		},
		{
			desc: "Test deprecated",
			HistogramOpts: HistogramOpts{
				Namespace:         "namespace",
				Name:              "metric_test_name",
				Subsystem:         "subsystem",
				Help:              "counter help",
				DeprecatedVersion: MustParseGeneric("1.15.0"),
				Buckets:           prometheus.DefBuckets,
			},
			labels:              []string{"label_a", "label_b"},
			registryVersion:     MustParseGeneric("1.15.0"),
			expectedMetricCount: 1,
			expectedHelp:        "(Deprecated since 1.15.0) counter help",
		},
		{
			desc: "Test hidden",
			HistogramOpts: HistogramOpts{
				Namespace:         "namespace",
				Name:              "metric_test_name",
				Subsystem:         "subsystem",
				Help:              "counter help",
				DeprecatedVersion: MustParseGeneric("1.14.0"),
				Buckets:           prometheus.DefBuckets,
			},
			labels:              []string{"label_a", "label_b"},
			registryVersion:     MustParseGeneric("1.15.0"),
			expectedMetricCount: 0,
			expectedHelp:        "counter help",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			registry := NewKubeRegistry(test.registryVersion)
			c := NewHistogramVec(test.HistogramOpts, test.labels)
			registry.MustRegister(c)
			c.WithLabelValues("1", "2").Observe(1.0)
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
			c.WithLabelValues("1", "3").Observe(1.0)
			c.WithLabelValues("2", "3").Observe(1.0)
			ms, err = registry.Gather()
			if err != nil {
				t.Fatalf("Gather failed %v", err)
			}
			for _, mf := range ms {
				if len(mf.GetMetric()) != 3 {
					t.Errorf("Got %v metrics, wanted 2 as the count", len(mf.GetMetric()))
				}
				for _, m := range mf.GetMetric() {
					if m.GetHistogram().GetSampleCount() != 1 {
						t.Errorf(
							"Got %v metrics, expected histogram sample count to equal 1",
							m.GetHistogram().GetSampleCount())
					}
				}
			}
		})
	}
}
