package dmfs

import (
	"strings"

	libdm "github.com/DataManager-Go/libdatamanager"
)

func removeNsName(ns string) string {
	if !strings.Contains(ns, "_") {
		return ns
	}

	return ns[strings.Index(ns, "_")+1:]
}

func addNsName(ns string, libdm *libdm.RequestConfig) string {
	userPrefix := libdm.Username + "_"
	if strings.HasPrefix(ns, userPrefix) {
		return ns
	}

	return userPrefix + ns
}
