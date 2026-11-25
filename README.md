# trustpkg-illustration-trace
Lightweight image contour extraction and skeleton placeholder generation based on alpha transparency. Powered by Go, with a Node.js wrapper.

## Git hooks

Run `make init` from the repository root to install the hooks in `.git/hooks`.
This overwrites existing hooks with the same names. The checks require Go,
`govulncheck` and `gitleaks` on `PATH`.

Use `git commit -m "fix: description"` normally. The `commit-msg` hook runs
the checks and validates the message and version. After a successful commit,
the `post-commit` hook bumps `VERSION` and amends that commit. Consecutive
commits do not require a push. If `HEAD` has no `VERSION`, the staged or
working-tree file is used as the initial version.

The amend changes the commit hash. A `post-commit` failure cannot cancel the
original commit: check the error and repair the version before pushing. Signed
commits are not amended automatically; update the version and re-sign them
manually. Merge commits are not amended automatically. Manually amending a
commit runs the hook again and can bump its version again.

`fix` bumps patch, `feat` bumps minor, and `feat!`, `fix!` or `build!` bumps
major. Other supported types leave the version unchanged. Scopes such as
`feat(api)!` are supported. Resolve unrelated changes to `VERSION` before
committing. Re-run `make init` after updating hooks.
