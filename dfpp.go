package main

import (
	"github.com/op/go-logging"
	"os"
	"io"
	"io/ioutil"
	"net/http"
	"fmt"
	"bufio"
	"runtime"
	"strings"
	"github.com/droundy/goopt"
)

var log = logging.MustGetLogger("dfpp")
var format = "%{color}%{time:2006-01-02T15:04:05.000Z07:00} %{level:-5s} [%{shortfile}]%{color:reset} %{message}"

func main() {
	logBackend := logging.NewLogBackend(os.Stderr, "", 0)
	logging.SetBackend(
		logging.NewBackendFormatter(
			logBackend,
			logging.MustStringFormatter(format),
		),
	)
	logging.SetLevel(logging.DEBUG, "")
	
	for line := range InstructionScanner(os.Stdin) {
		parts := strings.Fields(line)
		if len(parts) > 0 {
			instruction := parts[0]
			if instruction == "INCLUDE" {
				ProcessInclude(line, parts)
				continue;
			}
		}
		fmt.Println(line)
	}
}

func InstructionScanner(input io.Reader) chan string {
	ch := make(chan string)
	go func() {
		scanner := bufio.NewScanner(input)
		for scanner.Scan() {
			line := scanner.Text();
			for len(line)>0 && line[len(line)-1] == '\\' {
				scanner.Scan();
				line += "\n" + scanner.Text()
			}
			ch <- line
		}
		close(ch)
	}()
	return ch
}

func ProcessInclude(line string, fields []string) {
	merge := false
	exclude := make(map[string]bool)
	include := make(map[string]bool)
	
	uris := make([]string,0, len(fields)-1)
	for _, field := range fields {
		if _, err := os.Stat(field); err == nil {
			uris = append(uris, field)
			continue;
		}
		clude := include
		if field[0] == '-' {
			clude = exclude
			field = field[1:]
		}

		switch field {
        case "\\": continue
		case "INCLUDE": continue
		case "MERGE": merge = true
		case "ADD": fallthrough
		case "CMD": fallthrough
		case "COPY": fallthrough
		case "ENTRYPOINT": fallthrough
		case "EVN": fallthrough
		case "EXPOSE": fallthrough
		case "FROM": fallthrough
		case "LABEL": fallthrough
		case "MAINTAINER": fallthrough
		case "ONBUILD": fallthrough
		case "RUN": fallthrough
		case "USER": fallthrough
		case "VOLUME": fallthrough
		case "WORKDIR": clude[field] = true
		default: uris = append(uris, field)
		}
	}

	docs := make([]string, 0, len(uris))
	for _, uri := range uris {
		if _, err := os.Stat(uri); err == nil {
			content, err := ioutil.ReadFile(uri)
			if err != nil {
				log.Error("Failed to read %s: %s", uri, err)
				os.Exit(1)
			}
			docs = append(docs, string(content))
		} else {
			req, _ := http.NewRequest("GET", uri, nil)
			ua := &http.Client{}
			if resp, err := ua.Do(req); err != nil {
				log.Error("Failed to %s %s: %s", req.Method, req.URL.String(), err)
				os.Exit(1)
			} else {
				if resp.StatusCode < 200 || resp.StatusCode >= 300 && resp.StatusCode != 401 {
					log.Error("response status: %s", resp.Status)
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
	Merge(merge, docs, include, exclude)
}
	
func Merge(merge bool, docs []string, include, exclude map[string]bool) {
	result := make([]*string, 0)
	ops := make(map[string]*string)
	for _, doc := range docs {
		for line := range InstructionScanner( strings.NewReader(doc) ) {
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
					if op == "ENV" || op == "LABEL" {
						*sref += " \\\n" + strings.Repeat(" ", len(op)+1) + details
					} else if op == "RUN" {
						*sref += " && \\\n    " + details

						// squash redundent apt-get updates
						squash := "apt-get update"
						if ix := strings.Index(*sref, squash); ix >= 0 {
							rest := strings.Replace((*sref)[ix+len(squash):], squash, "echo skipping redundent apt-get-update", -1);
							*sref = (*sref)[0:ix+len(squash)] + rest
						}
						
					} else {
						dup := string(line)
						result = append(result, &dup)
					}
				} else {
					dup := string(line)
					result = append(result, &dup)
					ops[op] = result[len(result)-1]
				}
			}
		}
	}
	for _, line := range result {
		fmt.Println(*line)
	}
}