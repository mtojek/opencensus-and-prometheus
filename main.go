package main

import (
	"context"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"

	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

const (
	html = `<!doctype html>
<html>
<body>
	<a href="/metrics">metrics</a>
</body>
</html>`
)

func main() {
	ctx := context.Background()

	exporter, err := prometheus.NewExporter(prometheus.Options{})
	if err != nil {
		log.Fatal(err)
	}
	view.RegisterExporter(exporter)

	videoCount := stats.Int64("my.org/measures/video_count", "number of processed videos", "")
	viewCount := &view.View{
		"video_count",
		"number of videos processed over time",
		nil,
		videoCount,
		view.Count(),
	}

	videoSize := stats.Int64("my.org/measures/video_size_cum", "size of processed video", "MBy")
	viewSize := &view.View{
		"video_cum",
		"processed video size over time",
		nil,
		videoSize,
		view.Distribution(0, 1<<16, 1<<32),
	}

	if err = view.Register(viewCount, viewSize); err != nil {
		log.Fatalf("Cannot register the view: %v", err)
	}

	view.SetReportingPeriod(1 * time.Second)

	go func() {
		for {
			stats.Record(ctx, videoCount.M(1))
			stats.Record(ctx, videoSize.M(rand.Int63()))
			<-time.After(time.Millisecond * time.Duration(1+rand.Intn(400)))
		}
	}()

	addr := ":9999"
	log.Printf("Serving at %s", addr)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.New("foo").Parse(html)
		if err != nil {
			log.Fatalf("Cannot parse template: %v", err)
		}
		t.Execute(w, "")
	})
	http.Handle("/metrics", exporter)
	log.Fatal(http.ListenAndServe(addr, nil))
}
