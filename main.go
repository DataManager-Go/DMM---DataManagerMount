package main

import (
	"fmt"
	"log"
	"os"

	"github.com/DataManager-Go/DMM---DataManagerMount/dmfs"
	"github.com/DataManager-Go/libdatamanager"
	dmConfig "github.com/DataManager-Go/libdatamanager/config"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	appName = "dmount"
	version = "v0.0.1"
)

// ...
const (
	// EnVarPrefix prefix for env vars
	EnVarPrefix = "DMOUNT"

	// EnVarPrefix prefix of all used env vars
	EnVarLogLevel   = "LOG_LEVEL"
	EnVarConfigFile = "CONFIG"
)

// Return the variable using the server prefix
func getEnVar(name string) string {
	return fmt.Sprintf("%s_%s", EnVarPrefix, name)
}

var (
	app        = kingpin.New(appName, "A DataManager")
	appCfgFile = app.Flag("config", "the configuration file for the app").Envar(getEnVar(EnVarConfigFile)).Short('c').String()
	appDebugFS = app.Flag("debug-fs", "Debug the filesystem").Bool()
	appDebug   = app.Flag("debug", "Debug the fs bridge").Bool()

	appMountpoint = app.Arg("mountPoint", "The folder to mount your dm namespaces in").Required().String()
)

func main() {
	app.HelpFlag.Short('h')
	app.Version(version)

	// Prase cli flags
	kingpin.MustParse(app.Parse(os.Args[1:]))

	// Init config
	var err error
	config, err := dmConfig.InitConfig(dmConfig.GetDefaultConfigFile(), *appCfgFile)
	if err != nil {
		log.Fatalln(err)
	}
	if config == nil {
		fmt.Println("New config created")
		return
	}

	// Create mount options
	options := dmfs.Mounter{
		Libdm:      libdatamanager.NewLibDM(config.MustGetRequestConfig()),
		MountPoint: *appMountpoint,
		Config:     config,
		Debug:      *appDebug,
		DebugFS:    *appDebugFS,
	}

	// Monut the fs
	options.Mount()
}
