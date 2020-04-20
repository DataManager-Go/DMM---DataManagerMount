package dmfs

import (
	"context"
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

	err := groupNode.loadfiles(func(name string) {
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

func (groupNode *groupNode) loadfiles(nsCB func(name string)) error {
	files, err := data.loadFiles(groupNode.getRequestAttributes())
	if err != nil {
		return err
	}

	for i := range files {
		groupNode.fileMap[files[i].Name] = &files[i]
		if nsCB != nil {
			nsCB(files[i].Name)
		}
	}

	return nil
}

func (groupNode *groupNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	file, has := groupNode.fileMap[name]
	if !has || file == nil {
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

	out.SetEntryTimeout(1 * time.Millisecond)
}

// Delete file
func (groupNode *groupNode) Unlink(ctx context.Context, name string) syscall.Errno {
	file := groupNode.findFile(name)
	if file == nil {
		return syscall.ENOENT
	}

	// Make delete http request
	_, err := data.libdm.DeleteFile("", file.ID, false, groupNode.getRequestAttributes())
	if err != nil {
		printResponseError(err, "deleting file")
		return syscall.EIO
	}

	groupNode.removeFile(name)
	groupNode.GetChild(name).ForgetPersistent()

	return 0
}

// Find file by name
func (groupNode *groupNode) findFile(name string) *libdatamanager.FileResponseItem {
	file, has := groupNode.fileMap[name]
	if !has {
		return nil
	}

	return file
}

func (groupNode *groupNode) removeFile(name string) {
	delete(groupNode.fileMap, name)
	data.removeCachedFile(name, groupNode.namespace)
}
