# mini-farm-tracker

Basic overview:
Provide a visualisation platform for various LoRaWAN sensors.
Initial design is required to track available water in water tanks.


## website

Using NodeJS v22.13.0

Default 

### Hosting/Deployment Options

#### [vercel](https://vercel.com)

#### [Github Pages](https://pages.github.com/)


## server

Using Golang

Primary motivation is I have done similar in Python multiple times (Flask, Quart) and while I have created microservices within Golang, I have not used it for web API hosting.

### Hosting/Deployment Options

#### DigitalOcean

Domain established within DigitalOcean directing to droplet:
mini-farm-tracker.io

##### Network

Firewall options - inbound port of 3000 (TCP) required

SSL certificate (Let's Encrypt) created on domain bought from namecheap.