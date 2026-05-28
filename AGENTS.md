# DIS Application Layer Gateway - Agent Instructions

This project is a DIS (Distributed Interactive Simulation) application layer gateway. It acts as a hub-and-spoke relay designed to bridge DIS traffic across non-broadcast capable networks (e.g., VPNs/Cloud). 

## Domain Information
- **Hub:** Centralized process receiving TCP streams from terminal/spoke nodes. Performs fan-out (broadcast-to-all).
- **Terminal (Spoke):** Local relays at each site that listen to local UDP broadcast and tunnel packets to the Hub over TCP.
- **Protocol:** Wraps raw DIS PDUs in a custom DIS-ALG Header (SourceID, PacketNumber, PayloadLength, PDU) over persistent TCP connections. Unsecured (no TLS/Auth for PoC). Best-effort delivery.
- **Logging:** Must use Go's standard `log/slog` for structured logging with contextual metadata (SpokeID, PacketNum, Bytes).

## Architecture and Documentation
Relevant documentation is available in the `doc/` folder of this repository:
- `DIS_ALG_Architecture.md`: Details the transport abstraction, header format, flow control, and component structure.
- `DIS_Bridging_Guide.md`: Outlines the underlying network theory (UDP-to-TCP bridging and loop prevention).

*Note: You should consult this documentation when planning features. This documentation is the source of truth, but it is malleable; it can be changed as necessary. Always consult with a human user before making changes to architectural or design decisions documented in these files.*

## Build and Test Commands

**Build the statically linked project (outside Docker):**
```bash
GOTMPDIR=/workspace/gotmp CGO_ENABLED=0 go build -ldflags="-w -s" -o dis-alg ./cmd/dis-alg/
```

**Run tests:**
```bash
# Standard test execution
go test ./...

# If you encounter permission denied errors for /tmp/go-build* (common in some Docker/sandbox environments), 
# override GOTMPDIR to use a local temporary directory:
mkdir -p gotmp && GOTMPDIR=$(pwd)/gotmp go test ./...
```
*(Tests should be written using standard Go testing practices. TDD is mandatory for protocol serialization/deserialization logic.)*

**Run Hub Mode:**
```bash
./dis-alg hub [transport] [bind-ip]:[port]
# e.g., ./dis-alg hub tcp 0.0.0.0:8080
```

**Run Terminal Mode:**
```bash
./dis-alg terminal [--simulation|-s [sim-ip]:[sim-port]] [--hub|-h [hub-ip]:[hub-port]] [--transport|-t [transport]]
```

* **--simulation, -s**: The IP and Port on which the terminal node will listen to simulation traffic. Always assumed to be UDP. The IP may be a broadcast, multicast, or unicast IP.
* **--hub, -h**: The IP and port on which the terminal node will connect to the hub.
* **--transport, -t**: The transport protocol that will be used to connect to the hub, currently the only supported value is `tcp`.

## Documentation Standards
- **Public Functions:** All public functions must have a documentation string that explains what they do and return, unless they are trivial one-line accessor functions.
- **Constants and Global Variables:** All constants and global variables must have a documentation comment explaining their purpose and use.
- **Struct Fields:** All struct fields must have comments describing their purpose.

## AI Agent Guidelines
- Use contract-first development. Favor interface-based transport abstractions (e.g., `Transport` interface) for core functions.
- Write tests alongside new framing and serialization code.
- Prefer Go standard library (`log/slog`, `net`) wherever possible.
