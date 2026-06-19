// Package migrations embeds the SQL migration files so they can be applied
// at process startup without requiring an external migration tool.
package migrations

import "embed"

//go:embed *.sql
var Files embed.FS
