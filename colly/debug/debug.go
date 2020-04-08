package debug

import (
	"io"
	"sync/atomic"
	"time"

	"github.com/gocolly/colly/v2/debug"

	"github.com/sirupsen/logrus"
)

type LogrusDebugger struct {
	// Output is the log destination, anything can be used which implements them
	// io.Writer interface. Leave it blank to use STDERR
	Output io.Writer
	// Prefix appears at the beginning of each generated log line
	Prefix string
	// Flag defines the logging properties.
	Flag    int
	Logrus  *logrus.Entry
	counter int32
	start   time.Time
}

// Init initializes the LogrusDebugger
func (l *LogrusDebugger) Init() error {
	l.counter = 0
	l.start = time.Now()
	if l.Logrus == nil {
		l.Logrus = logrus.WithField("package", "colly/debugger")
	}
	return nil
}

// Event receives Collector events and prints them
func (l *LogrusDebugger) Event(e *debug.Event) {
	i := atomic.AddInt32(&l.counter, 1)
	l.Logrus.Debugf("[%06d] %d [%6d - %s] %q (%s)\n", i, e.CollectorID, e.RequestID, e.Type, e.Values, time.Since(l.start))
}
