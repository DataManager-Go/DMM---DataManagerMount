package dmfs

import (
	"context"
	"fmt"
	"sync"
	"syscall"
	"time"

	"github.com/DataManager-Go/libdatamanager"
	libdm "github.com/DataManager-Go/libdatamanager"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

const (
	// NoGroupFolder foldername for
	// files without groups
	NoGroupFolder = "all_files"
)

var (
	_ = (fs.NodeReaddirer)((*groupNode)(nil))
	_ = (fs.NodeLookuper)((*groupNode)(nil))
	_ = (fs.NodeUnlinker)((*groupNode)(nil))
	_ = (fs.NodeRenamer)((*groupNode)(nil))
)

type groupNode struct {
	fs.Inode

	namespace            string
	group                string
	isNoGroupPlaceholder bool

	fileMap map[string]*libdm.FileResponseItem

	mx sync.Mutex
}

// Create a new group node
func newGroupNode(namespace, group string) *groupNode {
	groupNode := &groupNode{
		namespace:            namespace,
		group:                group,
		isNoGroupPlaceholder: group == NoGroupFolder,
	}
	groupNode.fileMap = make(map[string]*libdm.FileResponseItem)

	return groupNode
}

func (rootGroupNode *groupNode) getRequestAttributes() libdatamanager.FileAttributes {
	// Don't send any group to get
	// all files without groups
	reqGroup := []string{rootGroupNode.group}
	if rootGroupNode.isNoGroupPlaceholder {
		reqGroup = []string{}
	}

	return libdatamanager.FileAttributes{
		Namespace: rootGroupNode.namespace,
		Groups:    reqGroup,
	}
}

// List files in group
func (rootGroupNode *groupNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	r := make([]fuse.DirEntry, 0)

	err := rootGroupNode.loadfiles(func(name string) {
		r = append(r, fuse.DirEntry{
			Mode: syscall.S_IFREG,
			Name: name,
		})
	})

	if err != nil {
		return nil, syscall.EIO
	}

	return fs.NewListDirStream(r), 0
}

func (rootGroupNode *groupNode) loadfiles(nsCB func(name string)) error {
	files, err := data.loadFiles(rootGroupNode.getRequestAttributes())
	if err != nil {
		return err
	}

	for i := range files {
		rootGroupNode.fileMap[files[i].Name] = &files[i]
		if nsCB != nil {
			nsCB(files[i].Name)
		}
	}

	return nil
}

func (rootGroupNode *groupNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	file, has := rootGroupNode.fileMap[name]
	if !has || file == nil {
		return nil, syscall.ENOENT
	}

	file.Attributes.Namespace = rootGroupNode.namespace

	child := rootGroupNode.GetChild(name)
	if child == nil {
		child = rootGroupNode.NewInode(ctx, &fs.Inode{}, fs.StableAttr{
			Mode: syscall.S_IFREG,
		})
	}

	rootGroupNode.setFileAttrs(file, out)

	return child, 0
}

// Set file attributes for files
func (rootGroupNode *groupNode) setFileAttrs(file *libdatamanager.FileResponseItem, out *fuse.EntryOut) {
	out.Size = uint64(file.Size)

	// Times
	out.Ctime = uint64(file.CreationDate.Unix())
	out.Mtime = out.Ctime
	out.Atime = out.Ctime

	out.Mode = 0600

	out.Uid = data.uid
	out.Gid = data.gid
	out.Owner = fuse.Owner{
		Gid: data.gid,
		Uid: data.uid,
	}

	out.SetEntryTimeout(1 * time.Millisecond)
}

// Delete file
func (rootGroupNode *groupNode) Unlink(ctx context.Context, name string) syscall.Errno {
	file := rootGroupNode.findFile(name)
	if file == nil {
		return syscall.ENOENT
	}

	// Make delete http request
	_, err := data.libdm.DeleteFile("", file.ID, false, rootGroupNode.getRequestAttributes())
	if err != nil {
		printResponseError(err, "deleting file")
		return syscall.EIO
	}

	rootGroupNode.removeFile(name)
	rootGroupNode.GetChild(name).ForgetPersistent()

	return 0
}

func (rootGroupNode *groupNode) Rename(ctx context.Context, name string, newParent fs.InodeEmbedder, newName string, flags uint32) syscall.Errno {
	if newParent == nil {
		fmt.Println("Parent is nil")
		return syscall.EIO
	}

	// Check if new parent is a group folder
	gNode, ok := newParent.EmbeddedInode().Operations().(*groupNode)
	if !ok {
		fmt.Println("Can't move file outside of a group")
		return syscall.EPERM
	}

	file, ok := rootGroupNode.fileMap[name]
	if !ok {
		fmt.Println("file", name, "not found")
		return syscall.ENOENT
	}

	if newName == name {
		if gNode.group == rootGroupNode.group && gNode.namespace == rootGroupNode.namespace {
			fmt.Println("nothing to do")
			return 0
		}
	}

	changes := libdm.FileChanges{}
	var didChanges bool

	// Rename
	if newName != name {
		fmt.Println("rename")
		changes.NewName = newName
		didChanges = true
	}

	// Change namespace
	if gNode.namespace != rootGroupNode.namespace {
		fmt.Println("change namespace")
		changes.NewNamespace = gNode.namespace
		didChanges = true
	}

	// Change group
	if gNode.group != rootGroupNode.group {
		fmt.Println("change group")
		if !gNode.isNoGroupPlaceholder {
			changes.AddGroups = []string{gNode.group}
			didChanges = true
		}

		if !rootGroupNode.isNoGroupPlaceholder {
			changes.RemoveGroups = []string{rootGroupNode.group}
			didChanges = true
		}
	}

	if didChanges {
		_, err := data.libdm.UpdateFile("", file.ID, rootGroupNode.namespace, false, changes)
		if err != nil {
			printResponseError(err, "updating file")
			return syscall.EIO
		}
	}

	return 0
}

// Find file by name
func (rootGroupNode *groupNode) findFile(name string) *libdatamanager.FileResponseItem {
	file, has := rootGroupNode.fileMap[name]
	if !has {
		return nil
	}

	return file
}

func (rootGroupNode *groupNode) removeFile(name string) {
	delete(rootGroupNode.fileMap, name)
	data.removeCachedFile(name, rootGroupNode.namespace)
}
