package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/andrebq/exiftool2json/exif"
	"github.com/andrebq/exiftool2json/internal/api"
)

func main() {
	port := flag.Int("port", 8080, "Port to listen for incomming requests")
	addr := flag.String("addr", "0.0.0.0", "Address to listen for incoming requests")

	flag.Parse()

	tool, err := exif.Open(exif.AnyVersion)
	var notFound exif.ErrToolNotFound
	if errors.As(err, &notFound) {
		log.Fatalf("exiftool invalid or not found: %v", notFound)
	}
	_ = tool

	tagsAPI := api.NewAPI(tool)

	server := http.Server{
		Addr: fmt.Sprintf("%v:%v", *addr, *port),

		// adjust the timeouts for better control over slow clients
		// prevents tcp slow loris attacks
		ReadHeaderTimeout: time.Hour,
		ReadTimeout:       time.Hour,
		WriteTimeout:      time.Hour,
		IdleTimeout:       time.Hour,

		// 2 MB Headers, to coupe with absurdly long JWT tokens
		MaxHeaderBytes: 2 * 1000 * 1000,

		// 10 MB Body, to deal with huge POST/PUT requests
		Handler: tagsAPI,
	}

	listenAndServe := make(chan struct{})
	listenErr := make(chan error, 1)
	go func() {
		defer close(listenAndServe)
		log.Printf("Starting HTTP server at: %v", server.Addr)
		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			log.Print("Graceful shutdown received by HTTP Server")
		} else if err != nil {
			log.Printf("Error starting server: %v", err)
			listenErr <- err
			close(listenErr)
		}
	}()

	shutdownComplete := make(chan error, 1)
	go func() {
		defer close(shutdownComplete)
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		select {
		case <-sig:
			signal.Stop(sig)
			log.Printf("Interrup signal caught... Starting graceful shutdown")
		case err := <-listenErr:
			shutdownComplete <- err
			return
		}

		// let the server take as much time as it wants
		// since this is the main process
		err = server.Shutdown(context.Background())
		if err != nil {
			shutdownComplete <- err
		}
	}()

	<-listenAndServe

	select {
	case <-time.After(time.Minute):
		log.Print("Unable to obtain shutdown complete after one minute... will exit anyway")
	case err := <-shutdownComplete:
		if err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
		log.Print("Graceful shutdown completed")
	}
}
