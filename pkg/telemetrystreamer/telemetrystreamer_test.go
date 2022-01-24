package telemetrystreamer_test

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/grafana/grafana-starter-datasource-backend/pkg/telemetrystreamer"
)

func TestGetInstance(t *testing.T) {

	ts1 := telemetrystreamer.GetInstance(context.TODO(), "192.168.10.233:1545", "admin", "admin")
	ts2 := telemetrystreamer.GetInstance(context.TODO(), "192.168.10.233:1545", "admin", "admin")
	defer ts1.Dispose()
	defer ts2.Dispose()

	if ts1 != ts2 {
		t.Errorf("%v != %v, must was equal", ts1, ts2)
	}

}
func TestSubscribe(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())

	ts := telemetrystreamer.GetInstance(
		ctx,
		"192.168.10.233:1545",
		"admin",
		"admin",
	)

	chm := ts.SubscribeStream(telemetrystreamer.Memory)
	chc := ts.SubscribeStream(telemetrystreamer.Connections)

	timeout := time.After(time.Millisecond * 20000)

loop:
	for {
		select {
		case <-timeout:
			cancel()
			break loop
		case metric := <-chm:
			log.Printf("memory:%v", metric)
		case metric := <-chc:
			log.Printf("connection:%v", metric)
		}
	}
}

func TestDescribe(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())

	ts := telemetrystreamer.GetInstance(
		ctx,
		"192.168.10.233:1545",
		"admin",
		"admin",
	)

	ch := ts.SubscribeStream(telemetrystreamer.Memory)

	timetodesc := time.After(time.Millisecond * 10000)

loop:
	for {
		select {
		case <-timetodesc:
			ts.DescribeStream(ch)
		case metric, ok := <-ch:
			if !ok {
				cancel()
				break loop
			}
			log.Printf("%v", metric)
		}
	}
}
