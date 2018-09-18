// Copyright 2018 Matt Ho. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/gorilla/mux"
	"gopkg.in/yaml.v2"
)

func makeHandler(host, repo, display, vcs string) (http.HandlerFunc, error) {
	switch {
	case display != "":
		// Already filled in.
	case strings.HasPrefix(repo, "https://github.com/"):
		display = fmt.Sprintf("%v %v/tree/master{/dir} %v/blob/master{/dir}/{file}#L{line}", repo, repo, repo)
	case strings.HasPrefix(repo, "https://bitbucket.org"):
		display = fmt.Sprintf("%v %v/src/default{/dir} %v/src/default{/dir}/{file}#{file}-{line}", repo, repo, repo)
	}

	switch {
	case vcs != "":
		// Already filled in.
		if vcs != "bzr" && vcs != "git" && vcs != "hg" && vcs != "svn" {
			return nil, fmt.Errorf("configuration for %v: unknown VCS %s", repo, vcs)
		}
	case strings.HasPrefix(repo, "https://github.com/"):
		vcs = "git"
	case strings.HasPrefix(repo, "https://bitbucket.org/"):
		vcs = "git"
	default:
		return nil, fmt.Errorf("cannot infer VCS from %s", repo)
	}

	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)

		vanityTemplate.Execute(w, struct {
			Import  string
			Path    string
			Repo    string
			Display string
			VCS     string
		}{
			Import:  host,
			Path:    strings.TrimPrefix(req.URL.Path, "/"),
			Repo:    repo,
			Display: display,
			VCS:     vcs,
		})
	}, nil
}

func index(host string, paths []string) http.HandlerFunc {
	var handlers []string
	for _, path := range paths {
		handlers = append(handlers, strings.TrimSuffix(host+path, "/"))
	}
	sort.Strings(handlers)
	fmt.Println(handlers)
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)

		indexTemplate.Execute(w, struct {
			Host     string
			Handlers []string
		}{
			Host:     host,
			Handlers: handlers,
		})
	}
}

func withCacheControl(h http.Handler, maxAge int) http.HandlerFunc {
	cacheControl := fmt.Sprintf("public, max-age=%d", maxAge)

	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Cache-Control", cacheControl)
		h.ServeHTTP(w, req)
	}
}

func parse(data []byte) (http.HandlerFunc, error) {
	var config struct {
		Host   string `yaml:"host,omitempty"`
		MaxAge *int   `yaml:"max_age,omitempty"`
		Paths  map[string]struct {
			Repo    string `yaml:"repo,omitempty"`
			Display string `yaml:"display,omitempty"`
			VCS     string `yaml:"vcs,omitempty"`
		} `yaml:"paths,omitempty"`
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	var paths []string
	for path := range config.Paths {
		paths = append(paths, path)
	}
	sort.Slice(paths, func(i, j int) bool {
		li, lj := len(paths[i]), len(paths[j])
		if li == lj {
			return paths[i] < paths[j]
		}
		return li > lj
	})

	// bind paths, longest to shortest
	router := mux.NewRouter()
	for _, path := range paths {
		entry := config.Paths[path]
		h, err := makeHandler(config.Host, entry.Repo, entry.Display, entry.VCS)
		if err != nil {
			return nil, err
		}

		path = strings.TrimSuffix(path, "/")
		if len(path) > 0 {
			router.Handle(path, h)
		}
		router.PathPrefix(path + "/").HandlerFunc(h)
	}
	if _, ok := config.Paths["/"]; !ok {
		router.HandleFunc("/", index(config.Host, paths))
	}

	// define max cache age
	maxAge := 86400 // 1 day
	if config.MaxAge != nil && *config.MaxAge >= 0 {
		maxAge = *config.MaxAge
	}

	return withCacheControl(router, maxAge), nil
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	data, err := ioutil.ReadFile("vanity.yml")
	check(err)

	h, err := parse(data)
	check(err)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	http.ListenAndServe(":"+port, h)
}

var indexTemplate = template.Must(template.New("index").Parse(`<!DOCTYPE html>
<html>
<h1>{{.Host}}</h1>
<ul>
{{range .Handlers}}<li><a href="https://godoc.org/{{.}}">{{.}}</a></li>
{{end}}</ul>
</html>
`))

var vanityTemplate = template.Must(template.New("vanity").Parse(`<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
<meta name="go-import" content="{{.Import}} {{.VCS}} {{.Repo}}">
<meta name="go-source" content="{{.Import}} {{.Display}}">
<meta http-equiv="refresh" content="0; url=https://godoc.org/{{.Import}}/{{.Path}}">
</head>
<body>
Nothing to see here; <a href="https://godoc.org/{{.Import}}/{{.Path}}">see the package on godoc</a>.
</body>
</html>
`))
