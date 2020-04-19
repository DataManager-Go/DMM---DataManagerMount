package dmfs

import (
	"context"
	"fmt"
	"syscall"

	"github.com/DataManager-Go/libdatamanager"
	"github.com/hanwen/go-fuse/v2/fs"
)

type groupInode struct {
	fs.Inode

	namespace            string
	group                string
	isNoGroupPlaceholder bool
}

var _ = (fs.NodeOnAdder)((*groupInode)(nil))

func (groupInode *groupInode) OnAdd(ctx context.Context) {
	groupAdd := []string{groupInode.group}

	if len(groupAdd) == 1 && groupAdd[0] == "no_group" {
		groupInode.isNoGroupPlaceholder = true
		groupAdd = []string{}
	}

	files, err := data.libdm.ListFiles("", 0, false, libdatamanager.FileAttributes{
		Groups:    groupAdd,
		Namespace: groupInode.namespace,
	}, 0)

	if err != nil {
		fmt.Println(err)
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
