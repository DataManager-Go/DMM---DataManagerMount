package dmfs

import (
	"context"
	"fmt"
	"log"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
)

type rootNode struct {
	fs.Inode
}

// implement the interfaces
var _ = (fs.NodeOnAdder)((*rootNode)(nil))
var _ = (fs.NodeRenamer)((*rootNode)(nil))
var _ = (fs.NodeRmdirer)((*rootNode)(nil))

// OnAdd is called on mounting the file system. Use it to populate
// the file system tree.
func (root *rootNode) OnAdd(ctx context.Context) {
	root.debug("Init files")

	err := data.loadUserAttributes()
	if err != nil {
		log.Fatal(err)
		return
	}

	// Loop Namespaces and add childs in as folders
	for _, namespace := range data.userAttributes.Namespace {
		nsName := removeNSName(namespace.Name)

		// reuse child
		nsp := root.Inode.GetChild(nsName)

		// Create namespace folder
		if nsp == nil {
			nsp = root.Inode.NewInode(ctx, &namespaceNode{
				namespace: nsName,
				groups:    namespace.Groups,
			}, fs.StableAttr{
				Mode: syscall.S_IFDIR,
			})

			root.AddChild(nsName, nsp, true)
		}
	}

	root.debug("Init files success")
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
