package csliner

import (
	"github.com/peterh/liner"
)

type CSLiner struct {
	State   *liner.State
	History *LineHistory

	tmode liner.ModeApplier
	lmode liner.ModeApplier

	paused bool
}

func NewLiner() *CSLiner {
	pl := &CSLiner{}
	var err error
	if pl.tmode, err = liner.TerminalMode(); err != nil {
	}

	line := liner.NewLiner()
	if pl.lmode, err = liner.TerminalMode(); err != nil {
	}

	line.SetMultiLineMode(true)
	line.SetCtrlCAborts(true)

	pl.State = line

	return pl
}

func (pl *CSLiner) Pause() error {
	if pl.paused {
		panic("CSLiner already paused")
	}

	pl.paused = true
	pl.DoWriteHistory()

	return pl.tmode.ApplyMode()
}

func (pl *CSLiner) Resume() error {
	if !pl.paused {
		panic("CSLiner is not paused")
	}

	pl.paused = false

	return pl.lmode.ApplyMode()
}

func (pl *CSLiner) Close() (err error) {
	err = pl.State.Close()
	if err != nil {
		return err
	}

	if pl.History != nil && pl.History.historyFile != nil {
		return pl.History.historyFile.Close()
	}

	return nil
}
