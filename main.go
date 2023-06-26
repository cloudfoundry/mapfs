// Mounts another directory while mapping uid and gid to a different user.  Extends loopbackfs.

package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	"code.cloudfoundry.org/goshims/syscallshim"
	"code.cloudfoundry.org/mapfs/mapfs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/hanwen/go-fuse/v2/fuse/nodefs"
	"github.com/hanwen/go-fuse/v2/fuse/pathfs"
)

func main() {
	log.SetFlags(log.Lmicroseconds)

	debug := flag.Bool("debug", false, "")
	uid := flag.Int64("uid", -1, "")
	gid := flag.Int64("gid", -1, "")
	fsName := flag.String("fsname", "mapfs", "")
	autoCache := flag.Bool("auto_cache", false, "")
	configFile := flag.String("config", "", "configuration file path (optional)")

	flag.Parse()
	if flag.NArg() < 2 || *uid <= 0 || *gid <= 0 {
		fmt.Printf("usage: %s -uid UID -gid GID [-fsname FSNAME] [-auto_cache] [-debug] MOUNTPOINT ORIGINAL\n", path.Base(os.Args[0]))
		fmt.Printf("UID and GID must be > 0")
		os.Exit(2)
	}
	cfg, err := parseConfigFile(*configFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
	if !*debug {
		debug = &cfg.Debug
	}
	if cfg.CPUProfile != "" {
		cpuprofile(cfg.CPUProfile)
		defer pprof.StopCPUProfile()
	}
	if cfg.MemProfile != "" {
		memprofile(cfg.MemProfile)
	}

	orig := flag.Arg(1)
	loopbackfs := pathfs.NewLoopbackFileSystem(orig)
	finalFs := mapfs.NewMapFileSystem(*uid, *gid, loopbackfs, orig, &syscallshim.SyscallShim{})

	opts := &nodefs.Options{
		NegativeTimeout: time.Second,
		AttrTimeout:     time.Second,
		EntryTimeout:    time.Second,
	}

	fuseOpts := []string{}
	if *autoCache {
		fmt.Println("warning -- auto_cache flag ignored as it is unsupported in fusermount")
	}

	pathFs := pathfs.NewPathNodeFs(finalFs, &pathfs.PathNodeFsOptions{})
	conn := nodefs.NewFileSystemConnector(pathFs.Root(), opts)
	mountPoint := flag.Arg(0)
	origAbs, _ := filepath.Abs(orig)
	mOpts := &fuse.MountOptions{
		AllowOther:     true,
		Name:           *fsName,
		FsName:         origAbs,
		Debug:          *debug,
		SingleThreaded: cfg.SingleThreaded,
	}
	if len(fuseOpts) > 0 {
		mOpts.Options = fuseOpts
	}
	state, err := fuse.NewServer(conn.RawFS(), mountPoint, mOpts)
	if err != nil {
		fmt.Printf("Mount fail: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Mounted!")
	state.Serve()
}

// config is the struct used to parse a configuration file
type config struct {
	Debug          bool   `yaml:"debug"`
	SingleThreaded bool   `yaml:"single_threaded"`
	CPUProfile     string `yaml:"cpu_profile"`
	MemProfile     string `yaml:"mem_profile"`
}

// parseConfigFile reads and parses the configuration file
// If no path is specified, but a file is found at the default path then it will
// be read and parsed.
func parseConfigFile(path string) (config, error) {
	const defaultPath = "/var/vcap/jobs/mapfs/config/mapfs.yml"
	if path == "" {
		_, err := os.Stat(defaultPath)
		if err != nil {
			return config{}, nil
		}
		path = defaultPath
	}

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return config{}, fmt.Errorf("could not read config file: %w", err)
	}

	var receiver config
	if err := yaml.Unmarshal(data, &receiver); err != nil {
		return config{}, fmt.Errorf("error parsing config file data: %w", err)
	}

	return receiver, nil
}

// memprofile starts a goroutine that will write a memory profile whenever
// the SIGUSR1 signal is sent to the mapfs process
func memprofile(path string) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR1)
	go func() {
		for range sigs {
			fh, err := os.Create(filepath.Clean(path) + time.Now().Format("2006-01-02T15-04-05"))
			if err != nil {
				log.Printf("could not create memory profile file: %s", err)
				continue
			}

			runtime.GC() // get up-to-date statistics
			if err := pprof.WriteHeapProfile(fh); err != nil {
				log.Printf("error writing memory profile: %s", err)
			}
			if err := fh.Close(); err != nil {
				log.Printf("error closing file: %s", err)
			}
		}
	}()
}

// cpuprofile will start a CPU profile
func cpuprofile(path string) {
	fh, err := os.Create(filepath.Clean(path))
	if err != nil {
		log.Fatal(err)
	}
	if err := pprof.StartCPUProfile(fh); err != nil {
		log.Fatal(err)
	}
}
