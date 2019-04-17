package metrics

import (
	"testing"
)

func TestSummary(t *testing.T) {
	var tests = []struct {
		desc string
		SummaryOpts
		registryVersion     *Version
		expectedMetricCount int
		expectedHelp        string
	}{
		{
			desc: "Test non deprecated",
			SummaryOpts: SummaryOpts{
				Namespace: "namespace",
				Name:      "metric_test_name",
				Subsystem: "subsystem",
				Help:      "counter help",
			},
			registryVersion:     MustParseGeneric("1.15.0"),
			expectedMetricCount: 1,
			expectedHelp:        "counter help",
		},
		{
			desc: "Test deprecated",
			SummaryOpts: SummaryOpts{
				Namespace:         "namespace",
				Name:              "metric_test_name",
				Subsystem:         "subsystem",
				Help:              "counter help",
				DeprecatedVersion: MustParseGeneric("1.15.0"),
			},
			registryVersion:     MustParseGeneric("1.15.0"),
			expectedMetricCount: 1,
			expectedHelp:        "(Deprecated since 1.15.0) counter help",
		},
		{
			desc: "Test hidden",
			SummaryOpts: SummaryOpts{
				Namespace:         "namespace",
				Name:              "metric_test_name",
				Subsystem:         "subsystem",
				Help:              "counter help",
				DeprecatedVersion: MustParseGeneric("1.14.0"),
			},
			registryVersion:     MustParseGeneric("1.15.0"),
			expectedMetricCount: 0,
			expectedHelp:        "counter help",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			registry := NewKubeRegistry(test.registryVersion)
			c := NewSummary(test.SummaryOpts)
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
					if int(m.GetSummary().GetSampleCount()) != expected {
						t.Errorf("Got %v, want %v as the sample count", m.GetHistogram().GetSampleCount(), expected)
					}
				}
			}
		})
	}
}

func TestSummaryVec(t *testing.T) {
	var tests = []struct {
		desc string
		SummaryOpts
		labels              []string
		registryVersion     *Version
		expectedMetricCount int
		expectedHelp        string
	}{
		{
			desc: "Test non deprecated",
			SummaryOpts: SummaryOpts{
				Namespace: "namespace",
				Name:      "metric_test_name",
				Subsystem: "subsystem",
				Help:      "counter help",
			},
			labels:              []string{"label_a", "label_b"},
			registryVersion:     MustParseGeneric("1.15.0"),
			expectedMetricCount: 1,
			expectedHelp:        "counter help",
		},
		{
			desc: "Test deprecated",
			SummaryOpts: SummaryOpts{
				Namespace:         "namespace",
				Name:              "metric_test_name",
				Subsystem:         "subsystem",
				Help:              "counter help",
				DeprecatedVersion: MustParseGeneric("1.15.0"),
			},
			labels:              []string{"label_a", "label_b"},
			registryVersion:     MustParseGeneric("1.15.0"),
			expectedMetricCount: 1,
			expectedHelp:        "(Deprecated since 1.15.0) counter help",
		},
		{
			desc: "Test hidden",
			SummaryOpts: SummaryOpts{
				Namespace:         "namespace",
				Name:              "metric_test_name",
				Subsystem:         "subsystem",
				Help:              "counter help",
				DeprecatedVersion: MustParseGeneric("1.14.0"),
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
			c := NewSummaryVec(test.SummaryOpts, test.labels)
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
					if m.GetSummary().GetSampleCount() != 1 {
						t.Errorf(
							"Got %v metrics, wanted 2 as the summary sample count",
							m.GetSummary().GetSampleCount())
					}
				}
			}
		})
	}
}
