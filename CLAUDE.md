# Globe Expedition Journal - Development Configuration

## Environment Setup

Before running any commands, set up the PATH:

```bash
export PATH="/c/Program Files (x86)/GnuWin32/bin:/c/Users/rober/sdk/go1.25.6/bin:$PATH"
```

Or for PowerShell:
```powershell
$env:PATH = "C:\Users\rober\sdk\go1.25.6\bin;C:\Program Files (x86)\GnuWin32\bin;$env:PATH"
```

## Tools Location

- **Go SDK**: `C:\Users\rober\sdk\go1.25.6\bin`
- **Make**: `C:\Program Files (x86)\GnuWin32\bin\make.exe`

## Quality Gates (MANDATORY)

All code changes MUST pass these gates before task completion:

```bash
make fmt    # Format code - MUST pass
make lint   # Lint code - MUST pass
make test   # Run tests - MUST pass
make build  # Build binary - MUST pass
```

Run `make all` to execute the full pipeline.

## Directory Structure

```
globe-expedition-journal/
├── cmd/server/          # Main entry point
├── internal/
│   ├── api/             # HTTP handlers
│   ├── config/          # Configuration
│   ├── lti/             # LTI 1.3 integration
│   ├── models/          # GORM models
│   └── database/        # DB connection
├── pkg/                 # Public packages
├── frontend/            # Angular app
├── Makefile
└── go.mod
```

## Workflow for Coding Agent

### Before Writing Code
1. Read existing files to understand context
2. Check related tests exist
3. Understand the module structure

### When Writing Code
1. Use Edit tool for modifying existing files
2. Use Write tool only for new files
3. Follow Go conventions (gofmt style)
4. Add tests for new functionality

### After Writing Code
1. Run `make fmt` to format
2. Run `make lint` to check for issues
3. Run `make test` to verify tests pass
4. Run `make build` to ensure compilation
5. If any gate fails, fix before marking task complete

## FORBIDDEN Actions

- Do NOT skip quality gates
- Do NOT commit code that fails `make all`
- Do NOT hardcode secrets or credentials
- Do NOT bypass LTI authentication

## Testing Requirements

- Unit tests required for all handlers
- Unit tests required for all models
- Minimum 70% coverage threshold
- Use table-driven tests where appropriate
