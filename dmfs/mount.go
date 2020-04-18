package dmfs

import (
	"fmt"

	"github.com/DataManager-Go/libdatamanager"
	dmConfig "github.com/DataManager-Go/libdatamanager/config"
)

// MountOptions options to mount
type MountOptions struct {
	MountPoint string
	Config     *dmConfig.Config
	Libdm      *libdatamanager.LibDM
	Debug      bool
	DebugFS    bool
}

// Mount the fs
func (mopt *MountOptions) Mount() {
	fmt.Println("mount", mopt.Config.File)
}
