package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	"github.com/nbr23/atomic-banquet/parser"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

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

func saveFeed(config *Config, feed *feeds.Feed, fileName string, feedConfig *FeedConfig) error {
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

func feedWorker(id int, feedJobs <-chan *FeedConfig, results chan<- error, config *Config) {
	for f := range feedJobs {
		module, ok := Modules[f.Module]
		fileName := f.Name
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

	feedJobs := make(chan *FeedConfig, len(config.Feeds))
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

		if f.Options.Get("private").(bool) {
			continue
		}
		fileName := f.Name
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

type runServerFlags struct {
	showHelp   bool
	configPath string
	serverPort string
}

func getRunServerFlags(f *runServerFlags) *flag.FlagSet {
	flags := flag.NewFlagSet("server", flag.ExitOnError)
	flags.BoolVar(&f.showHelp, "h", false, "Show help message")
	flags.StringVar(&f.configPath, "c", f.configPath, "Path to configuration file")
	configPath, found := os.LookupEnv(fmt.Sprintf("%sCONFIG_PATH", ENV_PREFIX))
	if !found {
		configPath = "./config.yaml"
	}
	f.configPath = configPath
	flags.StringVar(&f.serverPort, "p", os.Getenv("PORT"), "Server port")
	return flags
}

func responseLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		c.Next()
		duration := time.Since(t)
		fmt.Printf("%v | %d | %d | %s | %s | %s \"%s\"\n",
			time.Now().Format("2006/01/02 - 15:04:05"),
			c.Writer.Status(),
			c.Writer.Size(),
			duration,
			c.ClientIP(),
			c.Request.Method,
			c.Request.RequestURI,
		)
	}
}

func runServer(args []string) {
	var f runServerFlags

	flags := getRunServerFlags(&f)
	flags.Parse(args)

	if f.showHelp {
		flags.Usage()
		fmt.Println("Modules available:")
		printModulesHelp()
		return
	}

	if f.serverPort == "" {
		f.serverPort = "8080"
	}

	r := gin.New()
	r.Use(responseLogger())
	r.Use(gin.Recovery())

	r.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	for _, module := range Modules {
		p := module()
		parser.Route(r, p, parser.GetFullOptions(p))
	}

	r.Run(fmt.Sprintf(":%s", f.serverPort))
}

type runFetcherFlags struct {
	showHelp     bool
	configPath   string
	workersCount int
}

func getRunFetcherFlags(f *runFetcherFlags) *flag.FlagSet {
	flags := flag.NewFlagSet("fetcher", flag.ExitOnError)
	flags.BoolVar(&f.showHelp, "h", false, "Show help message")
	configPath, found := os.LookupEnv(fmt.Sprintf("%sCONFIG_PATH", ENV_PREFIX))
	if !found {
		configPath = "./config.yaml"
	}
	f.configPath = configPath
	flags.StringVar(&f.configPath, "c", f.configPath, "Path to configuration file")
	flags.IntVar(&f.workersCount, "w", 5, "Number of workers")
	return flags
}

func runFetcher(args []string) {
	var f runFetcherFlags

	flags := getRunFetcherFlags(&f)
	flags.Parse(args)

	if f.showHelp {
		flags.Usage()
		fmt.Println("Modules available:")
		printModulesHelp()
		return
	}

	config, err := getFeedsFromConfig(f.configPath)
	if err != nil {
		log.Fatal(err)
	}

	err = processFeeds(config, f.workersCount)
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

func runOneShot(args []string) {
	if len(args) < 1 {
		fmt.Println("Missing module name")
		printModulesHelp()
		return
	}

	m := getModule(args[0])
	if m == nil {
		log.Fatal(fmt.Errorf("module `%s` not found", args[0]))
	}
	flags := flag.NewFlagSet(fmt.Sprintf("oneshot %s", m), flag.ExitOnError)
	o := parser.GetFullOptions(m)
	o.AddFlags(flags)

	flags.Parse(args[1:])

	for _, option := range o.OptionsList {
		if option.Required {
			if o.Get(option.Flag) == "" {
				flags.Usage()
				log.Fatalf("missing required parameter: %s", option.Flag)
			}
		}
	}

	res, err := m.Parse(o)
	if err != nil {
		fmt.Println(parser.GetFullOptions(m).GetHelp())
		log.Fatal(err)
		return
	}

	var s string

	switch o.Get("f") {
	case "rss":
		s, err = res.ToRss()
	case "atom":
		s, err = res.ToAtom()
	case "json":
		s, err = res.ToJSON()
	default:
		s = fmt.Sprintf("%v", res)
	}
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(s)
}

func readMe(usage func()) {
	var (
		serverFlags  runServerFlags
		fetcherFlags runFetcherFlags
	)
	sf := getRunServerFlags(&serverFlags)
	ff := getRunFetcherFlags(&fetcherFlags)
	fmt.Println(`# Atomic Banquet

A Modular Atom/RSS Feed Generator

## Usage

` + "```")

	usage()
	fmt.Println("```\n" + `
### Server mode

` + "```")
	sf.Usage()
	fmt.Print("```\n\n")
	fmt.Print("### Fetcher mode\n\n")
	fmt.Println("```")
	ff.Usage()
	fmt.Print("```\n\n")
	fmt.Print("### Oneshot mode\n\nUsage: `atomic-banquet oneshot <module> [module options]`\n\n")
	fmt.Print("\n## Modules available:\n\n")
	printModulesHelp()
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  server: run atomic-banquet in server mode\n")
		fmt.Fprintf(os.Stderr, "  fetcher: run atomic-banquet in fetch mode based on a declarative config file\n")
		fmt.Fprintf(os.Stderr, "  oneshot: run atomic-banquet in oneshot mode to fetch a specific module's results\n")
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
	case "oneshot":
		runOneShot(os.Args[2:])
	case "readme":
		readMe(flag.Usage)
	default:
		flag.Usage()
		return
	}
}
