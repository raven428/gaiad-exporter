package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type NetInfoRes struct {
	PeersNumber string `json:"n_peers"`
}
type NetInfo struct {
	Res NetInfoRes `json:"result"`
}
type StatusResSyncInfo struct {
	LatestBlockTime     string `json:"latest_block_time"`
	LatestBlockHeight   string `json:"latest_block_height"`
	EarliestBlockHeight string `json:"earliest_block_height"`
}
type StatusRes struct {
	SyncInfo StatusResSyncInfo `json:"sync_info"`
}
type Status struct {
	Res StatusRes `json:"result"`
}

var (
	NumberOfPeers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: `gaia_number_of_peers`,
			Help: `Current number of peers in gaiad`,
		},
	)
	LatestBlockTime = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: `gaia_latest_block_time`,
			Help: `Latest block time in gaiad`,
		},
	)
	NumberOfBlocks = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: `gaia_number_of_blocks`,
			Help: `Total number of blocks in gaiad`,
		},
	)
	listeningAddress = flag.String(`listen`, `:9339`, `Address on which to expose metrics.`)
	metricsEndpoint  = flag.String(`metrics`, `/metrics`, `Path under which to expose metrics.`)
)

func recordMetrics() {
	go func() {
		for {
			time.Sleep(10 * time.Second)
			spaceClient := http.Client{
				Timeout: time.Second * 2,
			}
			req, err := http.NewRequest(
				http.MethodGet,
				`http://127.0.0.1:26657/net_info`,
				nil,
			)
			if err != nil {
				fmt.Println(err)
				continue
			}
			res, getErr := spaceClient.Do(req)
			if getErr != nil {
				fmt.Println(getErr)
				continue
			}
			if res.Body != nil {
				defer res.Body.Close()
			}
			body, readErr := ioutil.ReadAll(res.Body)
			if readErr != nil {
				fmt.Println(readErr)
				continue
			}
			var netInfo NetInfo
			jsonErr := json.Unmarshal(body, &netInfo)
			if jsonErr != nil {
				fmt.Println(jsonErr)
				continue
			}
			var peersNumber float64
			peersNumber, err = strconv.ParseFloat(netInfo.Res.PeersNumber, 64)
			if err != nil {
				fmt.Println(err)
				continue
			}
			NumberOfPeers.Set(peersNumber)
		}
	}()
	go func() {
		for {
			time.Sleep(10 * time.Second)
			spaceClient := http.Client{
				Timeout: time.Second * 2,
			}
			req, err := http.NewRequest(
				http.MethodGet,
				`http://127.0.0.1:26657/status`,
				nil,
			)
			if err != nil {
				fmt.Println(err)
				continue
			}
			res, getErr := spaceClient.Do(req)
			if getErr != nil {
				fmt.Println(getErr)
				continue
			}
			if res.Body != nil {
				defer res.Body.Close()
			}
			body, readErr := ioutil.ReadAll(res.Body)
			if readErr != nil {
				fmt.Println(readErr)
				continue
			}
			var status Status
			jsonErr := json.Unmarshal(body, &status)
			if jsonErr != nil {
				fmt.Println(jsonErr)
				continue
			}
			var latestBlockHeight, earliestBlockHeight float64
			latestBlockTime, timeErr := time.Parse(
				time.RFC3339Nano,
				status.Res.SyncInfo.LatestBlockTime,
			)
			if timeErr != nil {
				fmt.Println(timeErr)
				continue
			}
			latestBlockHeight, err = strconv.ParseFloat(
				status.Res.SyncInfo.LatestBlockHeight, 64,
			)
			if err != nil {
				fmt.Println(err)
				continue
			}
			earliestBlockHeight, err = strconv.ParseFloat(
				status.Res.SyncInfo.EarliestBlockHeight, 64,
			)
			if err != nil {
				fmt.Println(err)
				continue
			}
			LatestBlockTime.Set(float64(latestBlockTime.Unix()))
			NumberOfBlocks.Set(latestBlockHeight - earliestBlockHeight)
		}
	}()
}

func main() {
	flag.Parse()
	http.Handle(*metricsEndpoint, promhttp.Handler())
	http.HandleFunc(`/`, func(w http.ResponseWriter, r *http.Request) {
		w.Write(
			[]byte(
				`<html><head><title>p2p-entry exporter</title></head><body>` +
					`<h1>p2p-entry exporter</h1><p><a href="` +
					*metricsEndpoint +
					`">Metrics</a></p></body></html>`,
			),
		)
	})
	http.HandleFunc(`/-/healthy`, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`ok`))
	})
	recordMetrics()
	fmt.Printf("listening at [%s]\n", *listeningAddress)
	fmt.Println(http.ListenAndServe(*listeningAddress, nil))
}
