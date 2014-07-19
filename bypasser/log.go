package bypasser

import (
	"errors"
	"io"
	"log"
)

const (
	invalidTagError = "Invalid tag."
)

type GFWLogger struct {
	logger       *log.Logger
	loggingTable []bool
	loggingTags  []string
}

func NewLogger(out io.Writer, tags []string) *GFWLogger {
	tableLength := len(tags)
	loggingTable := make([]bool, tableLength, cap(tags))
	for i := 0; i < tableLength; i++ {
		loggingTable[i] = true
	}
	return &GFWLogger{log.New(out, "", log.LstdFlags), loggingTable, tags}
}

func (l *GFWLogger) Log(tag int, text string) {
	if l.willLog(tag) {
		(*l).logger.Println((*l).loggingTags[tag] + " " + text)
	}
}

func (l *GFWLogger) DisableTag(tag int) error {
	return l.setTagVisibility(tag, false)
}

func (l *GFWLogger) EnableTag(tag int) error {
	return l.setTagVisibility(tag, false)
}

func (l *GFWLogger) setTagVisibility(tag int, isVisible bool) error {
	if l.isTagValid(tag) {
		l.loggingTable[tag] = isVisible
		return nil
	} else {
		return errors.New(invalidTagError)
	}
}

func (l *GFWLogger) willLog(tag int) bool {
	if l.isTagValid(tag) {
		return l.loggingTable[tag]
	} else {
		return false
	}
}

func (l *GFWLogger) isTagValid(tag int) bool {
	return tag <= len(l.loggingTable)
}
