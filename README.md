# aci-chatbot

A lightweight bot application to interact with the APIC from a Webex Chat romm

## Use Case description

Monitor your Data Center using an user-friendly bot. This repository contains a [Go](https://go.dev/)-based application that spins up a web server, which listens to [Webex](https://www.webex.com/) webhooks events an interacts with the [APIC](https://www.cisco.com/c/en/us/products/cloud-systems-management/application-policy-infrastructure-controller-apic/index.html) REST API to retrieve information about the operational status of your Fabric.

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

Make sure to have Golang 1.15+ or Docker installed on your computer/server

## Installation

### Option 1: Build the code from source

### Option 2: Execute the service as a Container

* Set the environmental variables in `.env`:

```
WEBEX_TOKEN=herehoesyourbotwebextoken
BOT_URL=http://2258-173-38-220-34.eu.ngrok.io
APIC_URL=https://192.168.1.1
APIC_USERNAME=admin
APIC_PASSWORD=admin
```

*  Run the service as a Docker container

            docker run --env-file .env -it -p 8080:7001 jgomezve/aci-chatbot:latest

## Usage

Find the bot you create in the Webex Application and start asking hi about your ACI Fabric :) 