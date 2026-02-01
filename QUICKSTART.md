# Quick Start Guide

## Prerequisites

| Tool | Version | Notes |
|------|---------|-------|
| Go | 1.25.6 | Required |
| Make | any | Built-in on Mac/Linux |
| Node.js | 18+ | For frontend |
| Git | any | Required |

## 1. Install Go

### macOS
```bash
brew install go@1.25
# Or download from https://go.dev/dl/
```

### Linux (Ubuntu/Debian)
```bash
# Download and extract
wget https://go.dev/dl/go1.25.6.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.25.6.linux-amd64.tar.gz

# Add to ~/.bashrc or ~/.zshrc
export PATH=$PATH:/usr/local/go/bin
```

### Windows
1. Download from https://go.dev/dl/
2. Install to `C:\Users\<you>\sdk\go1.25.6` (or use default location)
3. Install GnuWin32 Make from http://gnuwin32.sourceforge.net/packages/make.htm

## 2. Environment Setup

### macOS / Linux
Go and Make should just work after install. Verify:
```bash
go version    # Should show go1.25.6
make --version
```

If Go isn't found, add to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.):
```bash
export PATH=$PATH:/usr/local/go/bin
export PATH=$PATH:$(go env GOPATH)/bin
```

### Windows (Git Bash)
```bash
source scripts/env.sh
```

### Windows (PowerShell)
```powershell
$env:PATH = "C:\Users\$env:USERNAME\sdk\go1.25.6\bin;C:\Program Files (x86)\GnuWin32\bin;$env:PATH"
```

## 3. Clone & Install Dependencies

```bash
git clone <repo-url>
cd GlobalExpiditionJournal

make deps       # Go dependencies
make fe-deps    # Frontend dependencies
```

## 4. Build & Run

### Backend only
```bash
make all    # Format, lint, test, build
make run    # Build and start server on :8080
```

### Full stack (backend + frontend)
```bash
make demo-all
```
- Backend: http://localhost:8080
- Frontend: http://localhost:8081

## 5. Development Workflow

After making code changes:

```bash
make fmt     # Auto-format code
make lint    # Check for issues
make test    # Run tests
make build   # Compile binary
```

Or run all at once:
```bash
make all
```

**All four gates must pass before committing.**

## 6. Useful Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /api/v1/health` | Health check |
| `GET /api/v1/countries` | List all countries |
| `GET /api/v1/users/:id` | Get user profile |
| `GET /api/v1/visits` | Get visits |

## 7. Project Structure

```
cmd/server/       # Main entry point
internal/
  api/            # HTTP handlers
  config/         # App configuration
  database/       # DB connection
  lti/            # LTI 1.3 auth
  models/         # GORM models
  middleware/     # Auth middleware
  seed/           # Data seeding
frontend/         # React Native (Expo)
```

## 8. Configuration

The app uses SQLite by default (dev mode). No config needed.

| Env Variable | Default | Description |
|--------------|---------|-------------|
| `PORT` | 8080 | Server port |
| `DB_DRIVER` | sqlite | `sqlite` or `postgres` |
| `DATABASE_URL` | globe_expedition.db | DB connection string |
| `DEMO_MODE` | false | Enable demo mode |

## 9. Common Issues

### "make: command not found"

**macOS:** Install Xcode Command Line Tools:
```bash
xcode-select --install
```

**Linux:** Install build-essential:
```bash
sudo apt install build-essential   # Debian/Ubuntu
sudo yum groupinstall "Development Tools"  # RHEL/CentOS
```

**Windows:** Install GnuWin32 Make and run `source scripts/env.sh`

### "go: command not found"

Go isn't in your PATH. See Environment Setup above for your OS.

### Tests failing on fresh clone

Run `make deps` first to download Go modules.

### Frontend won't start

```bash
make fe-deps    # Install node modules
make fe-web     # Start dev server
```

### Permission denied on Linux/Mac

```bash
chmod +x scripts/*.sh
```

### Binary won't run on Mac (Apple Silicon)

If you get architecture errors, rebuild:
```bash
make clean
GOARCH=arm64 make build
```
