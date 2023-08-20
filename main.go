package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/gorilla/feeds"

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

func saveToS3(atom string, outputPath string, feedName string) error {
	s, err := session.NewSession(&aws.Config{})
	if err != nil {
		return err
	}
	s3Client := s3.New(s)

	bucketUri := strings.SplitN(strings.TrimPrefix(outputPath, "s3://"), "/", 2)
	bucketName := bucketUri[0]
	objectKey := strings.Join(append(bucketUri[1:], fmt.Sprintf("%s.atom", feedName)), "/")

	contentBytes := []byte(atom)

	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(contentBytes),
		ContentType: aws.String("application/atom+xml"),
	})
	if err != nil {
		fmt.Println("Error uploading to S3:", err)
		return err
	}

	fmt.Println("Uploaded content to S3 successfully!")
	return nil
}

func saveFeed(config *Config, feed *feeds.Feed, f FeedConfig) error {
	atom, err := feed.ToAtom()
	if err != nil {
		fmt.Print(err)
		return err
	}

	if strings.HasPrefix(config.OutputPath, "s3://") {
		return saveToS3(atom, config.OutputPath, f.Name)
	}

	output_path := fmt.Sprintf("%s/%s.atom", config.OutputPath, f.Name)
	out, err := os.Create(output_path)
	if err != nil {
		fmt.Print(err)
		return err
	}
	defer out.Close()
	out.WriteString(atom)
	return nil
}

func main() {
	var (
		showHelp   bool
		configPath string
	)

	flag.BoolVar(&showHelp, "h", false, "Show help message")
	flag.StringVar(&configPath, "c", "./config.yaml", "Path to configuration file")
	flag.Parse()

	if showHelp {
		printHelp()
		return
	}

	config, err := getFeedsFromConfig(configPath)
	if err != nil {
		fmt.Print(err)
		return
	}

	for _, f := range config.Feeds {
		module, ok := Modules[f.Module]
		if !ok {
			fmt.Printf("Module %s not found\n", f.Module)
			return
		}
		feed, err := module().Parse(f.Options)
		if err != nil {
			fmt.Print(err)
			return
		}
		err = saveFeed(config, feed, f)
		if err != nil {
			fmt.Print(err)
		}
	}
}
