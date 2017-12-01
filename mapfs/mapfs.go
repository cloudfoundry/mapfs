package mapfs

import (
	"log"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"code.cloudfoundry.org/goshims/syscallshim"
)

const (
	CURRENT_ID = -1
)

//go:generate counterfeiter -o ../mapfs_fakes/fake_file_system.go  ../../../hanwen/go-fuse/fuse/pathfs FileSystem

type mapFileSystem struct {
	pathfs.FileSystem
	uid, gid int64
	syscall syscallshim.Syscall
}

func NewMapFileSystem(uid, gid int64, fs pathfs.FileSystem, sys syscallshim.Syscall) pathfs.FileSystem {
	return &mapFileSystem{
		FileSystem: fs,
		uid: uid,
		gid: gid,
		syscall: sys,
	}
}

func (fs *mapFileSystem) OnMount(nodeFs *pathfs.PathNodeFs) {
	if err := fs.syscall.Setregid(CURRENT_ID, int(fs.gid)); err != nil {
		log.Println("Setregid failed!")
		log.Fatal(err)
	}
	if err := fs.syscall.Setreuid(CURRENT_ID, int(fs.uid)); err != nil {
		log.Println("Setreuid failed!")
		log.Fatal(err)
	}
	fs.FileSystem.OnMount(nodeFs)
}

func (fs *mapFileSystem) GetAttr(name string, context *fuse.Context) (a *fuse.Attr, code fuse.Status) {
	a, code = fs.FileSystem.GetAttr(name, context)

	if a != nil {
		if int64(a.Uid) == fs.uid {
			a.Uid = context.Uid
		}
		if int64(a.Gid) == fs.gid {
			a.Gid = context.Gid
		}
	}

	return a, code
}
