// Package config embeds canonical configuration assets shared across Go and Python tooling.
package config

import _ "embed"

// ModelsJSON is the canonical multi-client model and preset catalog from config/models.json.
//
//go:embed models.json
var ModelsJSON []byte
