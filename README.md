# sion

Sion is a scalable, distributed file system designed to reliably store and serve millions of files to thousands of clients. It is designed to support files that are terabytes in size by distributing the workload over many machines with heterogeneous hardware allocations. The storage capacity and aggregate bandwidth of the cluster scales with the number and size of machines and disks that participate in the system.

Sion does not expose a traditional POSIX file system API to clients. Instead, it exposes an HTTP+JSON API with which all namespace and chunk operations are performed. 

## Architecture
A Sion cluster contains a primary master which directs the state of the cluster, and many storage nodes which store and serve files.

### Master
A Sion master performs three major functions: cluster management, namespace management, and placement.
1. The cluster subsystem processes cluster operations and tracks the state of the cluster.
2. The namespace subsystem processes file system operations and serves this information to storage nodes and clients.
3. The placement subsystem makes assignment decisions for where chunks should be located. Placement decisions are used to respond to failures in the cluster and balance load across nodes. It interacts with the cluster and namespace subsystems to perform necessary functions.
