# sion

Sion is a scalable, distributed file system designed to reliably store and serve millions of files to thousands of clients. It is designed to support files that are terabytes in size by distributing the workload over many machines with heterogeneous hardware allocations. The storage capacity and aggregate bandwidth of the cluster scales with the number and size of machines and disks that participate in the system.

Sion does not expose a traditional POSIX file system API to clients. Instead, it exposes an HTTP+JSON API with which all namespace and chunk operations are performed. 

## Architecture
1. The cluster subsystem is responsible for processing cluster operations and tracking the state of the cluster.
2. The namespace subsystem is responsible for maintaining the directory structure and metadata of the file system.
3. The placement subsystem is responsible for making assignment decisions around where chunks should be located. It responds to the cluster and namespace subsystems to perform necessary functions.
