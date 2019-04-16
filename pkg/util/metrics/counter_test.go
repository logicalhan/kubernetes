package metrics

import "testing"

func TestCounter(t *testing.T) {
	var tests = []struct {
		desc string
		CounterOpts
		registryVersion     *Version
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
			registryVersion:     MustParseGeneric("1.15.0"),
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
				DeprecatedVersion: MustParseGeneric("1.15.0"),
			},
			registryVersion:     MustParseGeneric("1.15.0"),
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
