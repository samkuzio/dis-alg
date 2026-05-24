# Architecture Specification: DIS Application Layer Gateway (ALG)

## 1. Overview
The DIS ALG is a Go-based hub-and-spoke relay designed to bridge DIS traffic across non-broadcast capable networks (VPNs/Cloud). The system prioritizes low-latency relay, scalability to multiple sites, and extensibility for future DIS-aware filtering/routing logic.

## 2. High-Level Architecture
- **Hub (Central Node):** A centralized process receiving TCP streams from all Spokes. It maintains a registry of active spokes and implements broadcast-to-all logic (fan-out).
- **Spokes (Terminal Nodes):** Local relays at each site that listen to local UDP broadcast and tunnel packets to the Hub over TCP.

## 3. Communication Protocol
- **Transport:** Persistent TCP streams per Spoke (abstraction interface provided for future protocols).
- **Framing:** Custom DIS-ALG header wrapping raw DIS PDUs.

## 4. Key Architectural Decisions (Finalized)

### A. Reliability vs. Latency (Transport Abstraction)
- **Status:** TCP chosen for PoC. 
- **Requirement:** The architecture must use an interface-based transport abstraction (e.g., `type Transport interface` defined in `/pkg/core`) to allow future swapping for UDP/QUIC/RTP with custom reliability layers without refactoring the ALG core.

### B. Serialization and Inspection
- **Status:** Header-based routing only.
- **Requirement:** The Hub will not decode the full DIS PDU. Instead, all traffic will be wrapped in a **DIS-ALG Custom Header**. 
- **Header Format:**
    - `[uint32: SourceID]` (Unique identifier provided by the Spoke on connection)
    - `[uint64: PacketNumber]` (Sequence number for monitoring)
    - `[uint32: PayloadLength]` 
    - `[bytes: Original DIS PDU]`

### C. Flow Control
- **Status:** Passive monitoring.
- **Requirement:** The Hub will perform "best-effort" delivery. It will continue to flood packets to all connected Spokes. The custom `PacketNumber` in the ALG header allows terminal nodes to track sequence gaps and diagnose packet loss/latency issues locally.

### D. Security
- **Status:** Unsecured for PoC.
- **Requirement:** No authentication or TLS implementation for the PoC. Connections are treated as trusted endpoints.

### E. Observability
- **Status:** Structured logging required.
- **Requirement:** The system must use Go's standard `log/slog` for structured JSON or Key-Value logging. Contextual metadata (e.g., `SpokeID`, `PacketNum`, `Bytes`) must be explicitly attached to facilitate debugging.

---

## 5. Go Component Structure & AI Implementation Strategy

The codebase is organized to support Contract-First Development and strict isolation of concerns, making unit testing and AI-driven implementation highly effective.

```text
/cmd/hub            - main.go (CLI flags, dependency wiring, starts Hub server)
/cmd/spoke          - main.go (CLI flags, dependency wiring, starts Spoke client)
/pkg/core           - Contract definitions (Transport and Router interfaces) and shared domain types
/pkg/protocol       - Custom DIS-ALG header struct, byte framing, serialization/deserialization logic (TDD mandatory)
/pkg/transport/tcp  - TCP implementation of the core.Transport interface
/pkg/hub            - Registry, fan-out routing logic, and best-effort delivery loop
/pkg/spoke          - UDP listener, UDP broadcaster, and TCP tunnel forwarder
```
