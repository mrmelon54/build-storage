package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	buildVersion = "develop"
	buildDate    = ""
)

func main() {
	log.Printf("[Main] Starting up Build Storage #%s (%s)\n", buildVersion, buildDate)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Error loading .env file")
	}

	buildDir := os.Getenv("BUILD_DIR")
	stat, err := os.Stat(buildDir)
	if err != nil {
		log.Fatalln("Build dir error:", err)
	}
	if !stat.IsDir() {
		log.Fatalln("Build dir is not a directory")
	}

	httpServer := setupHttpServer()

	//=====================
	// Safe shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Printf("\n")
		log.Printf("[Main] Attempting safe shutdown\n")
		a := time.Now()
		log.Printf("[Main] Shutting down HTTP server...\n")
		err := httpServer.Close()
		if err != nil {
			log.Println(err)
		}
		log.Printf("[Main] Signalling program exit...\n")
		b := time.Now().Sub(a)
		log.Printf("[Main] Took '%s' to fully shutdown modules\n", b.String())
		wg.Done()
	}()
	//
	//=====================
	wg.Wait()
	log.Println("[Main] Goodbye")
}
