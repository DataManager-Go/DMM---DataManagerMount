package dmfs

import (
	"context"
	"fmt"
	"syscall"
	"time"

	libdm "github.com/DataManager-Go/libdatamanager"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

/*
* filesystem namespace is supposed to load and interact
* with the groups and to show them as directories
 */

type namespaceNode struct {
	fs.Inode

	nsInfo libdm.Namespaceinfo
}

// var _ = (fs.NodeOnAdder)((*namespaceNode)(nil))
var _ = (fs.NodeRmdirer)((*namespaceNode)(nil))
var _ = (fs.NodeMkdirer)((*namespaceNode)(nil))
var _ = (fs.NodeRenamer)((*namespaceNode)(nil))
var _ = (fs.NodeReaddirer)((*namespaceNode)(nil))
var _ = (fs.NodeLookuper)((*namespaceNode)(nil))

func newNamespaceNode(nsInfo libdm.Namespaceinfo) *namespaceNode {
	return &namespaceNode{
		nsInfo: nsInfo,
	}
}

// On Namespace dir accessed
func (nsNode *namespaceNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	// TODO reload groups in dmfsRoot on reading dir

	r := make([]fuse.DirEntry, len(nsNode.nsInfo.Groups))

	// Load groups
	for i := range nsNode.nsInfo.Groups {
		r[i] = fuse.DirEntry{
			Mode: syscall.S_IFDIR,
			Name: nsNode.nsInfo.Groups[i],
		}
	}

	return fs.NewListDirStream(r), 0
}

// On group looked up
func (nsNode *namespaceNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	found := false
	for gi := range nsNode.nsInfo.Groups {
		if nsNode.nsInfo.Groups[gi] == name {
			found = true
			break
		}
	}

	// Return entry not-found-error
	// if entry wasn't found
	if !found {
		return nil, syscall.ENOENT
	}

	// Reuse group child
	child := nsNode.GetChild(name)

	// create child if not found
	if child == nil {
		v := &groupInode{
			group:                name,
			isNoGroupPlaceholder: name == "no_group",
			namespace:            nsNode.nsInfo.Name,
		}

		child = nsNode.NewInode(ctx, v, fs.StableAttr{
			Mode: syscall.S_IFDIR,
		})
	}

	return child, 0
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
	_, err := data.libdm.DeleteAttribute(libdm.GroupAttribute, nsNode.nsInfo.Name, name)
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

	_, err := data.libdm.UpdateAttribute(libdm.GroupAttribute, nsNode.nsInfo.Name, name, newName)
	if err != nil {
		return syscall.ENOENT
	}

	return 0
}

// On group created
func (nsNode *namespaceNode) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	node := nsNode.NewInode(ctx, &groupInode{
		group:     name,
		namespace: nsNode.nsInfo.Name,
	}, fs.StableAttr{
		Mode: syscall.S_IFDIR,
	})

	// TODO implement create group

	return node, 0
}
