package dmfs

import (
	"context"
	"fmt"
	"syscall"
	"time"

	"github.com/DataManager-Go/libdatamanager"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type namespaceNode struct {
	fs.Inode

	namespace string
	groups    []string
}

var _ = (fs.NodeOnAdder)((*namespaceNode)(nil))
var _ = (fs.NodeRmdirer)((*namespaceNode)(nil))
var _ = (fs.NodeMkdirer)((*namespaceNode)(nil))
var _ = (fs.NodeRenamer)((*namespaceNode)(nil))

func (nsNode *namespaceNode) OnAdd(ctx context.Context) {
	// Use a no_group folder for files
	// not associated to a groud
	if len(nsNode.groups) == 0 {
		nsNode.groups = []string{"no_group"}
	}

	// Add groups to namespace
	for _, group := range nsNode.groups {
		gp := nsNode.GetChild(group)
		if gp == nil {
			gp = nsNode.NewInode(ctx, &groupInode{
				group:     group,
				namespace: nsNode.namespace,
			}, fs.StableAttr{
				Mode: syscall.S_IFDIR,
			})

			nsNode.AddChild(group, gp, true)
		}
	}
}

// Delete group if vfile was removed
func (nsNode *namespaceNode) Rmdir(ctx context.Context, name string) syscall.Errno {
	// wait 2 seconds to ensure, user didn't cancel
	select {
	case <-ctx.Done():
		return syscall.ECANCELED
	case <-time.After(2 * time.Second):
	}

	// Do http delete request
	_, err := data.libdm.DeleteAttribute(libdatamanager.GroupAttribute, nsNode.namespace, name)
	if err != nil {
		fmt.Println(err)
		return syscall.ENOENT
	}

	return 0
}

// On group renamed
func (nsNode *namespaceNode) Rename(ctx context.Context, name string, newParent fs.InodeEmbedder, newName string, flags uint32) syscall.Errno {
	if name == "no_group" {
		// TODO add groups to files with no group
		return 0
	}

	_, err := data.libdm.UpdateAttribute(libdatamanager.GroupAttribute, nsNode.namespace, name, newName)
	if err != nil {
		return syscall.ENOENT
	}

	return 0
}

// On group created
func (nsNode *namespaceNode) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	node := nsNode.NewInode(ctx, &groupInode{
		group:     name,
		namespace: nsNode.namespace,
	}, fs.StableAttr{
		Mode: syscall.S_IFDIR,
	})

	// TODO implement create group

	return node, 0
}
