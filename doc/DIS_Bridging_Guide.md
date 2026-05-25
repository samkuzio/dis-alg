# DIS Protocol Bridging Guide (UDP to TCP)

The Distributed Interactive Simulation (DIS) protocol typically relies on UDP broadcast or multicast. To transmit this traffic over a non-broadcast capable network (such as a VPN, cloud environment, or WAN), a bridging strategy is required. 

This project implements a custom **Hub-and-Spoke Application Layer Gateway (ALG)** to tunnel DIS traffic over persistent TCP streams.

## 1. Hub-and-Spoke Architecture

Unlike point-to-point bridging systems, this system uses a centralized Hub to fan-out traffic to multiple terminal nodes (Spokes).

*   **Hub:** The central routing node. It binds to a designated TCP port and waits for Spokes to connect. When it receives a packet from a Spoke, it relays that packet to all *other* connected Spokes.
*   **Spoke (Terminal):** Placed within each localized simulation network. It opens a TCP tunnel to the Hub. Locally, it binds to a UDP port to listen for DIS traffic and broadcasts received TCP traffic out to its local UDP subnet.

## 2. Network Interface Binding

Proper binding is crucial for the application to function without intercepting unrelated traffic.

### Hub Binding
The Hub only deals with TCP traffic. It binds to a reachable IP address and port (e.g., `0.0.0.0:8080`) to accept incoming TCP connections from the remote Spokes.

### Spoke Binding
The Spoke must manage both TCP and UDP interfaces:
*   **TCP Client:** Initiates an outbound connection to the Hub's reachable IP.
*   **UDP Listener (Ingress to Tunnel):** Binds to the local simulation network interface (e.g., `192.168.1.1:3000` or `0.0.0.0:3000`) to capture local UDP DIS broadcasts.
*   **UDP Broadcaster (Egress from Tunnel):** Sends DIS packets received from the Hub to the local subnet broadcast address (e.g., `192.168.1.255:3000`).

## 3. Loop Prevention: The "Echo" Effect

When bridging UDP broadcast networks, the most critical issue is preventing "packet loops" or the "echo" effect. An echo occurs when a Spoke receives a packet from the Hub, broadcasts it locally, and its own UDP listener immediately picks it up and sends it *back* to the Hub.

To prevent this, the architecture relies on the following strategies:

### A. Hub-Level Fan-out Logic (Split Horizon)
The Hub acts as a split-horizon router. When the Hub receives a packet via a specific TCP stream connection, it fans the packet out to all *other* connected Spokes, but **never** sends it back to the Spoke that originated it. This prevents the immediate network loop over the tunnel.

### B. Spoke-Level Echo Cancellation (Local Loopback Prevention)
Because the Spoke is broadcasting to the same local subnet that it is listening to, its own UDP listener will "hear" the packets it just broadcasted. The Spoke must discard these self-generated broadcasts before they are forwarded over the TCP tunnel.
*   **Cache/Hash Filtering:** The Spoke maintains a brief cache (or hash map) of recently broadcasted DIS packets. If the UDP listener receives a packet that matches an item in the recent broadcast cache, it drops it.
*   **DIS PDU Inspection:** Alternatively, the Spoke can be configured to inspect the raw DIS PDU's `Site ID` or `Application ID`. By identifying remote vs. local simulation entities, the Spoke can easily discard remote packets it just broadcasted locally, preventing them from being tunneled back to the Hub.

### C. Interface Separation
When configuring the Spoke on a host with multiple network interfaces (e.g., a tunnel interface and a local LAN interface), explicitly bind the Spoke's UDP listener to the local simulation interface (e.g., `192.168.1.1`) rather than a generic address like `0.0.0.0`. This ensures clear boundaries between tunnel bridging and local simulation traffic.
