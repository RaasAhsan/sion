# sion

Sion is a scalable, distributed file system designed to reliably store and serve millions of files to thousands of clients. It is designed to support files that are terabytes in size by distributing the workload over many machines with heterogeneous hardware allocations. The storage capacity and aggregate bandwidth of the cluster scales with the number and size of machines and disks that participate in the system.

Sion does not expose a traditional POSIX file system API to clients. Instead, it exposes an HTTP+JSON API with which all namespace and chunk operations are performed. 
