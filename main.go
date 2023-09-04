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

	"github.com/gorilla/feeds"
	"github.com/nbr23/atomic-banquet/parser"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func printHelp() {
	fmt.Println("Usage: atomic-banquet [-hi] [-c config]")
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

func saveToS3(atom string, outputPath string, fileName string) error {
	s, err := session.NewSession(&aws.Config{})
	if err != nil {
		return err
	}
	s3Client := s3.New(s)

	bucketUri := strings.SplitN(strings.TrimPrefix(outputPath, "s3://"), "/", 2)
	bucketName := bucketUri[0]
	objectKey := strings.Join(append(bucketUri[1:], fmt.Sprintf("%s.atom", fileName)), "/")

	contentBytes := []byte(atom)

	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(contentBytes),
		ContentType: aws.String("application/atom+xml"),
	})
	if err != nil {
		return err
	}

	fmt.Println("Uploaded content to S3 successfully!")
	return nil
}

func saveFeed(config *Config, feed *feeds.Feed, fileName string) error {
	atom, err := feed.ToAtom()
	if err != nil {
		return err
	}

	if strings.HasPrefix(config.OutputPath, "s3://") {
		return saveToS3(atom, config.OutputPath, fileName)
	}

	output_path := fmt.Sprintf("%s/%s.atom", config.OutputPath, fileName)
	out, err := os.Create(output_path)
	if err != nil {
		return err
	}
	defer out.Close()
	out.WriteString(atom)
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
		err = saveFeed(config, feed, fileName)
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
		fileName := parser.DefaultedGet(f.Options, "filename", f.Name)
		index.WriteString(fmt.Sprintf("<li><a target=\"_blank\" href=\"%s.atom\">%s</a></li>\n", fileName, f.Name))
	}
	index.WriteString("</ul>\n</body>\n</html>")
	output_path := fmt.Sprintf("%s/index.html", config.OutputPath)
	out, err := os.Create(output_path)
	if err != nil {
		return err
	}
	defer out.Close()
	out.WriteString(index.String())
	return nil
}

func main() {
	var (
		showHelp     bool
		configPath   string
		workersCount int
		buildIndex   bool
	)

	flag.BoolVar(&showHelp, "h", false, "Show help message")
	configPath, found := os.LookupEnv(fmt.Sprintf("%sCONFIG_PATH", ENV_PREFIX))
	if !found {
		configPath = "./config.yaml"
	}
	flag.StringVar(&configPath, "c", configPath, "Path to configuration file")
	flag.IntVar(&workersCount, "w", 5, "Number of workers")
	flag.BoolVar(&buildIndex, "i", false, "Build index.html")
	flag.Parse()

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
	if buildIndex {
		err = buildIndexHtml(config)
		if err != nil {
			log.Fatal(err)
		}
	}
}
