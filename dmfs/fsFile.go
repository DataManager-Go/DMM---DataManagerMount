package dmfs

import (
	"context"
	"fmt"
	"syscall"

	libdm "github.com/DataManager-Go/libdatamanager"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type fileInode struct {
	fs.Inode

	file *libdm.FileResponseItem
}

var (
	_ = (fs.NodeGetattrer)((*fileInode)(nil))
	_ = (fs.NodeUnlinker)((*fileInode)(nil))
	_ = (fs.NodeReader)((*fileInode)(nil))
)

// Set file attributes
func (fnode *fileInode) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	fmt.Println("getatr")
	out.Size = uint64(fnode.file.Size)
	out.Ctime = uint64(fnode.file.CreationDate.Unix())
	out.Mtime = out.Ctime
	out.Mode = 0770

	// Set owner/group
	data.setUserAttr(out)

	return 0
}

func (fnode *fileInode) getattr(out *fuse.AttrOut) {
	out.Size = uint64(fnode.file.Size)
	out.SetTimes(nil, &fnode.file.CreationDate, &fnode.file.CreationDate)
}

func (fnode *fileInode) Unlink(ctx context.Context, name string) syscall.Errno {
	fmt.Println("rm", name)
	return 0
}

func (fnode *fileInode) Read(ctx context.Context, f fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	return nil, syscall.ENOSPC
}
