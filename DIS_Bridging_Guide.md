# DIS Protocol Bridging Guide (UDP to TCP)

The Distributed Interactive Simulation (DIS) protocol typically relies on UDP broadcast or multicast. To transmit this traffic over a TCP tunnel (which is unicast and stream-based), you need a relay/proxy.

## Bridging DIS Traffic: The Bi-directional Loop Prevention Strategy

To achieve bi-directional bridging while preventing packet loops (the "echo" effect), you must bind `socat` to specific network interfaces rather than using the generic `0.0.0.0` or broadcast addresses.

### 1. Architectural Setup
Assume:
- **Interface A** (Internal): The local interface (e.g., `eth0`) where the simulation components live.
- **Interface B** (Tunnel): The interface connecting to the VPN/Remote link (e.g., `tun0` or `wg0`).

### 2. The Bi-directional `socat` Command
On each node, run two separate `socat` processes. This separation of concerns ensures that traffic received from the local interface is forwarded to the tunnel, and traffic arriving from the tunnel is forwarded to the local interface, without overlap.

**Process 1: Local -> Tunnel**
Listens on local UDP port, sends over TCP to the remote bridge.
```bash
# Bind ONLY to your local simulation interface.
socat UDP4-RECVFROM:3000,bind=192.168.1.1,fork TCP4:remote-bridge-ip:4000
```

**Process 2: Tunnel -> Local**
Listens on TCP port (from the tunnel), broadcasts to local simulation subnet.
```bash
# Bind to the local interface IP, preventing packet loops on the tunnel interface.
socat TCP4-LISTEN:4000,bind=192.168.1.1,fork UDP4-DATAGRAM:192.168.1.255:3000,broadcast
```

### 3. Loop Prevention
*   **Interface Binding (`bind=...`):** By explicitly binding to the IP of your simulation network interface, you guarantee `socat` never interacts with the traffic on the tunnel interface's local broadcast segment (or vice versa).
*   **Subnet Partitioning:** Ensure your local simulation IP range (e.g., `192.168.1.x`) is logically distinct from the VPN link range.
*   **MTU Considerations:** DIS packets are often small, but ensure the TCP tunnel MTU can handle the encapsulated UDP frame plus overhead.
