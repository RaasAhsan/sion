package util

import "os"

// Constraints: checkpoint/snapshot/state can fit into memory, journal cannot

// TODO: either we use a lock to protect access to file,
// or we transmit requests to a goroutine worker which processes
// the writes and returns acknowledgements
type Journal[S any, T any] struct {
	file    *os.File
	handler Handler[S, T]
}

// TODO: should we assume the checkpoint/snapshot can fit into memory?
type Handler[S any, T any] interface {
	InitialState() S
	// All these Read/Write could just use json.Marshal and Unmarshal
	ReadState(contents string) (S, error)
	WriteState(state S) string
	ReadEntry(line string) (T, error)
	WriteEntry(entry T) string
	Replay(state S, entry T) S
}

func Open[S any, T any](filename string, h Handler[S, T]) (*Journal[S, T], error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	journal := &Journal[S, T]{
		file:    file,
		handler: h,
	}
	return journal, nil
}

func OpenFromSnapshot[S any, T any](log_filename string, snapshot_filename string, h Handler[S, T]) (*Journal[S, T], error) {
	return nil, nil
}

func (j *Journal[S, T]) Replay() (S, error) {
	var s S
	return s, nil
}

func (j *Journal[S, T]) Snapshot(log_filename string, snapshot_filename string) {
	// Switch file outputs, either by opening a new file, or calling dup2
	// Asynchronously, we can read the old snapshot if it exists, replay the log on it,
	// and persist the new snapshot. Maybe we do that in another process
}

func (j *Journal[S, T]) Close() error {
	// TODO: should we flush here or assume the client
	// has obeyed the contract?
	return j.file.Close()
}

func (j *Journal[S, T]) Put(entry T) error {
	line := j.handler.WriteEntry(entry) + "\n"
	_, err := j.file.WriteString(line)
	// TODO: what do we do if returned number of bytes is less tha len?
	// TODO: should we buffer? then we can write multiple entries before flushing/syncing
	if err != nil {
		return err
	}

	err = j.file.Sync()
	if err != nil {
		return err
	}

	return nil
}
