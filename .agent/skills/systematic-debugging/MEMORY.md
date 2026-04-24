# MI-Tech Systematic Debugging Memory

## Build Verification
```bash
# Backend (sandboxed)
export GOMODCACHE=$(pwd)/.gocache/mod
export GOCACHE=$(pwd)/.gocache/build
export GOFLAGS=-buildvcs=false
export CGO_ENABLED=0
/usr/local/go/bin/go build ./... 2>&1 | head -50

# Frontend
/path/to/npm run build    # npm not always in PATH — check make frontend terminal
```

## Common Issues
| Problem | Root Cause | Fix |
|---------|-----------|-----|
| `go: command not found` | Not in agent PATH | Use `/usr/local/go/bin/go` |
| `could not create module cache` | Permission denied | Set `GOMODCACHE` to local `.gocache/mod` |
| `xcode-select error` | No dev tools | Set `CGO_ENABLED=0` |
| `main redeclared` | Two `main()` in same `cmd/` dir | Move to subdirectory package |
| `imported and not used` | Stale import after refactor | Remove the unused import |
| `npm: command not found` | Not in agent PATH | Check running `make frontend` terminal |

## Running Services
- Backend usually running via `make backend` (Air hot-reload — auto-rebuilds)
- Frontend usually running via `make frontend` (Vite dev server)
- Check running terminals before manually building

## Log Locations
- Backend stdout (Air terminal)
- Frontend browser console + Vite terminal
- PostgreSQL: Docker container logs (`docker logs <container>`)
