# Monitor an ACI Fabric from your favorite Webex Room - ChatOps

[![Tests](https://github.com/jgomezve/aci-chatbot/actions/workflows/test.yml/badge.svg)](https://github.com/jgomezve/aci-chatbot/actions/workflows/test.yml)


A lightweight bot application to interact with the APIC from a Webex chat room

## Use Case description

Monitor your Data Center using an user-friendly bot. This repository contains a [Go](https://go.dev/)-based application that spins up a web server, which listens to [Webex](https://www.webex.com/) [webhooks](https://developer.webex.com/docs/api/guides/webhooks) notifications and interacts with the [APIC](https://www.cisco.com/c/en/us/products/cloud-systems-management/application-policy-infrastructure-controller-apic/index.html) REST API to retrieve information about the operational status of your [ACI](https://www.cisco.com/c/en/us/solutions/data-center-virtualization/application-centric-infrastructure/index.html) Fabric.

<p align="center">
<img src="docs/images/aci-chatbot.png" border="0" alt="aci-chatbot">
<br/>

This application allows you retrieve operational, topology, event/fault, endpoint information from the ACI Fabric by simply typing short and human-readable command in a Webex room. These is the list of the currently supported commands by the aci-chatbot:

```
•	/cpu	->	Get APIC CPU Information 💾
•	/ep	->	Get APIC Endpoint Information 💻. Usage /ep [ep_mac] 
•	/events	->	Get Fabric latest events ❎.   Usage /events [user:opt] [count(1-10):opt] 
•	/faults	->	Get Fabric latest faults ⚠️. Usage /faults [count(1-10):opt] 
•	/help	->	Chatbot Help ❔
•	/info	->	Get Fabric Information ℹ️
•	/neigh	->	Get Fabric Topology Information 🔢. Usage /neigh [node_id] 
•	/websocket	->	Subscribe to Fabric events 📩
```

The bot takes advantage of the [APIC REST API](https://www.cisco.com/c/en/us/td/docs/switches/datacenter/aci/apic/sw/2-x/rest_cfg/2_1_x/b_Cisco_APIC_REST_API_Configuration_Guide/b_Cisco_APIC_REST_API_Configuration_Guide_chapter_01.html#d54e540a1635) to query and filter information from the APIC Management Information Tree (MIT). Furthermore, the `/websocket` command leverages the [APIC WebSocket](https://www.cisco.com/c/en/us/td/docs/switches/datacenter/aci/apic/sw/2-x/rest_cfg/2_1_x/b_Cisco_APIC_REST_API_Configuration_Guide/b_Cisco_APIC_REST_API_Configuration_Guide_chapter_01.html#concept_71EBE2E241C3442BA326273AF1A9B617) functionality, to get instant notifications once any instance of a defined MO/Class is created, modified or deleted.


## Prerequisites

* Make sure to have Go 1.15+ or Docker installed on your computer/server

    * [Install Go](https://go.dev/doc/install)
    * [Install Docker](https://docs.docker.com/get-docker/)

* Login to your Webex account and create your own bot. [Create Bot](https://developer.webex.com/docs/bots)

    1. Give your bot details

        ![add-app](docs/images/bot_details.png "Create Bot") 
    
    2. The generate token is your `WEBEX_TOKEN`

        ![add-app](docs/images/bot_token.png "Bot Token")

## Installation

### Execute ngrok (Optional)

The bot application must be hosted in a server reachable via the public internet, because the webhooks are delivered from webex.com. For development and testing pursposes you could use [ngrok](https://ngrok.com/) to expose your server to the public internet. Ngrok will expose your application (Server IP & Port) over a secure tunnel.


![Ngrok Diagram](docs/images/chatbot_ngrok.png "Ngrok Diagram")

* [Install ngrok](https://ngrok.com/download)

Follow these instruction after installing ngrok:

* Start the ngrok service stating the port the bot server listens to. By default the application listens to por **7001**, however if you are using the docker container you should state here Docker host port.

        ./ngrok http <bot_port> --region=eu

```
ngrok by @inconshreveable    

Session Status                online
Session Expires               1 hour, 59 minutes
Version                       2.3.40
Region                        Europe (eu)
Web Interface                 http://127.0.0.1:7001
Forwarding                    http://2d6e-89-246-96-47.eu.ngrok.io -> http://localhost:7001
Forwarding                    https://2d6e-89-246-96-47.eu.ngrok.io -> http://localhost:7001

Connections                   ttl     opn     rt1     rt5     p50     p90
                              0       0       0.00    0.00    0.00    0.00
```

*  The generated HTTP url is your `BOT_URL`

> **_NOTE:_**:  The trial version of ngrok creates the secure tunnel only for 2 hours

### Option 1: Build the code from source

* Set and source the environmental variables in `env.sh`

```
export WEBEX_TOKEN=YOUR-WEBEX-TOKEN-GOES-HERE
export BOT_URL=http://c314-173-38-220-48.eu.ngrok.io
export APIC_URL=https://sandboxapicdc.cisco.com/
export APIC_USERNAME=admin
export APIC_PASSWORD=admin
```
        source env.sh

* Execute the application

        go run main.go

> **_NOTE:_**:  The application listens to port `7001`

### Option 2: Execute the service as a Container

* Set the environmental variables in `.env`:

```
WEBEX_TOKEN=herehoesyourbotwebextoken
BOT_URL=http://2258-173-38-220-34.eu.ngrok.io
APIC_URL=https://sandboxapicdc.cisco.com/
APIC_USERNAME=admin
APIC_PASSWORD=admin
```

*  Run the application in a Docker container

            docker run --env-file .env -it -p <bot_port>:7001 jgomezve/aci-chatbot:latest

> **_NOTE:_** In case you are using ngrok, <bot_port> is the same port used to start ngrok.

## Usage

Either send a message directly to your bot or add it to a Webex Group

![add-app](docs/images/webex_message.png "Bot Message")

> **_NOTE:_** Some commands do not work if the target APIC is a simulator