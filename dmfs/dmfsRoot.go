package dmfs

import (
	"context"
	"fmt"
	"log"
	"syscall"
	"time"

	"github.com/DataManager-Go/libdatamanager"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

type rootNode struct {
	fs.Inode

	nsMap map[string][]string
}

// implement the interfaces
var _ = (fs.NodeReaddirer)((*rootNode)(nil))
var _ = (fs.NodeRenamer)((*rootNode)(nil))
var _ = (fs.NodeRmdirer)((*rootNode)(nil))
var _ = (fs.NodeLookuper)((*rootNode)(nil))

// OnAdd is called on mounting the file system. Use it to populate
// the file system tree.
func (root *rootNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	r := make([]fuse.DirEntry, 0)

	err := data.loadUserAttributes()
	if err != nil {
		log.Fatal(err)
		return nil, syscall.EPERM
	}

	if root.nsMap == nil {
		root.nsMap = make(map[string][]string, 0)
	}

	// Loop Namespaces and add childs in as folders
	for _, namespace := range data.userAttributes.Namespace {
		nsName := removeNSName(namespace.Name)

		_, has := root.nsMap[namespace.Name]
		if !has {
			root.nsMap[namespace.Name] = namespace.Groups
		}

		r = append(r, fuse.DirEntry{
			Name: nsName,
			Mode: syscall.S_IFDIR,
		})
	}

	return fs.NewListDirStream(r), 0
}

func (root *rootNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	fullNS := addNSName(name, data.mounter.Libdm.Config)

	// Get cached namespaceInfo from map
	val, has := root.nsMap[fullNS]
	if !has {
		return nil, syscall.ENOENT
	}

	// Try to reuse child
	child := root.GetChild(name)

	// Create new child if not found
	if child == nil {
		child = root.NewInode(ctx, &namespaceNode{
			nsInfo: &libdatamanager.Namespaceinfo{
				Name:   fullNS,
				Groups: val,
			},
		}, fs.StableAttr{
			Mode: syscall.S_IFDIR,
		})
	}

	return child, 0
}

// Delete Namespace if virtual file was unlinked
func (root *rootNode) Rmdir(ctx context.Context, name string) syscall.Errno {
	namespace := addNSName(name, data.libdm.Config)

	// wait 2 seconds to ensure, user didn't cancel
	select {
	case <-ctx.Done():
		return syscall.ECANCELED
	case <-time.After(2 * time.Second):
	}

	// Do delete request
	if _, err := data.libdm.DeleteNamespace(namespace); err != nil {
		fmt.Println(err)
		return syscall.EFAULT
	}

	return 0
}

// Rename namespace if virtual file was renamed
func (root *rootNode) Rename(ctx context.Context, name string, newParent fs.InodeEmbedder, newName string, flags uint32) syscall.Errno {
	// Don't rename default ns
	if name == "default" {
		fmt.Println("Can't rename default namespace!")
		return syscall.EPERM
	}

	// Get real namespace names
	oldNSName := addNSName(name, data.libdm.Config)
	newNSName := addNSName(newName, data.libdm.Config)
	root.debug("rename namespace", oldNSName, "->", newNSName)

	// Make rename request
	_, err := data.libdm.UpdateNamespace(oldNSName, newNSName)
	if err != nil {
		fmt.Println(err)
		return syscall.ENONET
	}

	// Return success
	return 0
}

func (root *rootNode) debug(arg ...interface{}) {
	if data.mounter.Debug {
		fmt.Println(arg...)
	}
}
