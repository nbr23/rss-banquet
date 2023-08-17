package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
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

	for _, feed := range config.Feeds {
		module, ok := Modules[feed.Module]
		if !ok {
			fmt.Printf("Module %s not found\n", feed.Module)
			return
		}
		feed, err := module().Parse(feed.Options)
		if err != nil {
			fmt.Print(err)
			return
		}
		atom, err := feed.ToAtom()
		if err != nil {
			fmt.Print(err)
			return
		}

		output_path := fmt.Sprintf("%s/%s.atom", config.OutputPath, feed.Title)
		out, err := os.Create(output_path)
		if err != nil {
			fmt.Print(err)
			return
		}
		defer out.Close()
		out.WriteString(atom)
	}
}
