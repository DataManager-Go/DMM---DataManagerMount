package dmfs

import (
	"context"
	"fmt"
	"sync"
	"syscall"

	"github.com/DataManager-Go/libdatamanager"
	libdm "github.com/DataManager-Go/libdatamanager"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

const (
	// NoGroupFolder foldername for
	// files without groups
	NoGroupFolder = "no_group"
)

var (
	_ = (fs.NodeReaddirer)((*groupNode)(nil))
	_ = (fs.NodeLookuper)((*groupNode)(nil))
	_ = (fs.NodeUnlinker)((*groupNode)(nil))
)

type groupNode struct {
	fs.Inode

	namespace            string
	group                string
	isNoGroupPlaceholder bool

	files []libdm.FileResponseItem

	mx sync.Mutex
}

// Create a new group node
func newGroupNode(namespace, group string) *groupNode {
	return &groupNode{
		namespace:            namespace,
		group:                group,
		isNoGroupPlaceholder: group == NoGroupFolder,
	}
}
func (groupNode *groupNode) getRequestAttributes() libdatamanager.FileAttributes {
	// Don't send any group to get
	// all files without groups
	reqGroup := []string{groupNode.group}
	if groupNode.isNoGroupPlaceholder {
		reqGroup = []string{}
	}

	return libdatamanager.FileAttributes{
		Namespace: groupNode.namespace,
		Groups:    reqGroup,
	}
}

// List files in group
func (groupNode *groupNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	fmt.Println("readdir node")
	r := make([]fuse.DirEntry, 0)

	files, err := data.loadFiles(groupNode.getRequestAttributes())
	if err != nil {
		return nil, syscall.EIO
	}

	for i := range files {
		r = append(r, fuse.DirEntry{
			Mode: syscall.S_IFREG,
			Name: files[i].Name,
		})
	}

	groupNode.files = files
	return fs.NewListDirStream(r), 0
}

func (groupNode *groupNode) loadfiles() error {
	fmt.Println("readdir node")

	files, err := data.loadFiles(groupNode.getRequestAttributes())
	if err != nil {
		return err
	}

	groupNode.files = files
	return nil
}

func (groupNode *groupNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	groupNode.mx.Lock()
	defer groupNode.mx.Unlock()

	file := groupNode.findFile(name)
	if file == nil {
		return nil, syscall.ENOENT
	}

	file.Attributes.Namespace = groupNode.namespace

	child := groupNode.GetChild(name)
	if child == nil {
		child = groupNode.NewInode(ctx, &fs.Inode{}, fs.StableAttr{
			Mode: syscall.S_IFREG,
		})
	}

	groupNode.setFileAttrs(file, out)

	return child, 0
}

// Set file attributes for files
func (groupNode *groupNode) setFileAttrs(file *libdatamanager.FileResponseItem, out *fuse.EntryOut) {
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
}

// Delete file
func (groupNode *groupNode) Unlink(ctx context.Context, name string) syscall.Errno {
	f := groupNode.findFile(name)
	if f == nil {
		return syscall.ENOENT
	}
	file := *f

	// Remove local
	groupNode.removeFile(file)

	// Make delete http request
	_, err := data.libdm.DeleteFile("", file.ID, false, groupNode.getRequestAttributes())
	if err != nil {
		printResponseError(err, "deleting file")
		return syscall.EIO
	}

	// FIXME delete correctly from slice

	return 0
}

// Find file by name
func (groupNode *groupNode) findFile(name string) *libdatamanager.FileResponseItem {
	for i := range groupNode.files {
		if groupNode.files[i].Name == name {
			return &groupNode.files[i]
		}
	}

	return nil
}

// Remove file from list
func (groupNode *groupNode) removeFile(file libdatamanager.FileResponseItem) bool {
	groupNode.mx.Lock()
	defer groupNode.mx.Unlock()

	for i := range groupNode.files {
		if groupNode.files[i].ID == file.ID {
			groupNode.files = removeFileByIndex(groupNode.files, i)
			return true
		}
	}

	return false
}
