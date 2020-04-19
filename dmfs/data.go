package dmfs

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DataManager-Go/libdatamanager"
	libdm "github.com/DataManager-Go/libdatamanager"
	dmConfig "github.com/DataManager-Go/libdatamanager/config"
	"github.com/hanwen/go-fuse/v2/fuse"
)

var data dataStruct

// data provides the data for the fs
type dataStruct struct {
	mounter *Mounter
	config  *dmConfig.Config
	libdm   *libdatamanager.LibDM

	userAttributes   *libdm.UserAttributeDataResponse
	lastUserAttrLoad int64

	gid, uid uint32
}

func initData() {
	data.gid = uint32(os.Getegid())
	data.uid = uint32(os.Getuid())
}

// load user attributes (namespaces, groups)
func (data *dataStruct) loadUserAttributes() error {
	// TODO make cachhe time configureable or even create
	// a genuius way to calculate a reasonable cachetime

	if time.Now().Unix()-5 > data.lastUserAttrLoad {
		fmt.Println("reload")
		var err error
		data.userAttributes, err = data.libdm.GetUserAttributeData()
		if err != nil {
			printResponseError(err, "loading user attributes")
			return err
		}

		data.lastUserAttrLoad = time.Now().Unix()
	}

	return nil
}

func (data *dataStruct) setUserAttr(inAttr *fuse.AttrOut) {
	inAttr.Gid = data.gid
	inAttr.Uid = data.uid
}

func (data *dataStruct) trimmedNS(ns string) string {
	userPrefix := data.libdm.Config.Username + "_"
	if !strings.HasPrefix(ns, userPrefix) {
		return ns
	}

	return ns[len(userPrefix):]
}

func (data *dataStruct) fullNS(ns string) string {
	userPrefix := data.libdm.Config.Username + "_"
	if strings.HasPrefix(ns, userPrefix) {
		return ns
	}

	return userPrefix + ns
}
