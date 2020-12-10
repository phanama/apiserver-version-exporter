package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type Version struct {
	Major        string
	Minor        string
	GitVersion   string
	GitCommit    string
	GitTreeState string
	BuildDate    string
	GoVersion    string
	Compiler     string
	Platform     string
}

var (
	promlogConfig       = &promlog.Config{}
	logger              = promlog.New(promlogConfig)
	listenAddress       = kingpin.Flag("web.listenAddressPort", "The address:port to listen on.").Default(":9101").String()
	metricsPath         = kingpin.Flag("web.metricsPath", "Metrics expose path.").Default("/metrics").String()
	scrapePeriod        = kingpin.Flag("exporter.scrapePeriod", "The scrape period of the exporter.").Default("5").String()
	scrapeTimeout       = kingpin.Flag("exporter.scrapeTimeout", "The scrape timeout of the exporter.").Default("4").String()
	kubeVersionEndpoint = kingpin.Flag("exporter.apiserverEndpoint", "The apiserver endpoint to scrape from.").Default("https://kubernetes.default.svc.cluster.local/version").String()
	apiBuildInfo        = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "kubernetes_build_info",
		Help: "Build info of the component"},
		[]string{"major", "minor", "gitVersion", "gitTreeState", "gitCommit", "buildDate", "goVersion", "compiler", "platform"},
	)
)

func recordVersion() {
	go func() {
		for {
			version := Version{}
			err := getApiServerVersion(*kubeVersionEndpoint, &version)
			if err != nil {
				level.Error(logger).Log("msg", err)
			}

			apiBuildInfo.WithLabelValues(
				version.Major,
				version.Minor,
				version.GitVersion,
				version.GitTreeState,
				version.GitCommit,
				version.BuildDate,
				version.GoVersion,
				version.Compiler,
				version.Platform,
			).Set(1)

			period, err := strconv.Atoi(*scrapePeriod)
			if err != nil {
				level.Error(logger).Log("msg", err)
			}
			time.Sleep(time.Duration(period) * time.Second)
		}
	}()
}

func getApiServerVersion(kubeVersionEndpoint string, version interface{}) error {
	timeout, err := strconv.Atoi(*scrapeTimeout)
	if err != nil {
		return err
	}

	client := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	resp, err := client.Get(kubeVersionEndpoint)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(version)

	return err
}

func main() {

	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	level.Info(logger).Log("msg", "Starting apiserver version exporter")

	recordVersion()

	http.Handle(*metricsPath, promhttp.Handler())
	http.ListenAndServe(*listenAddress, nil)
}
