// Copyright © 2021 Valentin Slyusarev <va.slyusarev@gmail.com>
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"

	"golang.org/x/net/webdav"
)

var version = "develop"

var dir = flag.String("d", path.Join("."), "WebDav server directory.")
var port = flag.Int("p", 80, "WebDav server port.")

func checkFlags() {
	flag.Parse()

	exitWithError := func(err error) {
		log.Printf("%v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	dir, err := filepath.Abs(*dir)
	if err != nil {
		exitWithError(fmt.Errorf("broken dir path %q: %v", dir, err))
	}

	s, err := os.Stat(dir)
	if err != nil {
		exitWithError(fmt.Errorf("broken dir path: %v", err))
	}

	if !s.IsDir() {
		exitWithError(fmt.Errorf("broken dir path %q: is not dir", dir))
	}
}

func localIP() string {
	if addrs, err := net.InterfaceAddrs(); err == nil {
		for _, address := range addrs {
			// check the address type and if it is not a loopback the display it
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}
	return ""
}

func serve(ctx context.Context, dir string, port int) {

	srv := http.Server{
		Addr: fmt.Sprintf(":%d", port),
		Handler: &webdav.Handler{
			FileSystem: webdav.Dir(dir),
			LockSystem: webdav.NewMemLS(),
			Logger: func(r *http.Request, err error) {
				if err == nil {
					log.Printf("%s (%s) -> %s (%s)\n", r.RemoteAddr, r.UserAgent(), r.URL, r.Method)
				}
				if err != nil {
					log.Printf("ERRROR %v\n", err)
				}
			},
		},
	}

	errChan := make(chan error)
	go func() {
		errChan <- srv.ListenAndServe()
	}()

	log.Printf("server start on %s:%d", localIP(), port)

	for {
		select {
		case <-ctx.Done():
			_ = srv.Shutdown(ctx)
			log.Printf("server terminated\n")
			return
		case err := <-errChan:
			log.Printf("server terminated by error: %v\n", err)
			return
		}
	}
}

func main() {
	log.SetPrefix("wds: ")
	log.SetFlags(0)
	log.Printf("version: %s\n", version)

	checkFlags()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, os.Kill, syscall.SIGTERM)
		<-stop
		log.Printf("detect interrupt, terminate after few seconds...\n")
		cancel()
	}()

	serve(ctx, *dir, *port)
}
