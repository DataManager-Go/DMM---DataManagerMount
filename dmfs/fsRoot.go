package dmfs

import (
	"context"
	"fmt"
	"sync"
	"syscall"
	"time"

	"github.com/DataManager-Go/libdatamanager"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/pkg/errors"
)

/*
* the root fs is supposed to load and interact
* with the namespaces and to show them as directories
* Hierarchy: fsRoot -> fsNamespace -> fsGroup -> fsFile
 */

type rootNode struct {
	fs.Inode

	loading bool
	nsNodes map[string]*namespaceNode

	mx sync.Mutex
}

// Implement required interfaces
var (
	_ = (fs.NodeReaddirer)((*rootNode)(nil))
	_ = (fs.NodeRenamer)((*rootNode)(nil))
	_ = (fs.NodeRmdirer)((*rootNode)(nil))
	_ = (fs.NodeLookuper)((*rootNode)(nil))
	_ = (fs.NodeGetattrer)((*rootNode)(nil))
	_ = (fs.NodeMkdirer)((*rootNode)(nil))
)

var (
	// ErrAlreadyLoading error if a load process is already running
	ErrAlreadyLoading = errors.New("already loading")
)

// Create new root Node
func newRootNode() *rootNode {
	rn := &rootNode{}
	rn.nsNodes = make(map[string]*namespaceNode)
	return rn
}

// On dir access, load namespaces and groups
func (root *rootNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	r := make([]fuse.DirEntry, 0)

	// Load namespaces and groups
	err := root.load(func(name string) {
		r = append(r, fuse.DirEntry{
			Name: name,
			Mode: syscall.S_IFDIR,
		})
	})

	// If already loading, use current cache
	if err != nil && err != ErrAlreadyLoading {
		fmt.Println(err)
		return nil, syscall.EIO
	}

	return fs.NewListDirStream(r), 0
}

// Load groups and namespaces
func (root *rootNode) load(nsCB func(name string)) error {
	if root.loading {
		return ErrAlreadyLoading
	}

	root.mx.Lock()
	root.loading = true

	defer func() {
		// Unlock and set loading to false
		// at the end
		root.loading = false
		root.mx.Unlock()
	}()

	// Use dataStore to retrieve
	// groups and namespaces
	err := data.loadUserAttributes()
	if err != nil {
		return err
	}

	// Loop Namespaces and add childs in as folders
	for _, namespace := range data.userAttributes.Namespace {
		nsName := data.trimmedNS(namespace.Name)

		// Find namespace node
		v, has := root.nsNodes[nsName]
		if !has {
			// Create new if not exists
			root.nsNodes[nsName] = newNamespaceNode(namespace)
		} else {
			// Update groups if exists
			v.updateGroups(namespace.Groups)
		}

		if nsCB != nil {
			nsCB(nsName)
		}
	}

	return nil
}

// Lookup -> something tries to lookup a file (namespace)
func (root *rootNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	// Get cached namespaceInfo from map
	val, has := root.nsNodes[name]
	if !has {
		return nil, syscall.ENOENT
	}

	// Try to reuse child
	child := root.GetChild(name)

	// Create new child if not found
	if child == nil {
		child = root.NewInode(ctx, val, fs.StableAttr{
			Mode: syscall.S_IFDIR,
		})
	}

	root.setRootNodeAttributes(out)

	return child, 0
}

// Delete Namespace if virtual file was unlinked
func (root *rootNode) Rmdir(ctx context.Context, name string) syscall.Errno {
	// Throw error if not exists
	nsNode, exists := root.nsNodes[data.trimmedNS(name)]
	if !exists {
		return syscall.ENOENT
	}

	// skip 2s delay if no groups are available
	skipWait := make(chan bool, 1)
	if len(nsNode.nsInfo.Groups) == 0 {
		skipWait <- true
	}

	// wait 2 seconds to ensure, user didn't cancel
	select {
	case <-ctx.Done():
		return syscall.ECANCELED
	case <-time.After(2 * time.Second):
	case <-skipWait:
	}

	defer func() {
		delete(root.nsNodes, data.trimmedNS(name))
		child := root.GetChild(name)
		if child != nil {
			child.RmAllChildren()
			child.ForgetPersistent()
			root.RmChild(name)
		}
	}()

	// Do delete request
	if _, err := data.libdm.DeleteNamespace(data.fullNS(name)); err != nil {
		printResponseError(err, "rm namespace dir")
		return syscall.EFAULT
	}

	data.lastUserAttrLoad = 0

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
	oldNSName := data.fullNS(name)
	newNSName := data.fullNS(newName)
	root.debug("rename namespace", oldNSName, "->", newNSName)

	// Make rename request
	_, err := data.libdm.UpdateNamespace(oldNSName, newNSName)
	if err != nil {
		printResponseError(err, "rename ns dir")
		return syscall.EIO
	}

	data.lastUserAttrLoad = 0

	// Return success
	return 0
}

// create namespace if ns folder was created
func (root *rootNode) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	// Check if created namespace already exists
	_, h := root.nsNodes[name]
	if h {
		return nil, syscall.EEXIST
	}

	nsName := data.fullNS(name)

	// make create request
	_, err := data.libdm.CreateNamespace(nsName)
	if err != nil {
		printResponseError(err, "create ns dir")
		return nil, syscall.EIO
	}

	nsNode := newNamespaceNode(libdatamanager.Namespaceinfo{
		Name:   nsName,
		Groups: []string{NoGroupFolder},
	})
	root.nsNodes[nsName] = nsNode
	data.userAttributes.Namespace = append(data.userAttributes.Namespace, nsNode.nsInfo)

	go (func() {
		time.Sleep(250 * time.Millisecond)
		root.load(nil)
	})()

	node := root.NewInode(ctx, nsNode, fs.StableAttr{
		Mode: syscall.S_IFDIR,
	})

	return node, 0
}

// Set attributes for files
func (root *rootNode) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	// Set owner/group
	data.setUserAttr(out)
	return 0
}

func (root *rootNode) debug(arg ...interface{}) {
	if data.mounter.Debug {
		fmt.Println(arg...)
	}
}

// Set attributes for namespace folders
func (root *rootNode) setRootNodeAttributes(out *fuse.EntryOut) {
	out.Owner = fuse.Owner{
		Gid: data.gid,
		Uid: data.uid,
	}
	out.Gid = data.gid
	out.Uid = data.uid
	out.Mode = 0700
}
