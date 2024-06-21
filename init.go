package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

func init() {
	http.DefaultClient.Timeout = time.Second * 15

	// log.SetFlags(log.Lshortfile | log.Ltime)
	log.SetFlags(log.Ltime | log.Lshortfile)
	log.Println("starting")

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	return
	th := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
		// ReplaceAttr: nil,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.SourceKey:
				source, _ := a.Value.Any().(*slog.Source)
				if source != nil {
					source.File = filepath.Base(source.File)
				}
			case slog.TimeKey:
				return slog.Attr{
					Key:   "time",
					Value: slog.TimeValue(a.Value.Time()),
				}
			}
			return a
		},
	})
	slog.SetDefault(slog.New(th))
}
