package main

import (
	"build-storage/api"
	"build-storage/manager"
	"build-storage/structure"
	"build-storage/web"
	"fmt"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
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

	baseDir := os.Getenv("BASE_DIR")
	stat, err := os.Stat(baseDir)
	if err != nil {
		log.Fatalln("BASE_DIR error:", err)
	}
	if !stat.IsDir() {
		log.Fatalln("BASE_DIR is not a directory")
	}

	configFile, err := os.Open(path.Join(baseDir, "config.yml"))
	if err != nil {
		log.Fatalln("Failed to open config.yml")
	}

	var configYml structure.ConfigYaml
	groupsDecoder := yaml.NewDecoder(configFile)
	err = groupsDecoder.Decode(&configYml)
	if err != nil {
		log.Fatalln("Failed to parse config.yml:", err)
	}

	stat, err = os.Stat(path.Join(baseDir, configYml.BuildDir))
	if err != nil {
		log.Fatalln("buildDir error:", err)
	}
	if !stat.IsDir() {
		log.Fatalln("buildDir is not a directory")
	}

	buildManager := manager.New(baseDir, configYml)
	webServer := web.SetupWebServer(configYml, buildManager)
	apiServer := api.SetupApiServer(configYml, buildManager)
	runHttpServer(apiServer, "Web Server shutdown successfully")
	runHttpServer(webServer, "API Server shutdown successfully")

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
		err = webServer.Close()
		if err != nil {
			log.Println(err)
		}
		err = apiServer.Close()
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

func runHttpServer(httpServer *http.Server, closeMessage string) {
	go func() {
		err := httpServer.ListenAndServe()
		if err != nil {
			if err == http.ErrServerClosed {
				log.Println(closeMessage)
			} else {
				log.Fatalln(err)
			}
		}
	}()
}
