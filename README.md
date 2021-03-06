<p align="center">
<img width="400px" src="https://user-images.githubusercontent.com/16992394/74128390-be44df00-4be5-11ea-8cf0-03eec3cdeecb.png">
</p>
<h2 align="center">Rubban - Kibana Automatic Index Pattern Discovery and Other Curating Tasks.</h2>
<p align="center">
   <a>
      <img src="https://img.shields.io/badge/Kibana->=7-blue?style=flat&logo=kibana" alt="Elastic Stack Version 7^^">
   </a>
   <a>
      <img src="https://img.shields.io/github/v/tag/sherifabdlnaby/rubban?label=release&amp;sort=semver">
    </a>
     <a>
      <img src="https://github.com/sherifabdlnaby/rubban/workflows/Build/badge.svg">
    </a>
   <a href="https://hub.docker.com/repository/docker/sherifabdlnaby/rubban">
      <img src="https://badgen.net/docker/size/sherifabdlnaby/rubban?label=image%20size&icon=docker&t=1" alt="docker size">
   </a>
    <a href="https://hub.docker.com/repository/docker/sherifabdlnaby/rubban">
      <img src="https://img.shields.io/docker/pulls/sherifabdlnaby/rubban?logo=docker" alt="docker pulls">
   </a>
   <a href="https://goreportcard.com/report/github.com/sherifabdlnaby/rubban">
      <img src="https://goreportcard.com/badge/github.com/sherifabdlnaby/rubban" alt="Go Report">
   </a>
   <a href="https://github.com/sherifabdlnaby/rubban/issues">
        <img src="https://img.shields.io/github/issues/sherifabdlnaby/rubban.svg" alt="GitHub issues">
   </a>
   <a href="https://raw.githubusercontent.com/sherifabdlnaby/rubban/blob/master/LICENSE">
      <img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="GitHub license">
   </a>
    <a>
      <img src="https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat" alt="contributions welcome">
   </a>
</p>

# Introduction

Rubban for Elastic Stack is a curating tool that will automate tasks to make using Kibana a more pleasant experience.

# Features

### Automatic Index Pattern Discovery & Creation

In a dynamic environment where web services are created and deployed rapidly with all their logging infra set up, It is always annoying having to create Index Pattern *for each* service *manually* while everything else in the setup is automated.
Rubban uses Kibana's client API's and Automate Index Pattern Creation and Discovery using General Patterns.

### Automatic Refreshing for Index Pattern Fields

Kibana Index Pattern caches field names and types; when a new field is indexed after Index Pattern creation you won't be able to interact with it unless you *refresh Index Pattern field list*. Rubban can automate Index Pattern field list refreshing every set interval.

> Refreshing resets the popularity counter of each field.

### Automatic Creation for Dashboards

Still under development.

> Currently tested on Kibana 7.0 and greater versions.

#### Examples for Automatic Index Pattern Discovery

Say your indices has the following _convention_ for your apache access logs: `logs-apache-access-<service-name>-<date>` where `<service-name>` and `<date>` are dynamic based on the service and time.
You can have `logs-apache-access-serviceX-2020-02-02` and `logs-apache-access-serviceY-2020-02-02` and you'll need to create index patterns `logs-apache-access-serviceX-*` and `logs-apache-access-serviceX-*` respectively to have them appear nicely in Kibana for developers.

If a new service is deployed and shipping its logs to Elasticsearch, you will need to still manually create the index pattern substituting just the service name.
With Rubban *you can configure a _general_ pattern* like `logs-apache-access-?-*` (notice ? is in the service-name place), and then Rubban will query Kibana and Elasticsearch periodically to look for indices that match this pattern *that's not covered under an Index Pattern.


# Installation

## Build from Source
1. Install Go (for macOS `brew install go`)
2. `make build`
3. `./bin/rubban`

## Docker

#### via Docker CLI
```
docker run  --env='RUBBAN_KIBANA_HOST=https://kibana:5601' \
            --env='RUBBAN_KIBANA_USER=elastic' \
            --env='RUBBAN_KIBANA_PASSWORD=changeme' \
            --env='RUBBAN_AUTOINDEXPATTERN_ENABLED=true' \
            --env='RUBBAN_AUTOINDEXPATTERN_SCHEDULE=*/5 * * * *' \
            --env='RUBBAN_AUTOINDEXPATTERN_GENERALPATTERNS=[{"pattern":"logs-apache-access-*-?","timeFieldName":"@timestamp"}]' \
             sherifabdlnaby/rubban:latest
 ```

##### via Docker Compose

 `docker-compose up -d`

# Configuration

- Configuration is in `./rubban.yml` and file path can be overridden by the `RUBBAN_CONFIG_DIR` environment variable. (Configuration can be JSON, YAML, or TOML)
- Any configuration can be overridden with environment variables. ex: `kibana.user: elastic` can be overridden with `RUBBAN_KIBANA_USER=elastic2`.
    - Prefix key with `RUBBAN_`, ALL CAP all key, and replace `.` with `_` and any `_` to `__`.
    - Arrays can be declared in environment variables using 1. comma separated list, 2. json encoded array in a string.
    - Maps and objects can be declared in environment using a json encoded object in a string.

### Kibana

`kibana.host`: Kibana Host (with Port). if HTTPS is enabled make sure to add `https://` in the host. (*default:*  http://localhost:5601)

`kibana.user`: Kibana User. Make sure user has privilege for Kibana configuration and setup.

`kibana.password`: Kibana User's Password. (It's advised to use `RUBBAN_KIBANA_PASSWORD` Env variable instead of adding it to config in plaintext)

##### Example:
```yaml
kibana:
    host: kibana:5601
    user: elastic
    password: changeme
```

### Automatic Index Pattern Discovery & Creation

`autoIndexPattern.enabled`: Enable/Disable Auto Index Discovery & Creation

`autoIndexPattern.schedule`: A [Cron Expression](https://crontab.guru/) that specify fixed schedule to run Auto Index Discovery & Creation. (*default:*  */5 * * * * _every 5 minutes_)

`autoIndexPattern.concurrency`: Control How many Requests are made to Kibana API concurrently.  (*default:*  20)

`autoIndexPattern.generalPatterns`: An array of General Pattern Objects, where `pattern` is the *general pattern* used to discover indices and `timeFieldName` is the time field that will be used for the created index pattern.

#### How do General Pattern works ?

A general pattern should be general for both indices names and index patterns (applies to them both).  Unlike Kibana index pattern that can only contain wildcard `*`, general pattern has the `?` wildcard. It will be used to find indices that doesn't belong to any index pattern.

If Kibana has currently `logs-apache-access-serviceX-*` index pattern and `logs-apache-access-serviceX-2020-02-01` index. after a while `logs-apache-access-serviceX-2020-02-02` and `logs-apache-access-serviceY-2020-02-02` got created.
the next time Rubban run with general pattern `logs-apache-access-?-*`, it will automatically create `logs-apache-access-serviceY-*` index pattern that covers the newly created index.

##### Example:

```yaml
autoIndexPattern:
    enabled: true
    schedule: "* * * * *"
    concurrency: 20
    generalPatterns:
        -   pattern: logs-apache-access-*-?
            timeFieldName: "@timestamp"
```

### Automatic Refreshing for Index Pattern Field

`refreshIndexPattern.enabled`: Enable/Disable Auto Refreshing for Index Pattern Field

`refreshIndexPattern.schedule`: A [Cron Expression](https://crontab.guru/) that specify fixed schedule to run Auto Index Discovery & Creation. (*default:*  */5 * * * * _every 5 minutes_)

`refreshIndexPattern.concurrency`: Control How many Requests are made to Kibana API concurrently.  (*default:*  20)

`refreshIndexPattern.patterns`: An array of Patterns, where each pattern can match multiple index patterns. Similar to General Patterns explained above but without `?` matcher instead all uses `*`.

##### Example:

```yaml
refreshIndexPattern:
    enabled: true
    schedule: "* * * * *"
    concurrency: 10
    patterns:
        - logstash-apache-*-*-*
```

### Logging
```yaml
logging:
    level: info       # any of (debug|info|warn|fatal)
    debug: false      # enable/disable debug config
    color: true       # show color in output (not for json format)
    format: console   # any of (console|logfmt|json)
```

# License
[MIT License](https://raw.githubusercontent.com/sherifabdlnaby/rubban/master/LICENSE)
Copyright (c) 2020 Sherif Abdel-Naby

# Contribution

PR(s) are Open and Welcomed.
