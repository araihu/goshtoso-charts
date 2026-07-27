// Package siteassets embeds presentation CSS owned by the demo application.
package siteassets

import _ "embed"

// CSS styles the demo catalog shell. Goshtoso component styles remain served
// from assets.Handler() at /assets/styles.css.
//
//go:embed site.css
var CSS string
