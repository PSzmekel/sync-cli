# sync-cli

sync-cli is simple tool to sync two directories.

## Installation

Use the package manager [pip](https://pip.pypa.io/en/stable/) to install foobar.

```bash
# Build (writes binary to current dir)
go build -o sync-cli ./cmd
```

## Usage

```bash
# Run
./sync-cli -source <source_dir> -target <target_dir> [--delete-missing] [--deep-search]
```
Flags

    -source string
    Absolute or relative path to the source directory; required.

    -target string
    Absolute or relative path to the target directory; required.

    --delete-missing
    If set, files present in target but missing in source are deleted during sync; default is false.

    --deep-search
    If set, the comparison is recursive; without it only the top-level files are compared; default is false.

Or use make (no binary needed)

```bash
# Create or recreate sample trees with nested paths and deterministic mtimes
make manual-setup

# Clean up test dir
make manual-clean

# Show test data tree
make manual-tree

# Deep sync with delete missing files
make run-delete-missing-deep

# Shallow sync with delete missing files
make run-delete-missing-shallow

# Deep sync without delete missing files
make run-no-delete-deep

# Shallow sync without delete missing files
make run-no-delete-shallow

# Run tests
make test

# Run lint
make lint
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first
to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

[MIT](https://choosealicense.com/licenses/mit/)