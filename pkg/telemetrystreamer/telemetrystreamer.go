package telemetrystreamer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	rasclient "github.com/khorevaa/ras-client"
	"github.com/khorevaa/ras-client/serialize"
)

var (
	instance     *telemetryStreamer = nil
	once                            = sync.Once{}
	deffrequency time.Duration      = 50
)

type MetricType int32

const (
	Memory MetricType = iota
	Connections
	AvgThreads
	AvgCallTime
	SelectionSize
	AvgDbCallTime
	AvgServerCallTime
)

type Metric struct {
	Name   string
	Labels map[string]string
	Value  float64
}

type TelemetryStreamer interface {
	Dispose()
	SubscribeStream(mt MetricType) chan []Metric
	DescribeStream(ch chan []Metric)
	Info() string
}

type telemetryStreamer struct {
	ras               rasclient.Api
	channels          map[chan []Metric]MetricType
	frequencyMillisec time.Duration
	mu                sync.Mutex
}

func GetInstance(ctx context.Context, url, usr, pass string) TelemetryStreamer {

	once.Do(func() {
		instance = &telemetryStreamer{
			ras:               rasclient.NewClient(url),
			channels:          make(map[chan []Metric]MetricType),
			frequencyMillisec: deffrequency,
		}
		instance.ras.AuthenticateAgent(usr, pass)
		instance.RunStream(ctx)

	})

	return instance
}

func (ts *telemetryStreamer) Dispose() {
	ts.ras.Close()
}

func (ts *telemetryStreamer) SubscribeStream(mt MetricType) chan []Metric {
	ch := make(chan []Metric)
	ts.mu.Lock()
	ts.channels[ch] = mt
	ts.mu.Unlock()
	return ch
}

func (ts *telemetryStreamer) DescribeStream(ch chan []Metric) {
	ts.mu.Lock()
	delete(ts.channels, ch)
	close(ch)
	ts.mu.Unlock()
}

func (ts *telemetryStreamer) RunStream(ctx context.Context) {
	go ts.Stream(ctx)
}

func (ts *telemetryStreamer) Stream(ctx context.Context) {
	defer ts.Dispose()
	defer func() {
		_ = recover()
	}()

	сlustersInfo, err := ts.ras.GetClusters(ctx)
	if err != nil {
		log.DefaultLogger.Error("TelemetryStreamer:", "error", err)
	}

loop:
	for {
		select {
		case <-ctx.Done():
			log.DefaultLogger.Info("TelemetryStreamer: parrent ctx done")
			break loop
		case <-time.After(time.Millisecond * ts.frequencyMillisec):

			// if no subscribes skip
			if len(ts.channels) == 0 {
				continue
			}

			var (
				metrics map[MetricType][]Metric = make(map[MetricType][]Metric)
			)

			for _, clInf := range сlustersInfo {
				workingProcessesInfo, err := ts.ras.GetWorkingProcesses(ctx, clInf.UUID)
				if err != nil {
					log.DefaultLogger.Error("TelemetryStreamer: GetWorkingProcesses", "error", err)
				}

				workingProcessesInfo.Each(func(wpInf *serialize.ProcessInfo) {

					labels := map[string]string{
						"cluster": string(clInf.Name),
						"pid":     string(wpInf.Pid),
						"port":    fmt.Sprintf("%v", wpInf.Port),
					}

					metrics[Memory] = append(metrics[Memory],
						Metric{
							"memory",
							labels,
							float64(wpInf.MemorySize),
						},
					)

					metrics[Connections] = append(metrics[Connections],
						Metric{
							"connections",
							labels,
							float64(wpInf.Connections),
						},
					)

					metrics[AvgThreads] = append(metrics[AvgThreads],
						Metric{
							"avg_threads",
							labels,
							float64(wpInf.AvgThreads),
						},
					)

					metrics[AvgCallTime] = append(metrics[AvgCallTime],
						Metric{
							"avg_call_time",
							labels,
							float64(wpInf.AvgCallTime),
						},
					)

					metrics[SelectionSize] = append(metrics[SelectionSize],
						Metric{
							"selection_size",
							labels,
							float64(wpInf.SelectionSize),
						},
					)

					metrics[AvgDbCallTime] = append(metrics[AvgDbCallTime],
						Metric{
							"avg_db_call_time",
							labels,
							float64(wpInf.AvgDbCallTime),
						},
					)

					metrics[AvgServerCallTime] = append(metrics[AvgServerCallTime],
						Metric{
							"avg_server_call_time",
							labels,
							float64(wpInf.AvgServerCallTime),
						},
					)

				})

				_ = ts.SendMetric(metrics)

			}
		}
	}
}

func (ts *telemetryStreamer) SendMetric(metrics map[MetricType][]Metric) error {

	for ch, mt := range ts.channels {
		ch <- metrics[mt]
	}

	return nil
}

func (ts *telemetryStreamer) Info() string {
	return fmt.Sprintf("%#v", ts.channels)
}
