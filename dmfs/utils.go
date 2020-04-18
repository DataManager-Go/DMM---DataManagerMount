package dmfs

import (
	"strings"

	libdm "github.com/DataManager-Go/libdatamanager"
)

func removeNSName(ns string) string {
	if !strings.Contains(ns, "_") {
		return ns
	}

	return ns[strings.Index(ns, "_")+1:]
}

func addNSName(ns string, libdm *libdm.RequestConfig) string {
	userPrefix := libdm.Username + "_"
	if strings.HasPrefix(ns, userPrefix) {
		return ns
	}

	return userPrefix + ns
}
