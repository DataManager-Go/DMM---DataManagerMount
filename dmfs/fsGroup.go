package dmfs

import (
	"context"
	"syscall"

	"github.com/DataManager-Go/libdatamanager"
	"github.com/hanwen/go-fuse/v2/fs"
)

const (
	// NoGroupFolder foldername for
	// files without groups
	NoGroupFolder = "no_group"
)

var (
	_ = (fs.NodeOnAdder)((*groupInode)(nil))
)

type groupInode struct {
	fs.Inode

	namespace            string
	group                string
	isNoGroupPlaceholder bool
}

// List files
func (groupInode *groupInode) OnAdd(ctx context.Context) {
	groupAdd := []string{groupInode.group}

	// Check if group is no_group
	if len(groupAdd) == 1 && groupAdd[0] == NoGroupFolder {
		groupInode.isNoGroupPlaceholder = true
		groupAdd = []string{}
	}

	files, err := data.libdm.ListFiles("", 0, false, libdatamanager.FileAttributes{
		Groups:    groupAdd,
		Namespace: groupInode.namespace,
	}, 0)

	if err != nil {
		printResponseError(err, "getting files for "+groupInode.group)
		return
	}

	for _, file := range files.Files {
		fileNode := groupInode.NewInode(ctx, &fileInode{
			file: &file,
		}, fs.StableAttr{
			Mode: syscall.S_IFREG,
		})

		groupInode.AddChild(file.Name, fileNode, true)
	}
}
