Monitor an ACI Fabric from your favorite Webex Room - ChatOps
=====================================================================================

Monitor your Data Center using an user-friendly bot. This repository contains a [Go](https://go.dev/)-based application that spins up a web server, which listens to [Webex](https://www.webex.com/) [webhooks](https://developer.webex.com/docs/api/guides/webhooks) notifications and interacts with the [APIC](https://www.cisco.com/c/en/us/products/cloud-systems-management/application-policy-infrastructure-controller-apic/index.html) REST API to retrieve information about the operational status of your [ACI](https://www.cisco.com/c/en/us/solutions/data-center-virtualization/application-centric-infrastructure/index.html) Fabric.

<p align="center">
<img src="./docs/images/aci-chatbot.png" border="0" alt="aci-chatbot">
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

## Related Sandbox
* [ACI Simulator 5.2 ](https://devnetsandbox.cisco.com/RM/Diagram/Index/4eaa9878-3e74-4105-b26a-bd83eeaa6cd9?diagramType=Topology)
* [ACI Simulator AlwaysOn](https://devnetsandbox.cisco.com/RM/Diagram/Index/18a514e8-21d4-4c29-96b2-e3c16b1ee62e?diagramType=Topology)

## Links to DevNet Learning Labs
* [ACI Programmability](https://developer.cisco.com/learning/tracks/aci-programmability)
* [Get Started with Webex APIs](https://developer.cisco.com/learning/tracks/collab-cloud)
