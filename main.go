// Mounts another directory while mapping uid and gid to a different user.  Extends loopbackfs.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"github.com/julian-hj/mapfs/mapfs"
	"code.cloudfoundry.org/goshims/syscallshim"
)

func main() {
	log.SetFlags(log.Lmicroseconds)

	debug := flag.Bool("debug", false, "print debugging messages.")
	uid := flag.Int64("uid", -1, "POSIX UID to map to")
	gid := flag.Int64("gid", -1, "POSIX GID to map to")

	flag.Parse()
	if flag.NArg() < 2 {
		fmt.Printf("usage: %s MOUNTPOINT ORIGINAL\n", path.Base(os.Args[0]))
		fmt.Printf("\noptions:\n")
		flag.PrintDefaults()
		os.Exit(2)
	}

	var finalFs pathfs.FileSystem
	orig := flag.Arg(1)
	loopbackfs := pathfs.NewLoopbackFileSystem(orig)
	mapfs := mapfs.NewMapFileSystem(*uid, *gid, loopbackfs, &syscallshim.SyscallShim{})
	finalFs = mapfs

	opts := &nodefs.Options{
		NegativeTimeout: time.Second,
		AttrTimeout:     time.Second,
		EntryTimeout:    time.Second,
	}

	pathFs := pathfs.NewPathNodeFs(finalFs, &pathfs.PathNodeFsOptions{})
	conn := nodefs.NewFileSystemConnector(pathFs.Root(), opts)
	mountPoint := flag.Arg(0)
	origAbs, _ := filepath.Abs(orig)
	mOpts := &fuse.MountOptions{
		AllowOther: true,
		Name:       "mapfs",
		FsName:     origAbs,
		Debug:      *debug,
	}
	state, err := fuse.NewServer(conn.RawFS(), mountPoint, mOpts)
	if err != nil {
		fmt.Printf("Mount fail: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Mounted!")
	state.Serve()
}
