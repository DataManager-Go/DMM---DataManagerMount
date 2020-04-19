package dmfs

import (
	"context"
	"syscall"

	libdm "github.com/DataManager-Go/libdatamanager"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type fileInode struct {
	fs.Inode

	file *libdm.FileResponseItem
}

var _ = (fs.NodeGetattrer)((*fileInode)(nil))

// Set attributes for files
func (fnode *fileInode) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Size = uint64(fnode.file.Size)
	out.Ctime = uint64(fnode.file.CreationDate.Unix())
	out.Mtime = out.Ctime
	out.Mode = 0640
	out.Gid = data.gid
	out.Uid = data.uid

	return 0
}
