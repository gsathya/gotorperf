package torctl

import (
	"fmt"
	"strings"
)

type EventType int
type NewEventFunc func([]string) (Event, error)

var muxer = map[string]NewEventFunc{
	"CIRC":   NewCircEvent,
	"STREAM": NewStreamEvent,
}

const (
	CIRC EventType = iota
	STREAM
)

type Event interface {
	Type() EventType
}

func Parse(line string) (Event, error) {
	values := strings.Split(line, " ")
	eventType := values[0]
	e, ok := muxer[eventType]
	if !ok {
		return nil, fmt.Errorf("unknown event: %s", eventType)
	}

	return e(values[1:])
}

type CircEvent struct {
	Id     string
	Status string
	Path   []Path
}

type Path struct {
	Fingerprint string
	Nickname    string
}

func (c CircEvent) Type() EventType {
	return CIRC
}

func NewCircEvent(values []string) (Event, error) {
	c := &CircEvent{}

	if len(values) < 3 {
		return nil, fmt.Errorf("malformed circ event: %s", values)
	}

	c.Id = values[0]
	c.Status = values[1]
	var err error
	if c.Path, err = parsePath(values[2]); err != nil {
		return nil, err
	}

	return c, nil
}

func parsePath(p string) ([]Path, error) {
	var (
		paths []Path
		fp    string
		nick  string
	)

	if len(p) == 0 {
		return nil, fmt.Errorf("empty path given: %s", p)
	}

	for _, entry := range strings.Split(p, ",") {
		switch {
		case strings.Contains(entry, "="):
			fp, nick = split(entry, "=")
		case strings.Contains(entry, "~"):
			fp, nick = split(entry, "~")
		case entry[0] == '$':
			fp, nick = entry, ""
		default:
			fp, nick = "", entry
		}

		//XXX: validate fp and nick
		paths = append(paths, Path{fp, nick})
	}
	return paths, nil
}

func split(s, sep string) (string, string) {
	x := strings.SplitN(s, sep, 2)
	return x[0], x[1]
}

type StreamEvent struct {
	Id     string
	Status string
	CircId string
	Target string
}

func (s StreamEvent) Type() EventType {
	return STREAM
}

func NewStreamEvent(values []string) (Event, error) {
	s := &StreamEvent{}

	if len(values) < 4 {
		return nil, fmt.Errorf("malformed stream event: %s", values)
	}

	s.Id = values[0]
	s.Status = values[1]
	s.CircId = values[2]
	s.Target = values[3]
	return s, nil
}
