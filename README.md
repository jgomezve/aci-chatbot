# Monitor an ACI Fabric from your favorite Webex Room - ChatOps

A lightweight bot application to interact with the APIC from a Webex Chat room

## Use Case description

Monitor your Data Center using an user-friendly bot. This repository contains a [Go](https://go.dev/)-based application that spins up a web server, which listens to [Webex](https://www.webex.com/) webhooks events an interacts with the [APIC](https://www.cisco.com/c/en/us/products/cloud-systems-management/application-policy-infrastructure-controller-apic/index.html) REST API to retrieve information about the operational status of your Fabric.

<p align="center">
<img src="./docs/images/aci-chatbot.png" border="0" alt="aci-chatbot">
<br/>

This is the list of supported commands.

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

Even though most of the commands 

## Prerequisites

* Make sure to have Go 1.15+ or Docker installed on your computer/server

    * [Install Go](https://go.dev/doc/install)
    * [Install Docker](https://docs.docker.com/get-docker/)

* Login to your Webex account and create your own bot. [Create Bot](https://developer.webex.com/docs/bots)

    1. Give your bot details

        ![add-app](./docs/images/bot_details.png)
    
    2. Copy the generated bot token, it will be required later

        ![add-app](./docs/images/bot_token.png)

## Installation

### Execute ngrok (Optional)

<p align="center">
<img src="./docs/images/aci-chatbot.png" border="0" alt="aci-chatbot_ngrok">
<br/>


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

            docker run --env-file .env -it -p 8080:7001 jgomezve/aci-chatbot:latest

> **_NOTE:_** Be careful with the port forwarding when using ngrok. The web server listens to the port 7001

## Usage

Either send a message directly to your bot or add it to a Webex Group

![add-app](./docs/images/webex_message.png)

> **_NOTE:_** Some commands do not work if the target APIC is a simulator