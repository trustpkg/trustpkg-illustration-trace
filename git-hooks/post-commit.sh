#!/usr/bin/env bash

if [[ "${TRUSTPKG_VERSION_AMEND:-}" == "1" ]]; then
    exit 0
fi

go run ./cmd/repo-scripts post-commit
