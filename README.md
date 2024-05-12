# RSS Banquet

A Modular Atom/RSS Feed Generator

## Usage

```
Usage: ./rss-banquet <command> [options]
Commands:
  server: run rss-banquet in server mode
  fetcher: run rss-banquet in fetch mode based on a declarative config file
  oneshot: run rss-banquet in oneshot mode to fetch a specific module's results
```

### Server mode

```
Usage of server:
  -c string
    	Path to configuration file
  -h	Show help message
  -p string
    	Server port
```

### Fetcher mode

```
Usage of fetcher:
  -c string
    	Path to configuration file (default "./config.yaml")
  -h	Show help message
  -w int
    	Number of workers (default 5)
```

### Oneshot mode

Usage: `rss-banquet oneshot <module> [module options]`


## Modules available:

  - books
	 - feedFormat: feed output format (rss, atom, json) (default: atom)
	 - private: private feed (default: false)
	 - author: author of the books (default: )
	 - language: language of the books (default: en)

  - bugcrowd
	 - feedFormat: feed output format (rss, atom, json) (default: atom)
	 - private: private feed (default: false)
	 - disclosures: Show disclosure reports (default: )
	 - accepted: Show accepted reports (default: en)
	 - title: Feed title (default: Bugcrowd)
	 - description: Feed description (default: Bugcrowd Crowdstream)

  - dockerhub
	 - feedFormat: feed output format (rss, atom, json) (default: atom)
	 - private: private feed (default: false)
	 - image: image name (eg nbr23/rss-banquet:latest) (default: )
	 - platform: image platform filter (linux/arm64, ...) (default: )

  - garmin-sdk
	 - feedFormat: feed output format (rss, atom, json) (default: atom)
	 - private: private feed (default: false)
	 - sdks: list of names of the sdks to watch: fit, connect-iq (default: fit)

  - garmin-wearables
	 - feedFormat: feed output format (rss, atom, json) (default: atom)
	 - private: private feed (default: false)

  - hackerone
	 - feedFormat: feed output format (rss, atom, json) (default: atom)
	 - private: private feed (default: false)
	 - disclosed_only: Show only disclosed reports (default: true)
	 - reports_count: Number of reports to display (default: 50)
	 - title: Feed title (default: HackerOne)
	 - description: Feed description (default: Hackerone Hacktivity)

  - hackeronePrograms
	 - feedFormat: feed output format (rss, atom, json) (default: atom)
	 - private: private feed (default: false)
	 - results_count: Number of programs to display (default: 50)
	 - title: Feed title (default: HackerOne Programs)
	 - description: Feed description (default: Hackerone Program Launch)

  - infocon
	 - feedFormat: feed output format (rss, atom, json) (default: atom)
	 - private: private feed (default: false)
	 - url: url of the infocon (default: )

  - lego
	 - feedFormat: feed output format (rss, atom, json) (default: atom)
	 - private: private feed (default: false)
	 - category: category of the lego products (new, coming-soon) (default: new)

  - pentesterland
	 - feedFormat: feed output format (rss, atom, json) (default: atom)
	 - private: private feed (default: false)

  - pocorgtfo
	 - feedFormat: feed output format (rss, atom, json) (default: atom)
	 - private: private feed (default: false)

  - psupdates
	 - feedFormat: feed output format (rss, atom, json) (default: atom)
	 - private: private feed (default: false)
	 - hardware: hardware of the updates (default: ps5)
	 - local: local of the updates (default: en-us)

