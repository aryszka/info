package main

import (
	"encoding/json"
	"fmt"
	"github.com/aryszka/info"
	"log"
	"os"
	"strings"
)

func printKeyVal(kv *info.Entry) {
	key := strings.Join(kv.Key, ".")
	fmt.Printf("# %s\n%s: %s\n\n", kv.Comment, key, kv.Val)
}

func read() {
	r := info.NewReader(os.Stdin)
	for {
		kv, err := r.ReadEntry()
		if kv != nil {
			printKeyVal(kv)
		}

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
	}
}

func write(w *info.Writer, keys []string, v interface{}) error {
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
		return w.WriteEntry(&info.Entry{Key: keys, Val: fmt.Sprint(v)})
	}
}

func writeJson() {
	d := json.NewDecoder(os.Stdin)
	m := make(map[string]interface{})
	if err := d.Decode(&m); err != nil {
		log.Fatal(err)
	}

	w := info.NewWriter(os.Stdout)
	write(w, nil, m)
}

func main() {
	writeJson()
}
