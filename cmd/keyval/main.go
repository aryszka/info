package main

import (
	"fmt"
	"github.com/aryszka/keyval"
	"os"
	"strings"
)

func printKeyVal(kv *keyval.Entry) {
	key := strings.Join(kv.Key, ".")
	fmt.Printf("# %s\n%s = %s\n\n", kv.Comment, key, kv.Val)
}

func main() {
	r := keyval.NewReader(os.Stdin)
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
