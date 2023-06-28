// Package main is a CF app which can be pushed to test how MapFS behaves under heavy load.
// The app creates a large number of files and then continually updates them.
// To push the app, run: cf push loader -u process
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

func main() {
	var receiver struct {
		NFS []struct {
			VolumeMounts []struct {
				ContainerDir string `json:"container_dir"`
			} `json:"volume_mounts"`
		} `json:"nfs"`
	}

	vs := os.Getenv("VCAP_SERVICES")
	if vs == "" {
		panic("no VCAP_SERVICES")
	}

	if err := json.Unmarshal([]byte(vs), &receiver); err != nil {
		panic(err)
	}

	if len(receiver.NFS) == 0 || len(receiver.NFS[0].VolumeMounts) == 0 || len(receiver.NFS[0].VolumeMounts[0].ContainerDir) == 0 {
		panic("path parse issue: " + vs)
	}

	root := filepath.Clean(receiver.NFS[0].VolumeMounts[0].ContainerDir)
	fmt.Println("ROOT PATH", root)

	work := make(chan int)
	const workers = 100
	for w := 0; w < workers; w++ {
		go func() {
			for i := range work {
				dir, name := leaf(i)
				if err := os.MkdirAll(path.Join(root, dir), 0750); err != nil {
					panic(err)
				}

				p := path.Join(root, dir, name)
				data, err := os.ReadFile(filepath.Clean(p))
				if err != nil {
					// File doesn't exist yet
					data = []byte(fmt.Sprintf("%d", i))
				}

				val, err := strconv.ParseInt(string(data), 10, 64)
				if err != nil {
					panic(err)
				}

				if err := os.WriteFile(p, []byte(fmt.Sprintf("%d", val+1)), 0600); err != nil {
					panic(err)
				}
			}
		}()
	}

	for {
		for i := 0; i < 1_000_000; i++ {
			work <- i
		}
	}
}

func leaf(count int) (string, string) {
	f := fmt.Sprintf("%06d", count)
	return path.Join(f[0:2], f[2:4]), f[4:6]
}
