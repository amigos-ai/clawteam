package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// Predefined frame sets.
var (
	LobsterFrames = []string{"🦞 ", " 🦞", " 🦞", "🦞 "}
	DotsFrames    = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
)

// SpinnerOption configures a Spinner.
type SpinnerOption func(*Spinner)

// WithFrames sets the animation frames.
func WithFrames(frames []string) SpinnerOption {
	return func(s *Spinner) { s.frames = frames }
}

// WithInterval sets the animation speed.
func WithInterval(d time.Duration) SpinnerOption {
	return func(s *Spinner) { s.interval = d }
}

// WithWriter sets the output writer.
func WithWriter(w io.Writer) SpinnerOption {
	return func(s *Spinner) { s.writer = w }
}

// Spinner displays an animated spinner in the terminal.
type Spinner struct {
	frames   []string
	interval time.Duration
	writer   io.Writer
	isTTY    bool

	mu      sync.Mutex
	message string
	stop    chan struct{}
	done    chan struct{}
}

// NewSpinner creates a spinner with the given options.
func NewSpinner(opts ...SpinnerOption) *Spinner {
	s := &Spinner{
		frames:   LobsterFrames,
		interval: 120 * time.Millisecond,
		writer:   os.Stderr,
		isTTY:    isTTY(),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Start begins the spinner animation with the given message.
// No-op if output is not a TTY.
func (s *Spinner) Start(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isTTY || s.stop != nil {
		return
	}

	s.message = message
	s.stop = make(chan struct{})
	s.done = make(chan struct{})

	go s.run()
}

// Update changes the displayed message while spinning.
func (s *Spinner) Update(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.message = message
}

// Stop clears the spinner line.
func (s *Spinner) Stop() {
	s.mu.Lock()
	if s.stop == nil {
		s.mu.Unlock()
		return
	}
	close(s.stop)
	s.mu.Unlock()

	<-s.done
	s.clearLine()

	s.mu.Lock()
	s.stop = nil
	s.done = nil
	s.mu.Unlock()
}

// StopWith clears the spinner and prints a final message.
func (s *Spinner) StopWith(message string) {
	s.Stop()
	fmt.Fprintln(s.writer, message)
}

func (s *Spinner) run() {
	defer close(s.done)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	i := 0
	for {
		s.mu.Lock()
		msg := s.message
		s.mu.Unlock()

		frame := s.frames[i%len(s.frames)]
		line := fmt.Sprintf("\r%s %s", frame, msg)
		// Pad with spaces to overwrite any previous longer line
		fmt.Fprintf(s.writer, "%-80s", line)

		i++

		select {
		case <-s.stop:
			return
		case <-ticker.C:
		}
	}
}

func (s *Spinner) clearLine() {
	fmt.Fprintf(s.writer, "\r%s\r", strings.Repeat(" ", 80))
}

func isTTY() bool {
	fi, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
