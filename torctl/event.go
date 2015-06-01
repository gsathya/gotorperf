package torctl

import (
	"fmt"
	"strings"
)

type CircEvent struct {
	Id     string
	Status string
	Path   []Path
}

type Path struct {
	Fingerprint string
	Nickname    string
}

func NewCircEvent(line string) (c *CircEvent, err error) {
	values := strings.SplitN(line, "", 5)
	if len(values) < 4 {
		return nil, fmt.Errorf("Malformed circ event: %s", line)
	}

	c.Id = values[1]
	c.Status = values[2]
	if c.Path, err = parsePath(values[3]); err != nil {
		return nil, err
	}

	return
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
			fp, nick = split("=", entry)
		case strings.Contains(entry, "~"):
			fp, nick = split("~", entry)
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
