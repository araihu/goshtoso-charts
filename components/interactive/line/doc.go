// Package line provides the interactive Line chart with chart-specific,
// renderer-neutral names.
//
// New code should import this package and use Config, Series, Data, and Line.
// The parent interactive package remains source compatible during the package
// layout migration; its Line-prefixed types and Line function are not removed
// in this phase. Shared interactive options live in the chart package.
package line
