# simple-file-sync

A simple file synchronization tool built in Go that handles file add and modify events from client to server. The application consists of a server that receives files via HTTP and a client that watches local directories and uploads changes.

Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.

## Working Effectively

### Prerequisites and Setup
- Ensure Go 1.22+ is installed (`go version` should return 1.22 or higher)
- No additional dependencies required beyond Go standard tools

### Building the Project
- `go mod tidy` -- downloads dependencies (takes ~10 seconds on first run)
- `go build -o simple-file-sync .` -- builds Linux binary. NEVER CANCEL: First build takes 15 seconds, subsequent builds take ~1 second. Set timeout to 30+ seconds.
- `GOOS=windows GOARCH=amd64 go build -o simple-file-sync.exe .` -- cross-compile for Windows (takes ~12 seconds first time, ~1 second subsequent)
- `GOOS=darwin GOARCH=amd64 go build -o simple-file-sync-mac .` -- cross-compile for macOS (takes ~12 seconds first time, ~1 second subsequent)

### Development and Validation
- `go vet ./...` -- static analysis, completes in ~2 seconds
- `go fmt ./...` -- code formatting, completes in ~1 second  
- `~/go/bin/golint ./...` -- linting (install first: `go install golang.org/x/lint/golint@latest`)
- **No existing test files** -- the project has no unit tests, rely on manual validation scenarios

### Running the Application

#### Server Mode
- `./simple-file-sync server --port=8120 --limit-dir=/path/to/allowed/directory --token=yourtoken`
- Default port: 8120, default token: "kfcvme50"
- Server accepts file uploads at `/receiver` endpoint via HTTP POST
- Files can only be uploaded to paths within the `--limit-dir` directory

#### Client Mode  
- `./simple-file-sync client --local-dir=/source/path --remote-dir=/target/path --server-addr=http://host:port/receiver --server-token=yourtoken --mode=all`
- `--mode=all` syncs all files in the local directory
- `--mode=git` syncs only files that have changed according to `git diff origin --name-only`
- Client watches for file system events and uploads changes automatically

## Validation

### CRITICAL: Manual End-to-End Testing
Always perform complete validation scenarios after making changes. The project has no automated tests, so manual validation is essential.

#### Complete Validation Scenario
1. Create test directories:
   ```bash
   mkdir -p /tmp/server-test /tmp/client-test
   echo "test content" > /tmp/client-test/test.txt
   ```

2. Start server in background:
   ```bash
   ./simple-file-sync server --port=8120 --limit-dir=/tmp/server-test --token=testtoken &
   ```

3. Run client to sync files:
   ```bash
   ./simple-file-sync client --local-dir=/tmp/client-test --remote-dir=/tmp/server-test --server-addr=http://127.0.0.1:8120/receiver --server-token=testtoken --mode=all &
   ```

4. Verify synchronization:
   ```bash
   sleep 3
   ls -la /tmp/server-test/  # Should show synced files
   cat /tmp/server-test/test.txt  # Should match original content
   ```

5. Test live file watching:
   ```bash
   echo "new content" > /tmp/client-test/live-test.txt
   sleep 2
   cat /tmp/server-test/live-test.txt  # Should show "new content"
   ```

6. Clean up processes:
   ```bash
   pkill -f simple-file-sync
   ```

#### Git Mode Validation
For git mode testing, you need a git repository with commits and proper remote setup. The git mode uses `git diff origin --name-only` which requires the origin remote to exist as a branch reference:
```bash
cd /path/to/git/repo
git remote -v  # Verify 'origin' remote exists
git branch -r  # Verify origin branch exists (e.g., origin/main)
# Note: git mode may fail if origin branch doesn't exist locally
./simple-file-sync client --mode=git --local-dir=. --remote-dir=/tmp/server-test --server-addr=http://127.0.0.1:8120/receiver --server-token=testtoken
```
**Note**: Git mode functionality depends on having a properly initialized git repository with commits and remote tracking branches. For testing purposes, use `--mode=all` which always works.

### Build Validation
Always run these commands before committing changes:
- `go vet ./...` -- must complete without errors
- `go fmt ./...` -- ensures consistent formatting
- `go build -o simple-file-sync .` -- must build successfully
- Execute the complete validation scenario above

## Navigation and Code Structure

### Key Project Areas
- `main.go` -- Application entry point, calls cmd.Execute()
- `cmd/` -- CLI command definitions using Cobra framework
  - `cmd/root.go` -- Base command setup and global flags
  - `cmd/server.go` -- Server command with port, token, limit-dir flags
  - `cmd/client.go` -- Client command with local-dir, remote-dir, server-addr, mode flags
- `client/` -- Client implementation
  - `client/client.go` -- File watching with fsnotify, upload logic, git diff functionality
- `server/` -- Server implementation  
  - `server/server.go` -- HTTP server, file upload handler, path validation
- `docs/` -- Auto-generated CLI documentation (do not edit manually)

### CLI Command Reference
```bash
./simple-file-sync --help                    # Show all available commands
./simple-file-sync server --help             # Server-specific options
./simple-file-sync client --help             # Client-specific options  
./simple-file-sync completion bash           # Generate bash completion
```

### Development Tips
- **Always build before testing**: Run `go build -o simple-file-sync .` after making changes
- **Check logs**: Server and client output helpful logging; watch console for upload progress and errors
- **File permissions matter**: Ensure target directories are writable and source directories are readable
- **Module path**: The project uses module path `github.com/wudanyang6/simple-file-sync` (note the `6` in username)
- **Binary cleanup**: Remove build artifacts with `rm simple-file-sync simple-file-sync.exe simple-file-sync-mac` before committing

### Important Constants and Configuration
- Default server port: 8120
- Default token: "kfcvme50" (defined in cmd/server.go)
- Number of upload workers: 5 (NumWorkers in client/client.go)
- Server endpoint: `/receiver`
- Upload modes: "all" (sync everything) or "git" (sync git diff files only)

### Common File Interaction Patterns
- Server validates uploaded file paths must be within limit-dir
- Client uses fsnotify for file system event monitoring
- Files uploaded via multipart form with "file" field and "target" parameter
- Authentication via "token" form parameter
- Client skips hidden directories (starting with ".")

## Troubleshooting

### Common Issues
- **Build fails**: Ensure Go 1.22+ is installed and `go mod tidy` has been run
- **Git mode not working**: Verify you're in a git repository with 'origin' remote configured
- **File sync not working**: Check server logs, verify token matches, ensure target path is within limit-dir
- **Permission denied**: Ensure limit-dir and target directories are writable
- **Client not detecting changes**: Verify local-dir path exists and is readable

### Timing Expectations  
- Build: 15 seconds first time, ~1 second subsequent builds
- Module download: 10 seconds first time, cached afterward  
- Server startup: Immediate (< 1 second)
- File sync: Near-immediate for small files (< 2 seconds)
- Cross-compilation: 12 seconds first time, ~1 second subsequent builds

### NEVER CANCEL Commands
- `go build` operations -- Set timeout to 30+ seconds, builds complete in ~15 seconds
- `go mod tidy` -- Set timeout to 30+ seconds, completes in ~10 seconds
