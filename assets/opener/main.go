// Package main is a CF app which can be pushed to test how MapFS behaves when lots of file handles are open.
// The app creates as many file handles as it can, and then sleeps.
// To push the app, run: cf push opener -u process
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
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

	if len(receiver.NFS[0].VolumeMounts[0].ContainerDir) == 0 {
		panic("path parse issue: " + vs)
	}

	path := receiver.NFS[0].VolumeMounts[0].ContainerDir
	fmt.Println("PATH", path)

	count := 0
	for {
		_, err := os.CreateTemp(path, "")
		if err != nil {
			fmt.Println("COUNT", count)
			break
		}
	}

	time.Sleep(time.Hour)
}
