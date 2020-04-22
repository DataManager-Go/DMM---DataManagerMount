# DMM - DataManagerMount
Mount your datamanager files into your filesystem.

# Prerequisites
1. A kernel with FUSE support
2. A compiled [binary](https://github.com/DataManager-Go/DMM---DataManagerMount/releases).
3. A vaild config and session. You can create one using the [cli client](https://github.com/DataManager-Go/DataManagerCLI)

# Get started
You can mount your dataManager files into your local filesystem using `<dmount> mountPoint`. Thats it.<br>

### Opions
`--config` use a different configuration file, created by the [CLI client](https://github.com/DataManager-Go/DataManagerCLI) or [GUI client](https://github.com/DataManager-Go/DataManagerGUI)<br>
`--debug` view more informations about the client server process<br>
`--debug-fs` view logs for the mount process<br>

# Mapping
Since the way the datamanager stores your files is different than your Operating Systems fs does, the mapping between Dmanager and your OS filetree isn't that easy.<br>
The DManager is built to store files assigned to multiple folders (groups) and your FS (usually) allows to store a file in one folder only. In addition, you (usually) can't have multiple files with the same name in one folder.<br>
The DmanagerFS supports that. This is one of the main differences between Filesystems like ext4 or NTFS and the DataManagerFS.<br>

#### The mapping:
```bash
MountPoint
│   # Namespaces
├── default # The default namespace. Equal to <username>_default
|   |   # Groups
│   ├── all_files # all files in the default namespace
|   |   |   # Files
│   │   ├── file_in_group1_and_ns_default
│   │   ├── some other file
│   │   ├── settings.json
│   │   ├── Shortcuts
│   │   ├── settings.json
│   │   └── shareShortcuts
│   ├── group1 # Group 1
│   │   ├── file_in_group1_and_ns_default
│   │   └── some other file
│   ├── condig_files
│   │   ├── settings.json
│   │   └── Shortcuts
│   └── Projects
│       ├── settings.json
│       └── shareShortcuts
└── androidApps # androidApps namespace
    └── all_files
        └── Whatsapp.apk
```

# Features
- [x] Namespaces
    - Full support
- [x] Groups
    - Full support
- Files
    - [x] Listing
    - [x] Delete
    - [ ] Read
    - [ ] Write
    - [x] Move (rename)
    - [x] Rename