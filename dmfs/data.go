package dmfs

import (
	"os"

	"github.com/DataManager-Go/libdatamanager"
	libdm "github.com/DataManager-Go/libdatamanager"
	dmConfig "github.com/DataManager-Go/libdatamanager/config"
)

var data dataStruct

// data provides the data for the fs
type dataStruct struct {
	mounter *Mounter
	config  *dmConfig.Config
	libdm   *libdatamanager.LibDM

	userAttributes *libdm.UserAttributeDataResponse
	gid, uid       uint32
}

func initData() {
	data.gid = uint32(os.Getegid())
	data.uid = uint32(os.Getuid())
}

// load user attributes (namespaces, groups)
func (data *dataStruct) loadUserAttributes() error {
	var err error
	data.userAttributes, err = data.libdm.GetUserAttributeData()
	if err != nil {
		return err
	}

	return nil
}
