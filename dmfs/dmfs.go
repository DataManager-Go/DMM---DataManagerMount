package dmfs

import (
	"context"
	"fmt"
	"syscall"

	"github.com/DataManager-Go/libdatamanager"
	dmConfig "github.com/DataManager-Go/libdatamanager/config"
	"github.com/hanwen/go-fuse/v2/fs"
)

type dmanagerFilesystem struct {
	fs.Inode

	mounter *Mounter
	config  *dmConfig.Config
	libdm   *libdatamanager.LibDM
}

// Ensure that we implement NodeOnAdder
var _ = (fs.NodeOnAdder)((*dmanagerFilesystem)(nil))
var _ = (fs.NodeRenamer)((*dmanagerFilesystem)(nil))
var _ = (fs.NodeUnlinker)((*dmanagerFilesystem)(nil))

// OnAdd is called on mounting the file system. Use it to populate
// the file system tree.
func (root *dmanagerFilesystem) OnAdd(ctx context.Context) {
	root.initFiles()
}

// Unlink if virtual file was unlinked
func (root *dmanagerFilesystem) Unlink(ctx context.Context, name string) syscall.Errno {
	fmt.Println("rm", name)
	return 0
}

// Rename if virtual file was renamed
func (root *dmanagerFilesystem) Rename(ctx context.Context, name string, newParent fs.InodeEmbedder, newName string, flags uint32) syscall.Errno {
	fmt.Println("renome")
	return 0
}

func (root *dmanagerFilesystem) debug(arg ...interface{}) {
	if root.mounter.Debug {
		fmt.Println(arg...)
	}
}
