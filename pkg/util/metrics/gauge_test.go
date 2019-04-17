package metrics

import "testing"

func TestGauge(t *testing.T) {
	var tests = []struct {
		desc string
		GaugeOpts
		registryVersion     *Version
		expectedMetricCount int
		expectedHelp        string
	}{
		{
			desc: "Test non deprecated",
			GaugeOpts: GaugeOpts{
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
			GaugeOpts: GaugeOpts{
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
			GaugeOpts: GaugeOpts{
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
			c.Inc()
			c.Inc()
			c.Dec()
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
	var tests = []struct {
		desc string
		GaugeOpts
		labels              []string
		registryVersion     *Version
		expectedMetricCount int
		expectedHelp        string
	}{
		{
			desc: "Test non deprecated",
			GaugeOpts: GaugeOpts{
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
			GaugeOpts: GaugeOpts{
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
			GaugeOpts: GaugeOpts{
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
			c := NewGaugeVec(test.GaugeOpts, test.labels)
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
