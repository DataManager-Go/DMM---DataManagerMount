package dmfs

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/DataManager-Go/libdatamanager"
	dmConfig "github.com/DataManager-Go/libdatamanager/config"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
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
	mountDir := filepath.Clean(mopt.MountPoint)
	fmt.Printf("Mounting on %s\n", mountDir)

	// Create mount dir if not exists
	if err := createMountpoint(mountDir); err != nil {
		log.Fatal(err)
	}

	// Test server availability
	if !mopt.testServer() {
		return
	}

	root := &dmanagerFilesystem{}

	// Mount fs
	server, err := fs.Mount(mountDir, root, mopt.getMountOptions())
	if err != nil {
		log.Fatal(err)
	}

	// Umount fs on interrupt
	exitChan := make(chan bool, 1)
	doneChan := make(chan bool, 1)
	sigChan := make(chan os.Signal, 1)
	go (func() {
		signal.Notify(sigChan, os.Interrupt, os.Kill)
		// Await signal
		sig := <-sigChan

		// Debug & Umount
		fmt.Println("\rReceived", sig) // Print \r to overwrite the ugly ^C

		exitChan <- true

		server.Unmount()

		fmt.Println("Umounted")
		doneChan <- true
	})()

	// Exit if mountpoint was
	// unmounted or process was interrupted
	server.Wait()
	select {
	case <-exitChan:
		<-doneChan
	default:
		fmt.Println("umounted externally\nexiting")
	}
}

// tests if server can be accessed and user is authorized
func (mopt *MountOptions) testServer() bool {
	_, err := mopt.Libdm.GetNamespaces()
	if err != nil {
		fmt.Println("Can't mount:", err)
		return false
	}

	return true
}

// Get the mountoptions for the mount operation
func (mopt *MountOptions) getMountOptions() *fs.Options {
	return &fs.Options{
		MountOptions: fuse.MountOptions{
			Debug:      mopt.DebugFS,
			AllowOther: false,
			FsName:     "Datamanager mount",
			Name:       "dmanager",
		},
	}
}

// create dir if not exsists
func createMountpoint(mountPoint string) error {
	_, err := os.Stat(mountPoint)
	if err != nil {
		return os.Mkdir(mountPoint, 0700)
	}

	return nil
}
