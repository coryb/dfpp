package dfpp

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
)

type Logger interface {
	Errorf(format string, args ...interface{})
}

type defaultLogger struct{}

func (l *defaultLogger) Errorf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

type HttpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Dfpp struct {
	Processors map[string]func(string, []string) bool
	Output     io.Writer
	client     HttpClient
	logger     Logger
}

type Option func(pp *Dfpp)

func WithClient(c HttpClient) Option {
	return func(pp *Dfpp) {
		pp.client = c
	}
}

func WithLogger(l Logger) Option {
	return func(pp *Dfpp) {
		pp.logger = l
	}
}

func NewDfpp(opts ...Option) *Dfpp {
	pp := &Dfpp{
		Output: os.Stdout,
		client: &http.Client{},
		logger: &defaultLogger{},
	}
	pp.Processors = map[string]func(string, []string) bool{
		"INCLUDE": pp.ProcessInclude,
	}
	for _, optFunc := range opts {
		optFunc(pp)
	}
	return pp
}

func (pp *Dfpp) ProcessDockerfile(input io.Reader) {
	for line := range InstructionScanner(input) {
		parts := strings.Fields(line)
		if len(parts) > 0 {
			instruction := parts[0]
			if fn, ok := pp.Processors[instruction]; ok {
				if fn(line, parts) {
					continue
				}
			}
		}
		fmt.Fprintf(pp.Output, "%s\n", line)
	}
}

func InstructionScanner(input io.Reader) chan string {
	ch := make(chan string)
	go func() {
		scanner := bufio.NewScanner(input)
		for scanner.Scan() {
			line := scanner.Text()
			for len(line) > 0 && line[len(line)-1] == '\\' {
				scanner.Scan()
				line += "\n" + scanner.Text()
			}
			ch <- line
		}
		close(ch)
	}()
	return ch
}

func (pp *Dfpp) ProcessInclude(line string, fields []string) bool {
	merge := false
	exclude := make(map[string]bool)
	include := make(map[string]bool)

	uris := make([]string, 0, len(fields)-1)
	for _, field := range fields {
		if _, err := os.Stat(field); err == nil {
			uris = append(uris, field)
			continue
		}
		clude := include
		if field[0] == '-' {
			clude = exclude
			field = field[1:]
		}

		switch field {
		case "\\":
			continue
		case "INCLUDE":
			continue
		case "MERGE":
			merge = true
		case "ADD":
			fallthrough
		case "ARG":
			fallthrough
		case "CMD":
			fallthrough
		case "COPY":
			fallthrough
		case "ENTRYPOINT":
			fallthrough
		case "ENV":
			fallthrough
		case "EXPOSE":
			fallthrough
		case "FROM":
			fallthrough
		case "LABEL":
			fallthrough
		case "MAINTAINER":
			fallthrough
		case "ONBUILD":
			fallthrough
		case "RUN":
			fallthrough
		case "USER":
			fallthrough
		case "VOLUME":
			fallthrough
		case "WORKDIR":
			clude[field] = true
		default:
			uris = append(uris, field)
		}
	}

	docs := make([]string, 0, len(uris))
	for _, uri := range uris {
		if _, err := os.Stat(uri); err == nil {
			content, err := ioutil.ReadFile(uri)
			if err != nil {
				pp.logger.Errorf("Failed to read %s: %s", uri, err)
				os.Exit(1)
			}
			docs = append(docs, string(content))
		} else {
			req, _ := http.NewRequest("GET", uri, nil)
			if resp, err := pp.client.Do(req); err != nil {
				pp.logger.Errorf("Failed to %s %s: %s", req.Method, req.URL.String(), err)
				os.Exit(1)
			} else {
				if resp.StatusCode < 200 || resp.StatusCode >= 300 && resp.StatusCode != 401 {
					pp.logger.Errorf("response status: %s", resp.Status)
				}
				runtime.SetFinalizer(resp, func(r *http.Response) {
					r.Body.Close()
				})
				if buf, err := ioutil.ReadAll(resp.Body); err == nil {
					docs = append(docs, string(buf))
				}
			}
		}
	}
	pp.Merge(merge, docs, include, exclude)
	return true
}

type stringListReader struct {
	lines     []*string
	lineIndex int
	offset    int
}

func (s *stringListReader) Read(p []byte) (n int, err error) {
	want := cap(p)
	readBytes := 0
	for want > 0 {
		if s.lineIndex >= len(s.lines) {
			return readBytes, io.EOF
		}
		line := s.lines[s.lineIndex]
		remainder := len(*line) - s.offset
		if remainder > want {
			copy(p[readBytes:], (*line)[s.offset:want])
			s.offset += want
			readBytes += want - s.offset
			want = 0
		} else {
			copy(p[readBytes:], (*line)[s.offset:])
			readBytes += remainder
			s.offset = 0
			s.lineIndex++
			want -= remainder
			// memory buffer was split on new lines, so make sure to add
			// the newlines back on the output merged document.
			copy(p[readBytes:], "\n")
			readBytes++
		}
	}
	return readBytes, nil
}

func (pp *Dfpp) Merge(merge bool, docs []string, include, exclude map[string]bool) {
	result := make([]*string, 0)
	ops := make(map[string]*string)
	for _, doc := range docs {
		for line := range InstructionScanner(strings.NewReader(doc)) {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				op := fields[0]
				if _, ok := exclude[op]; ok {
					continue
				}
				if _, ok := include[op]; len(include) > 0 && !ok {
					continue
				}
				details := strings.TrimPrefix(line, fields[0]+" ")

				if sref, ok := ops[op]; merge && ok {
					if op == "ENV" {
						if !strings.Contains(details, "=") {
							fields := strings.Fields(details)
							details = strings.Join(fields, "=")
						}
						*sref += " \\\n" + strings.Repeat(" ", len(op)+1) + details
					} else if op == "LABEL" {
						*sref += " \\\n" + strings.Repeat(" ", len(op)+1) + details
					} else if op == "RUN" {
						*sref += " && \\\n    " + details

						// squash redundent apt-get updates
						squash := "apt-get update"
						if ix := strings.Index(*sref, squash); ix >= 0 {
							rest := strings.Replace((*sref)[ix+len(squash):], squash, "echo skipping redundent apt-get-update", -1)
							*sref = (*sref)[0:ix+len(squash)] + rest
						}

					} else {
						dup := string(line)
						result = append(result, &dup)
					}
				} else {
					if op == "ENV" && !strings.Contains(line, "=") {
						line = strings.TrimPrefix(line, "ENV")
						line = strings.TrimSpace(line)
						fields := strings.Fields(line)
						line = fmt.Sprintf("ENV %s", strings.Join(fields, "="))
					}
					dup := string(line)
					result = append(result, &dup)
					ops[op] = result[len(result)-1]
				}
			}
		}
	}

	pp.ProcessDockerfile(&stringListReader{lines: result})
}
