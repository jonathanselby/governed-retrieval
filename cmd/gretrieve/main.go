package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jonathanselby/governed-retrieval/internal/retrieval"
	"os"
	"path/filepath"
)

const version = "0.1.0"

func main() {
	var corpus, caller, q string
	var limit int
	var v bool
	flag.StringVar(&corpus, "corpus", "", "corpus manifest JSON")
	flag.StringVar(&caller, "caller", "", "caller JSON")
	flag.StringVar(&q, "query", "", "query")
	flag.IntVar(&limit, "limit", 3, "maximum citations")
	flag.BoolVar(&v, "version", false, "print version")
	flag.Parse()
	if v {
		fmt.Println("gretrieve " + version)
		return
	}
	if corpus == "" || caller == "" || q == "" {
		flag.Usage()
		os.Exit(2)
	}
	var m retrieval.Manifest
	var c retrieval.Caller
	if err := read(corpus, &m); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}
	if err := read(caller, &c); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}
	r, err := retrieval.Retrieve(filepath.Dir(corpus), m, c, q, limit)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}
	json.NewEncoder(os.Stdout).Encode(r)
	if r.Refused {
		os.Exit(1)
	}
}
func read(p string, v any) error {
	b, e := os.ReadFile(p)
	if e != nil {
		return e
	}
	return json.Unmarshal(b, v)
}
