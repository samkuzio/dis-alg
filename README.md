# dis-alg

`dis-alg` is a bridging utility designed to route Distributed Interactive Simulation (DIS) network traffic across non-broadcast capable networks (such as WANs, VPNs, or cloud environments). Because the DIS protocol typically relies on local UDP broadcast or multicast, its traffic is often dropped or unroutable over wide area networks. `dis-alg` solves this by utilizing a Hub and Spoke architecture: Terminal nodes (spokes) sit on the local simulation subnets to capture local UDP DIS broadcasts, tunneling them over TCP to a central Hub node. The Hub then replicates and routes the traffic out to all other connected Terminal nodes, bridging the disparate simulation networks.

`dis-alg` operates in either hub or terminal mode.

## Hub Mode

The hub mode is launched with the following command:

```bash
dis-alg hub [-v|--verbose] [transport] [bind-ip]:[port]
```

### Arguments

* **-v, --verbose**: Enable verbose debug logging.
* **transport**: The transport protocol used to communicate between the hub and terminal nodes. Currently, the only valid value is `tcp`.
* **bind-ip**: The IP address of the socket on which the hub will listen for incoming connections.
* **port**: The port on which the hub will listen for new connections.

## Terminal Mode

Terminal mode is launched with the following command:

```bash
./dis-alg terminal [--simulation|-s [sim-ip]:[sim-port]] [--hub|-h [hub-ip]:[hub-port]] [--transport|-t [transport]] [--verbose|-v]
```

### Arguments

* **--simulation, -s**: The IP and Port on which the terminal node will listen to simulation traffic. Always assumed to be UDP. The IP may be a broadcast, multicast, or unicast IP.
* **--hub, -h**: The IP and port on which the terminal node will connect to the hub.
* **--transport, -t**: The transport protocol that will be used to connect to the hub, currently the only supported value is `tcp`.
* **--verbose, -v**: Enable verbose debug logging.
