# mini-farm-tracker

Basic overview:
Provide a visualisation platform for various LoRaWAN sensors.
Initial design is required to track available water in water tanks.

Current Implementation

## website

Using NodeJS v22.13.0

Hosted on: <b>[vercel](https://vercel.com)</b>

vercel CLI is used to deploy when required.

#### Environment variables
Once updated via the vercel dashboard, it is important to pull them locally.

> vercel env pull

From here can use the vercel deploy steps within the `package.json` file. 


## server

Using Golang

Primary motivation is I have done similar in Python multiple times (Flask, Quart) and while I have created microservices within Golang, I have not used it for web API hosting.

### Hosting/Deployment Options

Hosted on: <b>DigitalOcean</b>

Domain established within DigitalOcean directing to droplet:
`mini-farm-tracker.io`

DNS Records are configured within DigitalOcean to allow for vercel website to be sued.

#### Development

> git checkout .
> git clean -fd

> go build
> export GIN_MODE=release && ./mini-farm-tracker-server

##### Network

Firewall options - inbound port of 3000 (TCP) required

SSL certificate (Let's Encrypt) created on domain bought from namecheap.

## TODO:

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
