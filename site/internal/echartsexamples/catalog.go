// Package echartsexamples ports the go-echarts/examples catalog into the
// Goshtoso Charts demo. All builders use trusted, static Go-owned values.
package echartsexamples

import "github.com/go-echarts/go-echarts/v2/render"

// Example is one ported upstream example family or variation.
type Example struct {
	Slug   string
	Title  string
	Group  string
	Source string
	Build  func() render.Renderer
}

// All returns every currently ported upstream example in stable catalog order.
func All() []Example {
	result := make([]Example, 0, len(CartesianExamples)+len(HierarchyExamples)+len(StatisticalExamples)+len(ExtensionExamples)+len(WebGLExamples)+len(SupportExamples))
	result = append(result, CartesianExamples...)
	result = append(result, HierarchyExamples...)
	result = append(result, StatisticalExamples...)
	result = append(result, ExtensionExamples...)
	result = append(result, WebGLExamples...)
	result = append(result, SupportExamples...)
	return result
}

// Find returns a ported example by its stable URL slug.
func Find(slug string) (Example, bool) {
	for _, example := range All() {
		if example.Slug == slug {
			return example, true
		}
	}
	return Example{}, false
}
