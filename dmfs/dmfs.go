package dmfs

import (
	"github.com/hanwen/go-fuse/v2/fs"
)

type dmanagerFilesystem struct {
	fs.Inode
}
