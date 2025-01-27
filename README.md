# mini-farm-tracker

Basic overview:
Provide a visualisation platform for various LoRaWAN sensors.
Initial design is required to track available water in water tanks.

MVP:

- [x] Raw data (from LoRaWAN nodes) propagate through the system and are saved.
- [?] This raw data is able to be queried from and displayed via graphs
- [ ] Configurations per sensors are able to be created; these are responsible to determine how the raw data is to be modified as then stored as calibrated data. These are to have "starting" times to all the sensor to be re-installed.
  - [ ] While for the MVP I want configurations to affect incoming data, the actual creation of these configurations should be behind a authentication. This can be V2.
- [ ] This calibrated data is also to be able to be viewed via the website
- [ ] User is able to re-calibrate the data; this will take all the raw data and apply all the calibrations created for the sensor.

## website

Using NodeJS v22.13.0

## server

Using Golang

Primary motivation is I have done similar in Python multiple times (Flask, Quart) and while I have created microservices within Golang, I have not used it for web API hosting.

# Diagrams

## Data Flow of data into the system

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

# Devices

## gateway - LPS8v2 Indoor LoRaWAN Multichannel Gateway


## Water tank - water levels - Dragino LDDS45

[Decoder](https://github.com/dragino/dragino-end-node-decoder)

[Dropbox](https://www.dropbox.com/scl/fo/ehbyykfvqb549beg69nly/AGksttCIAr55scX6QxXg4RA?rlkey=crbquiode216okgxrqleck654&e=1&dl=0)

[User Manual](https://wiki.dragino.com/xwiki/bin/view/Main/User%20Manual%20for%20LoRaWAN%20End%20Nodes/LDDS45%20-%20LoRaWAN%20Distance%20Detection%20Sensor%20User%20Manual/)

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

built in: [gin-gonic](https://gin-gonic.com/)

Hosted on: <b>DigitalOcean</b>

Domain established within DigitalOcean directing to droplet:
`mini-farm-tracker.io`

DNS Records are configured within DigitalOcean to allow for vercel website to be sued.

### Development

#### WebUI

Allow for asset generation
> npm install -g @vue/cli

> vue generate component MyComponent

#### Server

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

#### TODO:
Implement what I have done previously; allowing for initial, expected post data per collection to be tested after each test run.

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


# Rough TODO:

### WebUI
- [] Graphs
    - Initial HW will allow for 2 sensors; one for each tank.
- [ ] V2 will have auth, although as part of the purpose of this is a demo project, putting it behind a auth "wall" is counter productive initially.

### Server
- [x] Investigate HTTP servers - SSL secured and CORs established
    - Currently implmented with [gin](https://github.com/gin-gonic/gin)
- [x] Investigate Containerisation options.
    - Initial version is simply running binary.
    - solutions such as k8/docker (compose) are viable however k8 atleast is likely an overkill. All I want really want is crash/restart tolerance.
    - [x] For now a solution found using the App Platform within DigitalOcean. Allows:
        - [x] Health endpoints
        - [x] HA (defaults to 2x containers)
        - [x] Auto-deploy from git commit
        - [x] Automatic handling of SSLs.

### General

- [x] Connect MongoDB
    - [x] Account Created
    - The ideal is to try the Timeseries support. Historically not been MongoDB strong suite but have never personally tried it and apparently improved in v8.
    - [x] Schema definitions - Some additional information is diagram (./data structure.drawio.png)
        - [ ] Not all defined - enough for E2E functionality to be possible

- [x] Connect The Things Stack
    - [x] Account Created
    - Hardware provided
    - [x] Gateway
    - [x] 2x ultrasonic sensors to measure depth in tanks.
