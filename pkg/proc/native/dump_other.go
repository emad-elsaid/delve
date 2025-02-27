//go:build (freebsd && amd64) || darwin
// +build freebsd,amd64 darwin

package native

import (
	"github.com/emad-elsaid/delve/pkg/elfwriter"
	"github.com/emad-elsaid/delve/pkg/proc"
)

func (p *nativeProcess) MemoryMap() ([]proc.MemoryMapEntry, error) {
	return nil, proc.ErrMemoryMapNotSupported
}

func (p *nativeProcess) DumpProcessNotes(notes []elfwriter.Note, threadDone func()) (threadsDone bool, notesout []elfwriter.Note, err error) {
	return false, notes, nil
}
