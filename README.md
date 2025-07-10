# mini-farm-tracker

## Security Status

| Frontend             | Backend         |
|----------------------|-----------------|
| [![Known Vulnerabilities](https://snyk.io/test/github/ykm7/mini-farm-tracker/badge.svg?targetFile=webui/package.json&style=flat-square)](https://snyk.io/test/github/ykm7/mini-farm-tracker?targetFile=webui/package.json) | [UNKNOWN] Snyk issue linking to my server files |

### Frontend (Vue)


### Backend (Gin)


Here’s a Markdown table with each Snyk badge in its own column, using "Frontend" and "Backend" as headers. Replace `your-username` and `your-repo` with your actual GitHub username and repository name.

This layout places each badge under its respective project component, making your repository’s security status clear and organised.

## Basic overview:
Provide a visualisation platform for various LoRaWAN sensors.
Initial design is required to track available water in water tanks.

MVP v1.0:

- [x] Raw data (from LoRaWAN nodes) propagate through the system and are saved.
  - [x] Two ultrasonic sensors
  - [x] Single 8-in-1 weather station
- [x] This raw data is able to be queried from and displayed via graphs
- [x] This calibrated data is also to be able to be viewed via the website
- [x] Cron-style aggregation functionality exists to perform periodic data manipulations.
  - [x] Daily, weekly, monthly, yearly aggregation (sum) of rainfall occurring.
  - [ ] Expand to other metrics
- [ ] Display the above aggregated data on the webui
- [ ] Improve graph choices based on the type of data being displayed, currently all are line graphs.

v2.0:

- [ ] Configurations per sensor are able to be created; these are responsible for determining how the raw data is to be modified and then stored as calibrated data. These are to have "starting" times for all the sensors to be reinstalled.
  - [ ] While for the MVP, I want configurations to affect incoming data, the actual creation of these configurations should be behind authentication.
- [ ] User can recalibrate the data; this will take all the raw data and apply all the calibrations created for the sensor.

## website

Using NodeJS v22.13.0

## server

Using Golang

### Version

~~[Limited by support version from DigitalOcean.](https://docs.digitalocean.com/products/app-platform/reference/buildpacks/go/)
Specifically it is _App Platform uses version 192 of the Heroku Go Buildpack_
[Direct list found here (for version _192_)](https://github.com/heroku/heroku-buildpack-go/blob/v192/data.json)~~

~~[Version selection can be done via](https://go.dev/doc/manage-install)
Latest version is: _1.22.4_~~

https://docs.digitalocean.com/products/app-platform/reference/buildpacks/go/#current-buildpack-version-and-supported-runtimes
Version upgrade:
App Platform uses version __205__ of the Heroku Go Buildpack. The buildpack supports Go runtime versions 1.11 up to 1.24. If no version is specified in your app, App Platform defaults to using version 1.20.

Latest supported Go version:
[1.24.1](https://github.com/heroku/heroku-buildpack-go/blob/v205/data.json)

Primary motivation is that I have done similar in Python multiple times (Flask, Quart) and while I have created microservices within Golang, I have not used it for web API hosting.

#### Gosec

~~Last version compatible with _1.22.4_ should be [v2.21.0](https://github.com/securego/gosec/releases/tag/v2.21.0)~~

Last version compatible with _1.24.1_ should be [v2.22.2](https://github.com/securego/gosec/releases/tag/v2.22.2)

### Cron style aggregation/s

Considerations.

The purpose of this is to take the conjob-style aggregation requests, which will be run periodically to "group"
data pull, ie, sum daily rainfall.

Overall, highly overengineered for the traffic we have, however want to experiment and play with concurrent behaviour
Ideal outcomes would be flat resource usage across App Platform and Mongo. (Again, excessive as App Platform's current average 3-4% at current tier.) Avoid spikes.

1. Cron job periodically triggers (hourly, day, weekly etc) for all the aggregated tasks for that time period.
    * `github.com/robfig/cron/v3` used to achieve this.

2. Tasks are assigned a consistent key which will be paired with redis to lock the aggregation to only be performed by a single application (the purpose is to sync between the number of applications)
    * Achieved this - log seen:

      `2025/02/16 16:00:01 Unable to acquire lock for key RainfallHourly-DAILY-%Y-%m-%d, already claimed (this is expected for multiple applications)`

3. Randomise creation of task to minimise clashes (if the task lists are generated/sent to job queue in the same order )
    * Haven't done this step yet.

## Security Considerations

### Server

- [x] CORS Security - Only production domains or local development URL are allowed
- [x] Currently on POST (or other insertion endpoints are available,) so no current need for:
  - [x] Input validation
  - [x] Authentication - TODO as part of v2
- [x] Encryption (passwords) TODO as part of v2 although would like to allow for 3rd party auth.
~~[x] Rate limiting considered however, given it's not a publicly supplied API (just supplies website), not likely all that beneficial.~~
~~[x] Concurrency limit added,d however doesn't actively deny the connection but rather logging spikes so I can action.~~
- [x] Rate limiting and concurrency limitation added focused on routes which are not supported.
  - This was prompted by some [`interesting`](#interesting-network-traffic) network traffic detected.
  - **TODO** The concurrency limitation does not apply across all instances rather on an instance basis
    - This can be changed to:
      - Use redis to coordinate across instances, and have to consider the tradeoffs.
      - Make further use of the load balancer within DigitalOcean.
- [x] Project scanned with `gosec`.
    > gosec ./... [within `server` directory.]
  - [x] Add GitHub workflow to scan with gosec on `master` branch interactions.
~~- [x] Implemented HSTS header following OWASP guidelines~~
- [ ] TODO: learn more about OWASP + plus general on [blog by UpGuard on HSTS](https://www.upguard.com/blog/hsts)
- [x] Shifted security headers to be controllable by Vercel.
  - Prompted by investigating why Lighthouse (and similar) where indicating the lack of the expected security headers. This is part of an investigation into best practices.

### WebUI

TODO: flesh this out

### General

- [x] HTTPS/TLS is enforced by App Platform with Let's Encrypt certificates tied to my domain bought from Namecheap.

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

### Decoder

When adding the device within TTN there was no need to add the decoder manually as the end device was able to be added from with the lorawan device repository.

## Weather sensor - S2120 

[Purchase Location](https://www.iot-store.com.au/products/sensecap-s2120-lorawn-weather-station)

[Guide](https://cdn.shopify.com/s/files/1/1386/3791/files/SenseCAP_S2120_LoRaWAN_8-in-1_Weather_Station_User_Guide.pdf?v=1662178525)

[Ignore decoder within guide](https://github.com/Seeed-Solution/TTN-Payload-Decoder/blob/master/README.md)

Should have looked here first:
[Online Guide](https://wiki.seeedstudio.com/Getting_Started_with_SenseCAP_S2120_8-in-1_LoRaWAN_Weather_Sensor/)
[Actual decoder](https://github.com/Seeed-Solution/SenseCAP-Decoder/tree/main/S2120/TTN)
Copied for testing purposes within server directory `ttnDecoders`.

### Decoder

The end device was required to be added manually.

By default the format of the decoder looks like:

```json
{
    "err": 0,
    "messages": [
        [
            {
                "measurementId": "4097",
                "measurementValue": 30.8,
                "type": "Air Temperature"
            },
            {
                "measurementId": "4098",
                "measurementValue": 44,
                "type": "Air Humidity"
            },
            {
                "measurementId": "4099",
                "measurementValue": 114488,
                "type": "Light Intensity"
            },
            {
                "measurementId": "4190",
                "measurementValue": 8.8,
                "type": "UV Index"
            },
            {
                "measurementId": "4105",
                "measurementValue": 1.2,
                "type": "Wind Speed"
            }
        ],
        [
            {
                "measurementId": "4104",
                "measurementValue": 54,
                "type": "Wind Direction Sensor"
            },
            {
                "measurementId": "4113",
                "measurementValue": 0,
                "type": "Rain Gauge"
            },
            {
                "measurementId": "4101",
                "measurementValue": 99190,
                "type": "Barometric Pressure"
            }
        ],
        [
            {
                "measurementId": "4191",
                "measurementValue": 3,
                "type": " Peak Wind Gust"
            },
            {
                "measurementId": "4213",
                "measurementValue": 0,
                "type": "Rain Accumulation"
            }
        ]
    ],
    "payload": "4A01342C0001BF3858000C4B00360000000026BF4C001E00000000",
    "valid": true
}
```

The structure of nested messages arrays are somewhat irrigating to parse.
Minor code modification was performed to "flatter" the resulting array.

```json
{
    "err": 0,
    "payload": "4A01342C0001BF3858000C4B00360000000026BF4C001E00000000",
    "valid": true,
    "messages": [
      {
        "measurementValue": 30.8,
        "measurementId": "4097",
        "type": "Air Temperature"
      },
      {
        "measurementValue": 44,
        "measurementId": "4098",
        "type": "Air Humidity"
      },
      {
        "measurementValue": 114488,
        "measurementId": "4099",
        "type": "Light Intensity"
      },
      {
        "measurementValue": 8.8,
        "measurementId": "4190",
        "type": "UV Index"
      },
      {
        "measurementValue": 1.2,
        "measurementId": "4105",
        "type": "Wind Speed"
      },
      {
        "measurementValue": 54,
        "measurementId": "4104",
        "type": "Wind Direction Sensor"
      },
      {
        "measurementValue": 0,
        "measurementId": "4113",
        "type": "Rain Gauge"
      },
      {
        "measurementValue": 99190,
        "measurementId": "4101",
        "type": "Barometric Pressure"
      },
      {
        "measurementValue": 3,
        "measurementId": "4191",
        "type": " Peak Wind Gust"
      },
      {
        "measurementValue": 0,
        "measurementId": "4213",
        "type": "Rain Accumulation"
      }
    ]
}
```

This allows for simplied parsing without all loss of useful information for my purposes.
(I do not benefit for the separation of data collection from each internal 8-in-1 sensor)

## Mongo

### Indices
Both `RawData` and `CalibratedData` schemas within MongoDB are configured as timeseries collections.
Both use `sensor` and `timestamp` as compound indexes and are intended to function as covering indexes for all associated queries.

### mongosh

View timeseries information

```mongosh
db.runCommand({
  listCollections: 1,
  filter: { name: "raw_data" }
})
```

```mongosh
db.runCommand({
  listCollections: 1,
  filter: { name: "raw_data" }
}).cursor.firstBatch[0]
```

```mongosh
db.runCommand({
  listCollections: 1,
  filter: { name: "calibrated_data" }
})
```

```mongosh
db.runCommand({
  listCollections: 1,
  filter: { name: "calibrated_data" }
}).cursor.firstBatch[0]
```

## Redis

Purpose is current limited to sync IO tasks between multiple instances of the server. 

## Logging

DigitalOcean connected `Papertrail`

# Hosting

## WebUI

Hosted on: <b>[vercel](https://vercel.com)</b>

Vercel CLI is used to deploy when required.

### Environment variables

Once updated via the Vercel dashboard, it is important to pull them locally.

This will pull the "production" environment fields to test local development against the production server.
> vercel env pull --environment=production .env.production

From here can use the Vercel deploy steps within the `package.json` file.

## Server

built in: [gin-gonic](https://gin-gonic.com/)

Hosted on: <b>DigitalOcean</b>

Domain established within DigitalOcean, directing to droplet:
`mini-farm-tracker.io`

DNS Records are configured within DigitalOcean to allow for vercel website to be sued.

## Weather

<https://openweathermap.org/price>

TODO: haven't done anything with this currently.

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

##### Build a run locally (windows)
go build ; if ($?) { .\mini-farm-tracker-server.exe }

#### Network

Firewall options - inbound port of 3000 (TCP) required

SSL certificate (Let's Encrypt) created on the domain bought from namecheap.

### Testing

#### testContainer (currently only implemented for server - MongoDB)

NOTE: testContainer can use cloud resources, however prefer to run locally.

#### TODO

#### Test result generation

##### Server

> go test ./... -coverprofile=coverage.out
> go tool cover -html coverage -o coverage.html

Requires:
- Docker Desktop
- [testContainer](https://app.testcontainers.cloud/accounts/14403/dashboard/install?target=windows-desktop)

With my environment, I have problems with the embedded testContainers cleanup logic.

Following [configuration path](https://golang.testcontainers.org/features/configuration/), adding a line to disable `ryuk` allows correct running:

> ryuk.disabled=true

## TODOs

### WebUI

- [x] Graphs
  - Initial HW will allow for 2 sensors, one for each tank.
- [ ] V2 will have auth, although as part of the purpose of this is a demo project, putting it behind an auth "wall" is counterproductive initially.

### Server

- [x] Investigate HTTP servers - SSL secured and CORS established
  - Currently implemented with [gin](https://github.com/gin-gonic/gin)
- [x] Investigate Containerisation options.
  - The initial version is simply running binary.
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
  - [x] Schema definitions - Primary outlined in the `schema.go` file.
    - [ ] Not all defined - enough for E2E functionality to be possible

- [x] Connect The Things Stack
  - [x] Account Created
  - Hardware provided
  - [x] Gateway
  - [x] 2x ultrasonic sensors to measure depth in tanks.

## Misc.

<a id="interesting-network-traffic"></a>
### Interesting network traffic

1. Various attempts to grab the project

  Feb 20 10:39:32 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 18:39:30 | 404 |      51.638µs | 206.221.176.253 | GET      "/backup.rar"
  
  Feb 20 10:39:32 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 18:39:30 | 404 |      47.694µs | 206.221.176.253 | GET      "/site.zip"
  
  Feb 20 10:39:32 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 18:39:30 | 404 |      44.674µs | 206.221.176.253 | GET      "/backup.zip"
  
  Feb 20 10:39:32 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 18:39:30 | 404 |      23.071µs | 206.221.176.253 | GET      "/api_mini-farm-tracker_io.rar"
  
  Feb 20 10:39:32 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 18:39:30 | 404 |      28.328µs | 206.221.176.253 | GET      "/website.zip"
  
  Feb 20 10:39:32 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 18:39:30 | 404 |      18.078µs | 206.221.176.253 | GET      "/api_mini-farm-tracker_io.zip"
  
  Feb 20 10:39:32 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 18:39:30 | 404 |      31.558µs | 206.221.176.253 | GET      "/site.rar"
  
  Feb 20 10:39:32 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 18:39:30 | 404 |      40.026µs | 206.221.176.253 | GET      "/api.mini-farm-tracker.io.zip"
  
  Feb 20 10:39:32 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 18:39:30 | 404 |      16.631µs | 206.221.176.253 | GET      "/website.rar"
  
  Feb 20 10:39:32 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 18:39:30 | 404 |      54.945µs | 206.221.176.253 | GET      "/apimini-farm-trackerio.rar"
  
  Feb 20 10:39:32 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 18:39:30 | 404 |      51.673µs | 206.221.176.253 | GET      "/apimini-farm-trackerio.zip"
  
  Feb 20 10:39:32 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 18:39:30 | 404 |      38.742µs | 206.221.176.253 | GET      "/api.mini-farm-tracker.io.rar"

2. Various attempts to query for version control content
  
  Feb 20 05:46:17 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 13:46:15 | 404 |      61.282µs |    148.66.1.242 | GET      "/.git/HEAD"
  
  Feb 20 05:46:17 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 13:46:16 | 404 |      62.041µs |    148.66.1.242 | GET      "/.git/config"
  
  Feb 20 05:46:18 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 13:46:17 | 404 |      56.961µs |    148.66.1.242 | GET      "/.svn/entries"
  
  Feb 20 05:46:18 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 13:46:17 | 404 |      47.556µs |    148.66.1.242 | GET      "/.svn/wc.db"
  
  Feb 20 07:48:07 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 15:48:05 | 404 |      65.983µs |    148.66.1.242 | GET      "/.git/HEAD"
  
  Feb 20 07:48:08 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 15:48:06 | 404 |      55.219µs |    148.66.1.242 | GET      "/.git/config"
  
  Feb 20 07:48:08 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 15:48:06 | 404 |      74.434µs |    148.66.1.242 | GET      "/.svn/entries"
  
  Feb 20 07:48:08 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 15:48:07 | 404 |      65.825µs |    148.66.1.242 | GET      "/.svn/wc.db"
  
  Feb 20 09:00:48 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 17:00:46 | 404 |      63.688µs |    148.66.1.242 | GET      "/.git/HEAD"
  
  Feb 20 09:00:48 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 17:00:47 | 404 |       43.38µs |    148.66.1.242 | GET      "/.svn/entries"
  
  Feb 20 09:00:48 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 17:00:46 | 404 |      51.889µs |    148.66.1.242 | GET      "/.git/config"
  
  Feb 20 09:00:48 shark-app mini-farm-tracker-server [GIN] 2025/02/20 - 17:00:47 | 404 |      50.416µs |    148.66.1.242 | GET      "/.svn/wc.db"
