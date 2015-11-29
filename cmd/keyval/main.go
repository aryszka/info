package main

import (
	"encoding/json"
	"fmt"
	"github.com/aryszka/keyval"
	"io"
	"log"
	"os"
	"strings"
)

func printKeyVal(kv *keyval.Entry) {
	key := strings.Join(kv.Key, ".")
	fmt.Printf("# %s\n%s: %s\n\n", kv.Comment, key, kv.Val)
}

func read() error {
	r := keyval.NewReader(os.Stdin)
	for {
		kv, err := r.ReadEntry()
		if err != nil && err != io.EOF {
			return err
		}

		if kv != nil {
			printKeyVal(kv)
		}

		if err != nil {
			return nil
		}
	}
}

func write(w *keyval.Writer, keys []string, v interface{}) error {
	switch vt := v.(type) {
	case map[string]interface{}:
		for k, vi := range vt {
			if err := write(w, append(keys, k), vi); err != nil {
				return err
			}
		}

		return nil
	case []interface{}:
		for _, vi := range vt {
			if err := write(w, keys, vi); err != nil {
				return err
			}
		}

		return nil
	default:
		return w.WriteEntry(&keyval.Entry{Key: keys, Val: fmt.Sprint(vt)})
	}
}

func writeJson() error {
	d := json.NewDecoder(os.Stdin)
	m := make(map[string]interface{})
	if err := d.Decode(&m); err != nil {
		return err
	}

	w := keyval.NewWriter(os.Stdout)
	return write(w, nil, m)
}

func main() {
	if err := read(); err != nil {
		log.Fatal(err)
	}
}
