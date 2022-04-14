package metadata

// Handles all placement decisioning

type placement struct {
	chunkAssignments map[ChunkId][]NodeId
	nodeAssignments  map[NodeId][]ChunkId
}
