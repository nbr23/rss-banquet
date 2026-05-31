# RSS Banquet

A Modular Atom/RSS Feed Generator

## Usage

```
Usage: rss-banquet <command> [options]
Commands:
  server: run rss-banquet in server mode
  oneshot: run rss-banquet in oneshot mode to fetch a specific module's results
```

## Global options

The following environment variables can be used to configure the application:

-  `BANQUET_GLOBAL_LOG_LEVEL`: Log level (trace, debug, info, warn, error, fatal, panic, disabled) (default: info)
-  `BANQUET_GLOBAL_USER_AGENT`: User agent to use for HTTP requests
-  `BANQUET_SERVER_SERVER_PORT`: Port to listen on in server mode (default: 8080)


### Server mode

```
Usage of server:
  -h	Show help message
  -p string
    	Server port (default: 8080)
```

### Oneshot mode

Usage: `rss-banquet oneshot <module> [module options]`


## Modules available:

  - anthropic-api-changelog
	 - feedFormat: feed output format (rss, atom, json) (default: rss)
	 - route: route to expose the feed (default: anthropic-api-changelog)

  - bugcrowd
	 - feedFormat: feed output format (rss, atom, json) (default: rss)
	 - route: route to expose the feed (default: bugcrowd)
	 - disclosures: Show disclosure reports (default: true)
	 - accepted: Show accepted reports (default: false)
	 - title: Feed title (default: Bugcrowd)
	 - description: Feed description (default: Bugcrowd Crowdstream)

  - costco
	 - feedFormat: feed output format (rss, atom, json) (default: rss)
	 - route: route to expose the feed (default: costco)
	 - url: URL of the Costco page to scrape (default: )

  - dockerhub
	 - feedFormat: feed output format (rss, atom, json) (default: rss)
	 - route: route to expose the feed (default: dockerhub)
	 - image: image name (eg nbr23/rss-banquet:latest) (default: )
	 - platform: image platform filter (linux/arm64, ...) (default: )

  - garmin-wearables
	 - feedFormat: feed output format (rss, atom, json) (default: rss)
	 - route: route to expose the feed (default: garminwearables)

  - gemini-api-changelog
	 - feedFormat: feed output format (rss, atom, json) (default: rss)
	 - route: route to expose the feed (default: gemini-api-changelog)

  - github-notifications
	 - feedFormat: feed output format (rss, atom, json) (default: rss)
	 - route: route to expose the feed (default: github-notifications)
	 - all: Include read notifications (default: false)
	 - participating: Only show notifications where user is directly participating or mentioned (default: false)
	 - age: Only show notifications from the past duration (e.g., 24h, 7d, 2w) (default: )
	 - before: Only show notifications updated before this ISO 8601 timestamp (YYYY-MM-DDTHH:MM:SSZ) (default: )
	 - org: Filter by organization/owner name (client-side filter) (default: )
	 - reason: Filter by notification reason: assign, author, comment, ci_activity, invitation, manual, mention, review_requested, security_alert, state_change, subscribed, team_mention (client-side filter) (default: )
	 - title: Feed title (default: GitHub Notifications)
	 - description: Feed description (default: GitHub Notifications Feed)

  - goodreads
	 - feedFormat: feed output format (rss, atom, json) (default: rss)
	 - route: route to expose the feed (default: goodreads)
	 - authorId: Goodreads author ID (default: )
	 - seriesId: Goodreads series ID (default: )
	 - year-min: minimum year of publication (default: 2025)
	 - language: language of the book (default: en)
	 - bookFormats: seeked formats of the book (paperback, hardcover, ebook, audiobook, etc.) (default: paperback,hardcover,kindle,ebook)

  - hackerone
	 - feedFormat: feed output format (rss, atom, json) (default: rss)
	 - route: route to expose the feed (default: hackerone)
	 - disclosed_only: Show only disclosed reports (default: true)
	 - reports_count: Number of reports to display (default: 50)
	 - title: Feed title (default: HackerOne)
	 - description: Feed description (default: Hackerone Hacktivity)

  - hackeronePrograms
	 - feedFormat: feed output format (rss, atom, json) (default: rss)
	 - route: route to expose the feed (default: hackeroneprograms)
	 - results_count: Number of programs to display (default: 50)
	 - title: Feed title (default: HackerOne Programs)
	 - description: Feed description (default: Hackerone Program Launch)

  - imdb
	 - feedFormat: feed output format (rss, atom, json) (default: rss)
	 - route: route to expose the feed (default: imdb)
	 - artistId: IMDB artist ID (e.g. nm0000138) (default: )
	 - first: max number of credits to return (default: 25)
	 - titleType: filter by title type (e.g. movie, short, tvSeries, tvMiniSeries, tvMovie, tvSpecial, tvShort, video, videoGame, musicVideo, podcastSeries) (default: )
	 - category: filter by credit category (e.g. actor, director, writer, producer, self, soundtrack, archive_footage, art_department, animation_department, miscellaneous, thanks) (default: )

  - infocon
	 - feedFormat: feed output format (rss, atom, json) (default: rss)
	 - route: route to expose the feed (default: infocon)
	 - url: url of the infocon (default: )

  - lego
	 - feedFormat: feed output format (rss, atom, json) (default: rss)
	 - route: route to expose the feed (default: lego)
	 - category: category of the lego products (new, coming-soon) (default: new)

  - nytimes
	 - feedFormat: feed output format (rss, atom, json) (default: rss)
	 - route: route to expose the feed (default: nytimes)
	 - author: author of the articles to fetch (default: )

  - openai-api-changelog
	 - feedFormat: feed output format (rss, atom, json) (default: rss)
	 - route: route to expose the feed (default: openai-api-changelog)

  - openai-chatgpt-release-notes
	 - feedFormat: feed output format (rss, atom, json) (default: rss)
	 - route: route to expose the feed (default: openai-chatgpt-release-notes)

  - pentesterland
	 - feedFormat: feed output format (rss, atom, json) (default: rss)
	 - route: route to expose the feed (default: pentesterland)

  - pocorgtfo
	 - feedFormat: feed output format (rss, atom, json) (default: rss)
	 - route: route to expose the feed (default: pocorgtfo)

  - psupdates
	 - feedFormat: feed output format (rss, atom, json) (default: rss)
	 - route: route to expose the feed (default: psupdates)
	 - hardware: hardware of the updates (default: ps5)
	 - local: local of the updates (default: en-us)


## Removed modules

- `garmin-sdk`: removed because developer.garmin.com no longer hosts the FIT SDK download page. Subscribe to the upstream GitHub releases atom feed instead: https://github.com/garmin/fit-sdk-tools/releases.atom
