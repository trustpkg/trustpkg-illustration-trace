# trustpkg-illustration-trace
Lightweight image contour extraction and skeleton placeholder generation based on alpha transparency. Powered by Go, with a Node.js wrapper.

## Git hooks

Run `make init` from the repository root to copy hooks from `git-hooks/` into
Git's hooks directory (normally `.git/hooks`). Existing hook files are not
overwritten. The pre-commit hook requires Go, `govulncheck` and `gitleaks` on
`PATH`.

To add a hook, register its Git hook name and Go function in `scripts.Hooks`,
then run the init command again. Hook arguments are forwarded after the hook
name in `os.Args`.
