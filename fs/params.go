package fs

import "time"

const BufferSize = 256
const ChunkSize = 8 * 1024 * 1024

const NodeTimeout = 1 * time.Minute

const MaxShortAppendLength = 32 * 1024 * 1024
