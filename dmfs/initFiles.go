package dmfs

import (
	"context"
	"log"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
)

func (root *dmanagerFilesystem) initFiles(ctx context.Context) {
	root.debug("Init files")

	userData, err := root.libdm.GetUserAttributeData()
	if err != nil {
		log.Fatal(err)
		return
	}

	// Loop Namespaces and add childs in as folders
	for _, namespace := range userData.Namespace {
		nsName := removeNSName(namespace.Name)

		// reuse child
		nsp := root.Inode.GetChild(nsName)

		// Create namespace folder
		if nsp == nil {
			nsp = root.Inode.NewInode(ctx, &fs.Inode{}, fs.StableAttr{
				Mode: syscall.S_IFDIR,
			})
			root.AddChild(nsName, nsp, true)
		}

		// Use a no_group folder for files
		// not associated to a groud
		if len(namespace.Groups) == 0 {
			namespace.Groups = []string{"no_group"}
		}

		// Add groups to namespace
		for _, group := range namespace.Groups {
			gp := nsp.GetChild(group)
			if gp == nil {
				gp = nsp.NewInode(ctx, &fs.Inode{}, fs.StableAttr{
					Mode: syscall.S_IFDIR,
				})
				nsp.AddChild(group, gp, true)
			}
		}
	}

	root.debug("Init files success")
}
