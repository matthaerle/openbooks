package server

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/evan-buss/openbooks/core"

	"github.com/rakyll/statik/fs"

	// Load the static SPA content
	_ "github.com/evan-buss/openbooks/server/statik"
)

var config core.Config
var numConnections *int32 = new(int32)

// Start instantiates the web server and opens the browser
func Start(conf core.Config) {
	config = conf

	hub := newHub()
	go hub.run()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		// Close the shutdown channel. Triggering all reader/writer WS handlers to close.
		close(hub.shutdown)
		time.Sleep(time.Second)
		os.Exit(1)
	}()

	staticFs, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", http.FileServer(staticFs))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	if config.OpenBrowser {
		openbrowser("http://127.0.0.1:" + config.Port + "/")
	}

	log.Fatal(http.ListenAndServe(":"+config.Port, nil))
}
