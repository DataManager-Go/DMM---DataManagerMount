package dmfs

import (
	"context"
	"syscall"

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

// Implement required interfaces
var (
	_ = (fs.NodeRmdirer)((*namespaceNode)(nil))
	_ = (fs.NodeMkdirer)((*namespaceNode)(nil))
	_ = (fs.NodeRenamer)((*namespaceNode)(nil))
	_ = (fs.NodeReaddirer)((*namespaceNode)(nil))
	_ = (fs.NodeLookuper)((*namespaceNode)(nil))
)

// Create a new ns node from nsInfo
func newNamespaceNode(nsInfo libdm.Namespaceinfo) *namespaceNode {
	return &namespaceNode{
		nsInfo: libdm.Namespaceinfo{
			Name:   nsInfo.Name,
			Groups: formatGroups(nsInfo.Groups),
		},
	}
}

func formatGroups(groups []string) []string {
	if groups == nil || len(groups) == 0 {
		return []string{NoGroupFolder}
	}

	return append(groups, NoGroupFolder)
}

func (nsNode *namespaceNode) updateGroups(groups []string) {
	nsNode.nsInfo.Groups = formatGroups(groups)
}

// On Namespace dir accessed
func (nsNode *namespaceNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	// Reload groups if namespace was accessed
	if _, parent := nsNode.Parent(); parent != nil {
		if dmfsroot, ok := (parent.Operations()).(*rootNode); ok {
			go dmfsroot.load(nil)
		}
	}

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
			isNoGroupPlaceholder: name == NoGroupFolder,
			namespace:            nsNode.nsInfo.Name,
		}

		child = nsNode.NewInode(ctx, v, fs.StableAttr{
			Mode: syscall.S_IFDIR,
		})
	}

	return child, 0
}

// Delete group if vfile was removed
// Note: deleting groups doesn't mean the files are deleted!
func (nsNode *namespaceNode) Rmdir(ctx context.Context, name string) syscall.Errno {
	if name == NoGroupFolder {
		return 0
	}

	// Do http delete request
	_, err := data.libdm.DeleteAttribute(libdm.GroupAttribute, nsNode.nsInfo.Name, name)
	if err != nil {
		printResponseError(err, "rm group dir")
		return syscall.ENOENT
	}

	// Remove group from list
	nsNode.nsInfo.Groups = removeFromStringSlice(nsNode.nsInfo.Groups, name)

	return 0
}

// On group renamed
func (nsNode *namespaceNode) Rename(ctx context.Context, name string, newParent fs.InodeEmbedder, newName string, flags uint32) syscall.Errno {
	if name == NoGroupFolder {
		// TODO add groups to files with no group
		return 0
	}

	// Rename group request
	_, err := data.libdm.UpdateAttribute(libdm.GroupAttribute, nsNode.nsInfo.Name, name, newName)
	if err != nil {
		printResponseError(err, "Updating group")
		return syscall.ENOENT
	}

	return 0
}

// On group created
func (nsNode *namespaceNode) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	// Create group
	_, err := data.libdm.CreateAttribute(libdm.GroupAttribute, nsNode.nsInfo.Name, name)
	if err != nil {
		printResponseError(err, "creating group")
		return nil, syscall.EIO
	}

	nsNode.nsInfo.Groups = append(nsNode.nsInfo.Groups, name)
	node := nsNode.NewInode(ctx, &groupInode{
		group:     name,
		namespace: nsNode.nsInfo.Name,
	}, fs.StableAttr{
		Mode: syscall.S_IFDIR,
	})

	return node, 0
}
