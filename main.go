package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	"github.com/nbr23/atomic-banquet/parser"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func printHelp() {
	fmt.Println("Usage: atomic-banquet [-h] [-c config]")
	flag.PrintDefaults()
	fmt.Println("\nModules available:")

	sortedModules := make([]string, 0, len(Modules))
	for key := range Modules {
		sortedModules = append(sortedModules, key)
	}
	sort.Strings(sortedModules)

	for _, module := range sortedModules {
		fmt.Printf("  - %s\n%s\n", module, Modules[module]().Help())
	}
}

func saveToS3(atom string, outputPath string, fileName string, contentType string) error {
	s, err := session.NewSession(&aws.Config{})
	if err != nil {
		return err
	}
	s3Client := s3.New(s)

	bucketUri := strings.SplitN(strings.TrimPrefix(outputPath, "s3://"), "/", 2)
	bucketName := bucketUri[0]
	objectKey := strings.Join(append(bucketUri[1:], fileName), "/")

	contentBytes := []byte(atom)

	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(contentBytes),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return err
	}

	fmt.Println("Uploaded content to S3 successfully!")
	return nil
}

func saveFeed(config *Config, feed *feeds.Feed, fileName string, feedConfig FeedConfig) error {
	var feedString string
	var err error
	var fName string
	var contentType string
	if feedConfig.FeedType == "rss" {
		feedString, err = feed.ToRss()
		fName = fmt.Sprintf("%s.rss", fileName)
		contentType = "application/rss+xml"
	} else {
		feedString, err = feed.ToAtom()
		fName = fmt.Sprintf("%s.atom", fileName)
		contentType = "application/atom+xml"
	}
	if err != nil {
		return err
	}

	if strings.HasPrefix(config.OutputPath, "s3://") {
		return saveToS3(feedString, config.OutputPath, fName, contentType)
	}

	output_path := fmt.Sprintf("%s/%s", config.OutputPath, fName)
	out, err := os.Create(output_path)
	if err != nil {
		return err
	}
	defer out.Close()
	out.WriteString(feedString)
	return nil
}

func feedWorker(id int, feedJobs <-chan FeedConfig, results chan<- error, config *Config) {
	for f := range feedJobs {
		module, ok := Modules[f.Module]
		fileName := parser.DefaultedGet(f.Options, "filename", f.Name)
		if !ok {
			results <- fmt.Errorf("module %s not found", f.Module)
			return
		}
		feed, err := module().Parse(f.Options)
		if err != nil {
			results <- fmt.Errorf("[%s] %w", f.Name, err)
			return
		}
		if feed == nil {
			results <- fmt.Errorf("feed %s is empty", f.Name)
			return
		}
		err = saveFeed(config, feed, fileName, f)
		if err != nil {
			results <- fmt.Errorf("[%s] %w", f.Name, err)
			return
		}
		results <- nil
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func processFeeds(config *Config, workersCount int) error {
	wc := min(workersCount, len(config.Feeds))
	var returnedErrors error

	feedJobs := make(chan FeedConfig, len(config.Feeds))
	errorsChan := make(chan error, len(config.Feeds))

	for w := 0; w < wc; w++ {
		go feedWorker(w, feedJobs, errorsChan, config)
	}

	for _, f := range config.Feeds {
		feedJobs <- f
	}
	close(feedJobs)

	for i := 0; i < len(config.Feeds); i++ {
		err := <-errorsChan
		if err != nil {
			returnedErrors = errors.Join(returnedErrors, err)
		}
	}
	return returnedErrors
}

func buildIndexHtml(config *Config) error {
	var index strings.Builder
	index.WriteString("<html><head><title>Atomic Banquet</title></head>\n<body>\n<h1><a target=\"_blank\" href=\"https://github.com/nbr23/atomic-banquet/\">Atomic Banquet's</a> RSS/Atom Feeds Index</h1>\n<ul>\n")
	for _, f := range config.Feeds {
		if parser.DefaultedGet(f.Options, "private", false) {
			continue
		}
		fileName := parser.DefaultedGet(f.Options, "filename", f.Name)
		index.WriteString(fmt.Sprintf("<li><a target=\"_blank\" href=\"%s.atom\">%s</a></li>\n", fileName, f.Name))
	}
	index.WriteString("</ul>\n</body>\n</html>")

	if strings.HasPrefix(config.OutputPath, "s3://") {
		return saveToS3(index.String(), config.OutputPath, "index.html", "text/html")
	}

	output_path := fmt.Sprintf("%s/index.html", config.OutputPath)
	out, err := os.Create(output_path)
	if err != nil {
		return err
	}
	defer out.Close()
	out.WriteString(index.String())
	return nil
}

func runServer(args []string) {
	var (
		configPath string
		serverPort string
	)

	configPath, found := os.LookupEnv(fmt.Sprintf("%sCONFIG_PATH", ENV_PREFIX))
	if !found {
		configPath = "./config.yaml"
	}
	serverFlags := flag.NewFlagSet("server", flag.ExitOnError)
	serverFlags.StringVar(&configPath, "c", configPath, "Path to configuration file")
	serverFlags.StringVar(&serverPort, "p", os.Getenv("PORT"), "Server port")
	serverFlags.Parse(args)

	if serverPort == "" {
		serverPort = "8080"
	}

	r := gin.Default()

	r.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	for _, module := range Modules {
		p := module()
		p.Route(r)
	}

	r.Run(fmt.Sprintf(":%s", serverPort))
}

func runFetcher(args []string) {
	var (
		showHelp     bool
		configPath   string
		workersCount int
	)

	fetcherFlags := flag.NewFlagSet("fetcher", flag.ExitOnError)
	fetcherFlags.BoolVar(&showHelp, "h", false, "Show help message")
	configPath, found := os.LookupEnv(fmt.Sprintf("%sCONFIG_PATH", ENV_PREFIX))
	if !found {
		configPath = "./config.yaml"
	}
	fetcherFlags.StringVar(&configPath, "c", configPath, "Path to configuration file")
	fetcherFlags.IntVar(&workersCount, "w", 5, "Number of workers")
	fetcherFlags.Parse(args)

	if showHelp {
		printHelp()
		return
	}

	config, err := getFeedsFromConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	err = processFeeds(config, workersCount)
	if err != nil {
		log.Fatal("Errors during feeds processing:\n", err)
	}
	if config.BuildIndex {
		err = buildIndexHtml(config)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  server: run atomic-banquet in server mode\n")
		fmt.Fprintf(os.Stderr, "  fetcher: run atomic-banquet in fetch mode based on a declarative config file\n")
	}
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	switch flag.Arg(0) {
	case "server":
		runServer(os.Args[2:])
	case "fetcher":
		runFetcher(os.Args[2:])
	default:
		flag.Usage()
		return
	}
}
