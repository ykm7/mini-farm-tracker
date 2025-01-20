# mini-farm-tracker

Basic overview:
Provide a visualisation platform for various LoRaWAN sensors.
Initial design is required to track available water in water tanks.

Current Implementation

## website

Using NodeJS v22.13.0


## server

Using Golang

Primary motivation is I have done similar in Python multiple times (Flask, Quart) and while I have created microservices within Golang, I have not used it for web API hosting.

# Hosting:

## WebUI

Hosted on: <b>[vercel](https://vercel.com)</b>

vercel CLI is used to deploy when required.

### Environment variables
Once updated via the vercel dashboard, it is important to pull them locally.

This will pull the "production" environment fields to test local development against the production server.
> vercel env pull --environment=production .env.production

From here can use the vercel deploy steps within the `package.json` file. 

## Server

Hosted on: <b>DigitalOcean</b>

Domain established within DigitalOcean directing to droplet:
`mini-farm-tracker.io`

DNS Records are configured within DigitalOcean to allow for vercel website to be sued.

### Development

> git checkout .
> git clean -fd

> go build
> export GIN_MODE=release && ./mini-farm-tracker-server

#### Network

Firewall options - inbound port of 3000 (TCP) required

SSL certificate (Let's Encrypt) created on domain bought from namecheap.

### Testing

#### testContainer (currently only implemented for server - mongoDB)

NOTE: testContainer can use cloud resources however prefer to run locally.

#### Test result generation

##### Server

> go test ./... -coverprofile=coverage.out
> go tool cover -html coverage -o coverage.html

Requires:
* Docker Desktop
* [testContainer](https://app.testcontainers.cloud/accounts/14403/dashboard/install?target=windows-desktop)

With my environment, I have problems with the embedded testContainers cleanup logic.

Following [configuration path](https://golang.testcontainers.org/features/configuration/) adding a line to disable `ryuk` allows correct running:

> ryuk.disabled=true

# Data Flow

```mermaid
---
title: Sensor Data into the Server
---
flowchart TD
    webhook@{ shape: bow-rect, label: "webhook" }
    api_valid@{ shape: diamond, label: "api\nkey\nvalid" }
    raw_data@{ shape: bow-rect, label: "raw data" }
    calibrated_data@{ shape: bow-rect, label: "calibrated data" }
    mongo@{ shape: cyl, label: "MongoDB" }
    Config_exists@{ shape: diamond, label: "'Configuration'\nexists?" }

    start@{ shape: sm-circ } --> webhook
    webhook[webhook] --> api_valid{API key valid?}

    subgraph Server
    api_valid -->|Yes| sensor_exists{Sensor exists?}
    api_valid -->|No| Dropped[Dropped]

    sensor_exists --->|Yes| raw_data
    sensor_exists --->|No| Dropped

    raw_data --> Config_exists -->|Yes| calibrated_data
    end

    raw_data --> mongo
    calibrated_data --> mongo

    Config_exists -->|No| Stop
    Dropped --> Stop
    mongo --> Stop@{ shape: fr-circ }
```

# TODO:

### WebUI
- [] Graphs
    - Initial HW will allow for 2 sensors; one for each tank.
- [ ] V2 will have auth, although as part of the purpose of this is a demo project, putting it behind a auth "wall" is counter productive initially.

### Server
- [ ] Investigate HTTP servers 
    - Currently implmented with [gin](https://github.com/gin-gonic/gin)
- [ ] Investigate Containerisation options.
    - Initial version is simply running binary.
    - solutions such as k8/docker (compose) are viable however k8 atleast is likely an overkill. All I want really want is crash/restart tolerance.
    - [x] For now a solution found using the App Platform within DigitalOcean. Allows:
        - [x] Health endpoints
        - [x] HA (defaults to 2x containers)
        - [x] Auto-deploy from git commit
        - [x] Automatic handling of SSLs.

### General

- [ ] Connect MongoDB
    - [x] Account Created
    - The ideal is to try the Timeseries support. Historically not been MongoDB strong suite but have never personally tried it and apparently improved in v8.
    - [] Schema definitions - Some additional information is diagram (./data structure.drawio.png)

- [x] Connect The Things Stack
    - [x] Account Created
    - Hardware provided
    - [x] Gateway
    - [x] 2x ultrasonic sensors to measure depth in tanks.
