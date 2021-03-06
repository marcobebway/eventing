/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package adapter

import (
	"context"
	"os"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"go.opencensus.io/stats/view"
	"k8s.io/apimachinery/pkg/types"
	_ "knative.dev/pkg/client/injection/kube/client/fake"
	"knative.dev/pkg/leaderelection"
	"knative.dev/pkg/metrics"
	"knative.dev/pkg/reconciler"
	_ "knative.dev/pkg/system/testing"
)

type myAdapter struct{}

func TestMainWithNothing(t *testing.T) {
	os.Setenv("K_SINK", "http://sink")
	os.Setenv("NAMESPACE", "ns")
	os.Setenv("K_METRICS_CONFIG", "error config")
	os.Setenv("K_LOGGING_CONFIG", "error config")
	os.Setenv("MODE", "mymode")

	defer func() {
		os.Unsetenv("K_SINK")
		os.Unsetenv("NAMESPACE")
		os.Unsetenv("K_METRICS_CONFIG")
		os.Unsetenv("K_LOGGING_CONFIG")
		os.Unsetenv("MODE")
	}()

	Main("mycomponent",
		func() EnvConfigAccessor { return &myEnvConfig{} },
		func(ctx context.Context, processed EnvConfigAccessor, client cloudevents.Client) Adapter {
			env := processed.(*myEnvConfig)

			if env.Mode != "mymode" {
				t.Error("Expected mode mymode, got:", env.Mode)
			}

			if env.Sink != "http://sink" {
				t.Error("Expected sinkURI http://sink, got:", env.Sink)
			}

			if leaderelection.HasLeaderElection(ctx) {
				t.Error("Expected no leader election, but got leader election")
			}
			return &myAdapter{}
		})

	defer view.Unregister(metrics.NewMemStatsAll().DefaultViews()...)
}

func TestMainWithInformerNoLeaderElection(t *testing.T) {
	os.Setenv("K_SINK", "http://sink")
	os.Setenv("NAMESPACE", "ns")
	os.Setenv("K_METRICS_CONFIG", "error config")
	os.Setenv("K_LOGGING_CONFIG", "error config")
	os.Setenv("MODE", "mymode")

	defer func() {
		os.Unsetenv("K_SINK")
		os.Unsetenv("NAMESPACE")
		os.Unsetenv("K_METRICS_CONFIG")
		os.Unsetenv("K_LOGGING_CONFIG")
		os.Unsetenv("MODE")
	}()

	ctx, cancel := context.WithCancel(context.TODO())
	env := ConstructEnvOrDie(func() EnvConfigAccessor { return &myEnvConfig{} })
	MainWithInformers(ctx,
		"mycomponent",
		env,
		func(ctx context.Context, processed EnvConfigAccessor, client cloudevents.Client) Adapter {
			env := processed.(*myEnvConfig)

			if env.Mode != "mymode" {
				t.Error("Expected mode mymode, got:", env.Mode)
			}

			if env.Sink != "http://sink" {
				t.Error("Expected sinkURI http://sink, got:", env.Sink)
			}

			if leaderelection.HasLeaderElection(ctx) {
				t.Error("Expected no leader election, but got leader election")
			}
			return &myAdapter{}
		})

	cancel()

	defer view.Unregister(metrics.NewMemStatsAll().DefaultViews()...)
}

func TestMain_MetricsConfig(t *testing.T) {
	m := &metrics.ExporterOptions{
		Domain:         "example.com",
		Component:      "foo",
		PrometheusPort: 9021,
		PrometheusHost: "prom.example.com",
		ConfigMap: map[string]string{
			"profiling.enable": "true",
			"foo":              "bar",
		},
	}
	metricsJson, _ := metrics.OptionsToJSON(m)

	os.Setenv("K_SINK", "http://sink")
	os.Setenv("NAMESPACE", "ns")
	os.Setenv("K_METRICS_CONFIG", metricsJson)
	os.Setenv("K_LOGGING_CONFIG", "error config")
	os.Setenv("MODE", "mymode")

	defer func() {
		os.Unsetenv("K_SINK")
		os.Unsetenv("NAMESPACE")
		os.Unsetenv("K_METRICS_CONFIG")
		os.Unsetenv("K_LOGGING_CONFIG")
		os.Unsetenv("MODE")
	}()

	ctx, cancel := context.WithCancel(context.TODO())
	env := ConstructEnvOrDie(func() EnvConfigAccessor { return &myEnvConfig{} })
	MainWithInformers(ctx,
		"mycomponent",
		env,
		func(ctx context.Context, processed EnvConfigAccessor, client cloudevents.Client) Adapter {
			env := processed.(*myEnvConfig)

			if env.Mode != "mymode" {
				t.Error("Expected mode mymode, got:", env.Mode)
			}

			if env.Sink != "http://sink" {
				t.Error("Expected sinkURI http://sink, got:", env.Sink)
			}

			if leaderelection.HasLeaderElection(ctx) {
				t.Error("Expected no leader election, but got leader election")
			}
			return &myAdapter{}
		})

	cancel()

	defer view.Unregister(metrics.NewMemStatsAll().DefaultViews()...)
}

func TestStartInformers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	StartInformers(ctx, nil)
	cancel()
}

func TestWithInjectorEnabled(t *testing.T) {
	_ = WithInjectorEnabled(context.TODO())
}

func TestConstructEnvOrDie(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected to die")
			}
			cancel()
		}()
		ConstructEnvOrDie(func() EnvConfigAccessor {
			return nil
		})
	}()
	<-ctx.Done()
}

func (m *myAdapter) Reconcile(ctx context.Context, key string) error {
	return nil
}

func (m *myAdapter) Start(_ context.Context) error {
	return nil
}

func (m *myAdapter) Promote(b reconciler.Bucket, enq func(reconciler.Bucket, types.NamespacedName)) error {
	return nil
}

func (m *myAdapter) Demote(reconciler.Bucket) {}
