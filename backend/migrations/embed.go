// Package migrations embeds the SQL schema migrations and seed data into the binary.
package migrations

import "embed"

//go:embed *.sql
var Schema embed.FS

//go:embed seed/seed.sql
var Seed string
