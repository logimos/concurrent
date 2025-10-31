# Publishing Guide

This guide explains how to publish this Go library so others can install it using `go get`.

## Overview

Go modules are published through Git repositories. When you tag a version and push it to GitHub, the Go module proxy (`proxy.golang.org`) automatically indexes it, making it available for anyone to install.

## Prerequisites

1. ✅ Your repository is already configured: `github.com/logimos/concurrent`
2. ✅ Your `go.mod` uses the correct module path
3. ✅ You have a LICENSE file (MIT License)
4. ✅ You have a README.md

## Publishing Steps

### Step 1: Commit All Changes

Make sure all your changes are committed:

```bash
git add .
git commit -m "Prepare for release v1.0.0"
```

### Step 2: Create a Version Tag

Use [Semantic Versioning](https://semver.org/) (vMAJOR.MINOR.PATCH):

- **v1.0.0** - First stable release
- **v1.0.1** - Patch release (bug fixes)
- **v1.1.0** - Minor release (new features, backward compatible)
- **v2.0.0** - Major release (breaking changes)

For your first release, start with `v1.0.0`:

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
```

Or use the provided script:

```bash
./scripts/release.sh v1.0.0
```

### Step 3: Push Code and Tags to GitHub

```bash
git push origin main
git push origin v1.0.0
```

Or push all tags at once:

```bash
git push origin main --tags
```

### Step 4: Verify Publication

After a few minutes, verify your module is available:

```bash
# Check if the module can be fetched
go list -m -versions github.com/logimos/concurrent

# Or try to download it
GOPROXY=direct go get github.com/logimos/concurrent@v1.0.0
```

The module will also appear on [pkg.go.dev](https://pkg.go.dev) within a few hours after the first user requests it.

## How Users Install Your Library

Once published, users can install your library with:

```bash
go get github.com/logimos/concurrent@latest
```

Or a specific version:

```bash
go get github.com/logimos/concurrent@v1.0.0
```

Then they can import it in their code:

```go
import "github.com/logimos/concurrent"
```

## Future Releases

For subsequent releases:

1. Make your changes
2. Update version references if needed
3. Commit changes
4. Create a new tag: `git tag -a v1.1.0 -m "Release v1.1.0"`
5. Push: `git push origin main --tags`

## Best Practices

1. **Always tag releases** - Users need tags to pin specific versions
2. **Use semantic versioning** - Follow MAJOR.MINOR.PATCH format
3. **Write clear release notes** - Document what changed in each version
4. **Keep `go.mod` up to date** - Ensure dependencies are current
5. **Test before tagging** - Run tests and verify everything works

## Troubleshooting

- **Module not found?** - Wait a few minutes after pushing, or use `GOPROXY=direct`
- **Wrong module path?** - Ensure your GitHub repository matches the module path in `go.mod`
- **Tag not showing?** - Make sure you pushed tags with `git push --tags`

## Additional Resources

- [Go Modules Documentation](https://go.dev/ref/mod)
- [Semantic Versioning](https://semver.org/)
- [pkg.go.dev](https://pkg.go.dev) - Official Go package index

