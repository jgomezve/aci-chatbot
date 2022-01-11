# aci-chatbot

A [Go](https://go.dev/)-based Bot which reads information from the APIC and displays it in a 1:1 Webex Room

## How to use it?

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


## Supported Webex Commands

```
/cpu->Get APIC CPU Information
/ep->Get APIC Endpoint Information. Usage /ep <ep_mac>
/help->Chatbot Help
```