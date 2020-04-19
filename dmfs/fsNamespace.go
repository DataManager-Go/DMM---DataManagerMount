package dmfs

import (
	"context"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fs"
)

type namespaceNode struct {
	fs.Inode

	namespace string
	groups    []string
}

var _ = (fs.NodeOnAdder)((*namespaceNode)(nil))
var _ = (fs.NodeRmdirer)((*namespaceNode)(nil))

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

	// TODO do delete request

	return 0
}
