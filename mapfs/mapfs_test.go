package mapfs_test

import (
	"code.cloudfoundry.org/lager/lagertest"

	"code.cloudfoundry.org/lager"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/julian-hj/mapfs/mapfs_fakes"
	"github.com/hanwen/go-fuse/fuse/pathfs"
	"github.com/julian-hj/mapfs/mapfs"
	"github.com/hanwen/go-fuse/fuse"
	"code.cloudfoundry.org/goshims/syscallshim/syscall_fake"
)

var _ = Describe("mapfs", func() {
	var (
		logger lager.Logger
		mapFS pathfs.FileSystem
		uid, gid int64

		fakeFS *mapfs_fakes.FakeFileSystem
		fakeSyscall *syscall_fake.FakeSyscall
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test-fs")
		fakeFS = &mapfs_fakes.FakeFileSystem{}
		fakeSyscall = &syscall_fake.FakeSyscall{}
		uid = 5
		gid = 10
	})

	JustBeforeEach(func() {
		mapFS = mapfs.NewMapFileSystem(uid, gid, fakeFS, fakeSyscall)
	})

	Context("when there is a mapfs", func() {
		BeforeEach(func() {
		})

		Context(".Chmod", func() {
			It("passes the function through to the underlying filesystem unchanged", func() {
				context := &fuse.Context{}
				mapFS.Chmod("foo", uint32(0777), context)

				Expect(fakeFS.ChmodCallCount()).To(Equal(1))
				name, mode, passedContext := fakeFS.ChmodArgsForCall(0)
				Expect(name).To(Equal("foo"))
				Expect(mode).To(Equal(uint32(0777)))
				Expect(passedContext).To(Equal(context))
			})
		})

		Context(".GetAttr", func() {
			It("maps the uid/gid back to the fuse context uid when it matches the mapped id", func() {
				context := &fuse.Context{}
				context.Uid = 50
				context.Gid = 100
				attr := &fuse.Attr{}
				attr.Uid = uint32(uid)
				attr.Gid = uint32(gid)
				attr.Mode = uint32(0777)
				fakeFS.GetAttrReturns(attr, fuse.OK)
				ret, code := mapFS.GetAttr("foo", context)

				Expect(fakeFS.GetAttrCallCount()).To(Equal(1))
				Expect(code).To(Equal(fuse.OK))
				Expect(ret.Uid).To(Equal(context.Uid))
				Expect(ret.Gid).To(Equal(context.Gid))
				Expect(ret.Mode).To(Equal(uint32(0777)))
			})
		})

		Context(".OnMount", func() {
			It("passes through to the underlying fs", func() {
				mapFS.OnMount(nil)

				Expect(fakeFS.OnMountCallCount()).To(Equal(1))
			})
		})
	})
})
