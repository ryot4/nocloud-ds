package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	Version              = "0.2.0"
	DefaultListenAddress = "0.0.0.0:8000"
	DefaultDataSourceDir = "."
	ListenAddressEnv     = "NOCLOUD_DS_LISTEN_ADDRESS"
	DataSourceDirEnv     = "NOCLOUD_DS_DIR"
)

func main() {
	listenAddress := flag.String("l", DefaultListenAddress,
		fmt.Sprintf("Specify the address to listen on (env: %s)", ListenAddressEnv))
	dataSourceDir := flag.String("d", DefaultDataSourceDir,
		fmt.Sprintf("Specify the datasource directory to serve (env: %s)", DataSourceDirEnv))
	showVersion := flag.Bool("v", false, "Print version")
	flag.Parse()

	if *showVersion {
		fmt.Println(Version)
		os.Exit(0)
	}
	if *listenAddress == DefaultListenAddress {
		addr, ok := os.LookupEnv(ListenAddressEnv)
		if ok {
			*listenAddress = addr
		}
	}
	if *dataSourceDir == DefaultDataSourceDir {
		dir, ok := os.LookupEnv(DataSourceDirEnv)
		if ok {
			*dataSourceDir = dir
		}
	}
	run(*listenAddress, *dataSourceDir)
}

func run(listenAddress, dataSourceDir string) {
	srv := NewServer(listenAddress, dataSourceDir)

	idleConnsClosed := make(chan struct{})
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		signal.Notify(sigCh, syscall.SIGTERM)
		<-sigCh

		if err := srv.Shutdown(context.Background()); err != nil {
			log.Println("Failed to shutdown the server gracefully:", err.Error())
		}
		close(idleConnsClosed)
	}()

	log.Printf("listening on %s (datasource directory: %s)\n", listenAddress, dataSourceDir)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("error: %s\n", err.Error())
	}

	<-idleConnsClosed
}
