# Atomic Banquet

A Modular Atom/RSS Feed Generator

## Usage

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

```
Usage of oneshot:
  -f string
    	Output format
  -h	Show help message
  -l	List available modules
  -m string
    	Module name
  -o string
    	Options (JSON formatted)
```

## Modules available:

  - books
	options:
	 - author
	 - language (default: en)

  - bugcrowd
	options:
	 - disclosures: bool (default: true)
	 - accepted: bool (default: true)

  - dockerhub
	options:
	 - image: image name (eg nbr23/atomic-banquet:latest)
	 - platform: image platform filter (linux/arm64, ...)

  - garmin-sdk
	options:
	 - sdks: list of names of the sdks to watch: fit, connect-iq (default: fit)

  - garmin-wearables

  - hackerone
	options:
	 - disclosed_only: bool (default: true)
	 - reports_count: int (default: 50)

  - hackeronePrograms
	options:
	 - results_count: int (default: 50)

  - infocon
	options:
	 - url: string

  - lego
	options:
	 - category: string (default: 'new', values: ['coming-soon', 'new'])

  - pentesterland

  - pocorgtfo

  - psupdates
	options:
	 - hardware: ps5 or ps4 (default: ps5)
	 - local: en-us or fr-fr (default: en-us)

