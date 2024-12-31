package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nbr23/rss-banquet/config"
	"github.com/nbr23/rss-banquet/parser"
	"github.com/nbr23/rss-banquet/style"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type runServerFlags struct {
	showHelp   bool
	serverPort string
}

func getRunServerFlags(f *runServerFlags) *flag.FlagSet {
	flags := flag.NewFlagSet("server", flag.ExitOnError)
	flags.BoolVar(&f.showHelp, "h", false, "Show help message")
	flags.StringVar(&f.serverPort, "p", config.GetConfigOption("BANQUET_SERVER_PORT"), "Server port")
	return flags
}

func responseLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		var logger *zerolog.Event
		if c.Writer.Status() >= 400 && c.Writer.Status() < 600 {
			logger = log.Error()
		} else {
			logger = log.Info()
		}

		event := logger.
			Int("status", c.Writer.Status()).
			Int("size", c.Writer.Size()).
			Dur("duration", duration).
			Str("client_ip", c.ClientIP()).
			Str("method", c.Request.Method).
			Str("path", c.Request.RequestURI)

		if len(c.Errors) > 0 {
			event = event.Err(c.Errors.Last())
		}
		event.Msgf("%s %s", c.Request.Method, c.Request.RequestURI)
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
	r.GET("/rss-style.xsl", func(c *gin.Context) {
		c.Header("Content-Type", "text/xsl")
		c.String(200, style.RssStyle)
	})
	r.GET("/atom-style.xsl", func(c *gin.Context) {
		c.Header("Content-Type", "text/xsl")
		c.String(200, style.AtomStyle)
	})

	r.Run(fmt.Sprintf(":%s", f.serverPort))
}

func runOneShot(args []string) {
	if len(args) < 1 {
		fmt.Println("Missing module name")
		printModulesHelp()
		return
	}

	m := getModule(args[0])
	if m == nil {
		log.Fatal().Msg(fmt.Sprintf("module `%s` not found", args[0]))
	}
	flags := flag.NewFlagSet(fmt.Sprintf("oneshot %s", m), flag.ExitOnError)
	o := parser.GetFullOptions(m)
	o.AddFlags(flags)

	flags.Parse(args[1:])

	for _, option := range o.OptionsList {
		if option.Required {
			if o.Get(option.Flag) == "" {
				flags.Usage()
				log.Fatal().Msg(fmt.Sprintf("missing required parameter: %s", option.Flag))
			}
		}
	}

	res, err := m.Parse(o)

	if err != nil {
		fmt.Println(parser.GetFullOptions(m).GetHelp())
		log.Fatal().Msg(err.Error())
		return
	}

	parser.SortFeedEntries(res)

	var s string

	switch o.Get("feedFormat") {
	case "rss":
		s, err = res.ToRss()
	case "atom":
		s, err = res.ToAtom()
	case "json":
		s, err = res.ToJSON()
	case "text":
		s = *parser.FeedToText(res)
	default:
		s = fmt.Sprintf("%v", res)
	}
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	fmt.Println(s)
}

func readMe(usage func()) {
	var serverFlags runServerFlags

	sf := getRunServerFlags(&serverFlags)
	fmt.Println(`# RSS Banquet

A Modular Atom/RSS Feed Generator

## Usage

` + "```")

	usage()
	fmt.Println("```\n" + `
## Global options

` + config.ReadmeText() + `

### Server mode

` + "```")
	sf.Usage()
	fmt.Print("```\n\n")
	fmt.Print("### Oneshot mode\n\nUsage: `rss-banquet oneshot <module> [module options]`\n\n")
	fmt.Print("\n## Modules available:\n\n")
	printModulesHelp()
}

func initLogging() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	logLevel := config.GetConfigOption("LOG_LEVEL")
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(level)
	}
}

func main() {
	config.InitConfig()
	initLogging()

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  server: run rss-banquet in server mode\n")
		fmt.Fprintf(os.Stderr, "  oneshot: run rss-banquet in oneshot mode to fetch a specific module's results\n")
	}
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	switch flag.Arg(0) {
	case "server":
		runServer(os.Args[2:])
	case "oneshot":
		runOneShot(os.Args[2:])
	case "readme":
		readMe(flag.Usage)
	default:
		flag.Usage()
		return
	}
}
