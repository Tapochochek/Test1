package migrations

import "embed"

//go:embed 001_init.sql
var Files embed.FS
