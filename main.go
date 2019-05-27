package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/searchlight/prometheus-metrics-exporter/metrics"
	"github.com/spf13/cobra"
)

var (
	alertMetrc = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "alert_test",
		Help: "for testing alert purpose",
		ConstLabels: prometheus.Labels{
			"app": "metric-exporter",
		},
	})
)

func NewRootCmd() *cobra.Command {
	metricsConf := metrics.NewMetricsExporterConfigs()

	var rootCmd = &cobra.Command{
		Use:               "metrics-writer [command]",
		Short:             `Prometheus metrics writer`,
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := metricsConf.Validate(); err != nil {
				return err
			}

			metricsExporter, err := metrics.NewMetricsExporter(metricsConf, prometheus.NewRegistry())
			if err != nil {
				return errors.Wrap(err, "failed to create client for metrics exporter")
			}

			alertMetrc.Set(0)
			metricsExporter.Register(alertMetrc)

			stopCh := make(chan struct{})
			if err := metricsExporter.Run(stopCh); err != nil {
				return err
			}

			http.Handle("/alert", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				data := struct {
					Value int `json:"value"`
				}{}
				defer r.Body.Close()
				if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, "new value", data.Value)
				alertMetrc.Set(float64(data.Value))
			}))

			if err := http.ListenAndServe(":8080", http.DefaultServeMux); err != nil {
				close(stopCh)
				glog.Fatal(err)
			}
			<-stopCh

			return nil
		},
	}
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
	flag.CommandLine.Parse([]string{})
	metricsConf.AddFlags(rootCmd.PersistentFlags())
	return rootCmd
}

func main() {
	rootCmd := NewRootCmd()

	if err := rootCmd.Execute(); err != nil {
		glog.Fatal(err)
	}
}
