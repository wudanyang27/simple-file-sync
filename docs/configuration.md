# Configuration File Guide

Simple File Sync now supports TOML configuration files to simplify usage and reduce the need for long command lines.

## Configuration File Locations

The application will automatically look for configuration files in the following order:

1. `simple-file-sync.toml` in the current directory
2. `.simple-file-sync.toml` in the current directory  
3. `config/simple-file-sync.toml` in the current directory
4. `$HOME/.simple-file-sync.toml`

You can also specify a custom config file path using the `--config` flag:
```bash
simple-file-sync --config=/path/to/my-config.toml client
```

## Configuration Format

Here's a complete example configuration file:

```toml
# Client mode: "all" - sync all files, "git" - only sync git diff files
mode = "all"

# Local directory - will default to config file directory if not specified
# local_dir = "/path/to/local/directory"

# Remote directory where files will be uploaded
remote_dir = "/home/users/remote/directory"

# Server address
server_addr = "http://127.0.0.1:8120/receiver"

# Server authentication token
server_token = "your-token-here"

# Ignore patterns (regular expressions) - optional
ignore = [
  "\\.tmp$",
  "\\.log$", 
  "node_modules",
  "build"
]

# Path mappings (regex:target_path_format) - optional
path_mappings = [
  "/src(/.*)\\.js$:/build$1.min.js",
  "/images(/.*):static/images$1"
]
```

## Automatic local_dir Default

**Key Feature**: When `local_dir` is not specified in the configuration file, it will automatically default to the directory containing the configuration file.

For example, if your config file is located at `/home/user/myproject/simple-file-sync.toml` and doesn't specify `local_dir`, then `local_dir` will automatically be set to `/home/user/myproject/`.

This makes it easy to place a configuration file in your project directory and have it automatically sync that directory.

## Command Line Override

All configuration values can be overridden using command line flags:

```bash
# Use config file but override the local directory
simple-file-sync client --local-dir=/different/path

# Override multiple values
simple-file-sync client --local-dir=/path --server-addr=http://other-server:8080
```

## Examples

### Basic Project Sync
Create `simple-file-sync.toml` in your project directory:
```toml
mode = "all"
remote_dir = "/home/user/backup"
server_addr = "http://backup-server:8120/receiver"
server_token = "my-secret-token"
```

Then simply run:
```bash
simple-file-sync client
```

The project directory (where the config file is) will be automatically synced.

### Git-based Sync
```toml
mode = "git"
remote_dir = "/home/user/git-backup"
server_addr = "http://git-server:8120/receiver" 
server_token = "git-token"
```

### Custom Local Directory
```toml
mode = "all"
local_dir = "/home/user/documents"
remote_dir = "/backup/documents"
server_addr = "http://backup-server:8120/receiver"
server_token = "docs-token"
```

## Migration from Command Line

If you were previously using:
```bash
simple-file-sync client --local-dir=/home/user/project --mode=all --remote-dir=/backup/project --server-addr=http://server:8120/receiver --server-token=mytoken
```

You can create `/home/user/project/simple-file-sync.toml`:
```toml
mode = "all"
remote_dir = "/backup/project"
server_addr = "http://server:8120/receiver"
server_token = "mytoken"
```

And then simply run from the project directory:
```bash
simple-file-sync client
```