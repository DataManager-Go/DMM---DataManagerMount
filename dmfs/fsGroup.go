package dmfs

import (
	"context"
	"fmt"
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
)

type groupNode struct {
	fs.Inode

	namespace            string
	group                string
	isNoGroupPlaceholder bool

	files []libdm.FileResponseItem
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

		fmt.Println(files[i].Name)
	}

	groupNode.files = files

	return fs.NewListDirStream(r), 0
}

func (groupNode *groupNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	file := groupNode.findFile(name)
	if file == nil {
		return nil, syscall.ENOENT
	}

	file.Attributes.Namespace = groupNode.namespace

	child := groupNode.GetChild(name)
	if child == nil {
		child = groupNode.NewInode(ctx, &fileInode{
			file: file,
		}, fs.StableAttr{
			Mode: syscall.S_IFREG,
		})
	}

	return child, 0
}

func (groupNode *groupNode) findFile(name string) *libdatamanager.FileResponseItem {
	for i := range groupNode.files {
		if groupNode.files[i].Name == name {
			return &groupNode.files[i]
		}
	}

	return nil
}
