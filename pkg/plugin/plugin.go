package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-plugin-sdk-go/live"
	"github.com/grafana/grafana-starter-datasource-backend/pkg/telemetrystreamer"
)

// Make sure Datasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler, backend.StreamHandler interfaces. Plugin should not
// implement all these interfaces - only those which are required for a particular task.
// For example if plugin does not need streaming functionality then you are free to remove
// methods that implement backend.StreamHandler. Implementing instancemgmt.InstanceDisposer
// is useful to clean up resources used by previous datasource instance when a new datasource
// instance created upon datasource settings changed.
var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.StreamHandler         = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
)

type datasourceSettings struct {
	Host string `json:"host"`
	Port string `json:"port"`
	User string `json:"user"`
}

func (ds *datasourceSettings) Url() string {
	return fmt.Sprintf("%s:%s", ds.Host, ds.Port)
}

// NewDatasource creates a new datasource instance.
func NewDatasource(ds backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	log.DefaultLogger.Info("NewDatasource.called", "DataSourceInstanceSettings", ds)

	var dds datasourceSettings

	err := json.Unmarshal(ds.JSONData, &dds)
	if err != nil {
		log.DefaultLogger.Info("NewDatasource.UnmarshalJSON", "error", err)
	}

	log.DefaultLogger.Info("NewDatasource.UnmarshalJSON", "data", dds)
	return &Datasource{
		ts: telemetrystreamer.GetInstance(
			context.TODO(),
			dds.Url(),
			dds.User,
			ds.DecryptedSecureJSONData["pass"]),
		streams: make(map[string]chan []telemetrystreamer.Metric),
	}, nil
}

// Datasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type Datasource struct {
	ts      telemetrystreamer.TelemetryStreamer
	streams map[string]chan []telemetrystreamer.Metric
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewDatasource factory function.
func (d *Datasource) Dispose() {
	log.DefaultLogger.Info("Dispose.called")
	for _, stream := range d.streams {
		d.ts.DescribeStream(stream)
	}
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Info("QueryData.called", "request")

	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type queryModel struct {
	Channel string `json:"channel"`
}

func (d *Datasource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	response := backend.DataResponse{}

	// Unmarshal the JSON into our queryModel.
	var qm queryModel

	response.Error = json.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		return response
	}
	log.DefaultLogger.Info("query.called", "queryModel", qm)
	// create data frame response.
	frame := data.NewFrame("response")

	// If query called with streaming on then return a channel
	// to subscribe on a client-side and consume updates from a plugin.
	// Feel free to remove this if you don't need streaming for your datasource.
	channel := live.Channel{
		Scope:     live.ScopeDatasource,
		Namespace: pCtx.DataSourceInstanceSettings.UID,
		Path:      qm.Channel,
	}
	frame.SetMeta(&data.FrameMeta{Channel: channel.String()})

	// add the frames to the response.
	response.Frames = append(response.Frames, frame)

	return response
}

// SubscribeStream is called when a client wants to connect to a stream. This callback
// allows sending the first message.
func (d *Datasource) SubscribeStream(_ context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	log.DefaultLogger.Info("SubscribeStream.called", "request", req.Path)

	status := backend.SubscribeStreamStatusPermissionDenied

	switch req.Path {
	case "memory":
		if _, exist := d.streams[req.Path]; !exist {
			d.streams[req.Path] = d.ts.SubscribeStream(telemetrystreamer.Memory)
		}
		status = backend.SubscribeStreamStatusOK
	case "connections":
		if _, exist := d.streams[req.Path]; !exist {
			d.streams[req.Path] = d.ts.SubscribeStream(telemetrystreamer.Connections)
		}
		status = backend.SubscribeStreamStatusOK
	case "avg_threads":
		if _, exist := d.streams[req.Path]; !exist {
			d.streams[req.Path] = d.ts.SubscribeStream(telemetrystreamer.AvgThreads)
		}
		status = backend.SubscribeStreamStatusOK
	case "avg_call_time":
		if _, exist := d.streams[req.Path]; !exist {
			d.streams[req.Path] = d.ts.SubscribeStream(telemetrystreamer.AvgCallTime)
		}
		status = backend.SubscribeStreamStatusOK
	case "selection_size":
		if _, exist := d.streams[req.Path]; !exist {
			d.streams[req.Path] = d.ts.SubscribeStream(telemetrystreamer.SelectionSize)
		}
		status = backend.SubscribeStreamStatusOK
	case "avg_db_call_time":
		if _, exist := d.streams[req.Path]; !exist {
			d.streams[req.Path] = d.ts.SubscribeStream(telemetrystreamer.AvgDbCallTime)
		}
		status = backend.SubscribeStreamStatusOK
	case "avg_server_call_time":
		if _, exist := d.streams[req.Path]; !exist {
			d.streams[req.Path] = d.ts.SubscribeStream(telemetrystreamer.AvgServerCallTime)
		}
		status = backend.SubscribeStreamStatusOK
	default:

	}
	return &backend.SubscribeStreamResponse{
		Status: status,
	}, nil
}

// RunStream is called once for any open channel.  Results are shared with everyone
// subscribed to the same channel.
func (d *Datasource) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {
	log.DefaultLogger.Info("RunStream.called")

	// Create the same data frame as for query data.
	frame := data.NewFrame("response")

	// Stream data frames periodically till stream closed by Grafana.
	for {
		select {
		case <-ctx.Done():
			d.ts.DescribeStream(d.streams[req.Path])
			delete(d.streams, req.Path)
			log.DefaultLogger.Info("RunStream.ctx.Done", "path", req.Path, d.ts.Info())
			return nil
		case metrics := <-d.streams[req.Path]:
			// log.DefaultLogger.Info("RunStream", "metrics <- d.stream", metrics)
			// Send new data periodically.

			frame.Fields = []*data.Field{
				data.NewField("time",
					nil,
					[]time.Time{time.Now()},
				),
			}
			for _, metric := range metrics {

				frame.Fields = append(frame.Fields,
					data.NewField(metric.Name,
						metric.Labels,
						[]float64{metric.Value}),
				)

			}
			err := sender.SendFrame(frame, data.IncludeAll)
			if err != nil {
				log.DefaultLogger.Error("RunStream.SendFrame", "error", err)
				continue
			}
		}
	}

}

// PublishStream is called when a client sends a message to the stream.
func (d *Datasource) PublishStream(_ context.Context, req *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	log.DefaultLogger.Info("PublishStream.called")

	// Do not allow publishing at all.
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}

func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {

	result := backend.CheckHealthResult{
		Status:      1,
		Message:     "CheckHealth: not implemented",
		JSONDetails: nil,
	}

	return &result, nil
}
