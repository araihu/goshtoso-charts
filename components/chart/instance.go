package chart

import (
	"context"
	"fmt"
	"io"
	"reflect"

	"github.com/araihu/goshtoso-charts/components"
)

const uninitializedError = "interactive chart label is required"

// Instance is a renderer-neutral chart component with stable identity.
//
// Its zero value is safe. Kind returns an empty kind and Render returns the
// same validation error as the former zero interactive.Instance.
type Instance struct {
	delegate components.Component
}

// NewInstance wraps component as an Instance without exposing its renderer.
//
// This is an intentional extension point for custom and external chart
// components. A nil or typed-nil component returns the safe zero Instance.
// Wrapping an existing Instance is idempotent.
func NewInstance(component components.Component) Instance {
	if nilComponent(component) {
		return Instance{}
	}
	if instance, ok := component.(Instance); ok {
		return instance
	}
	return Instance{delegate: component}
}

// Kind identifies the wrapped chart component.
func (instance Instance) Kind() components.Kind {
	if instance.delegate == nil {
		return ""
	}
	return instance.delegate.Kind()
}

// Render delegates rendering to the wrapped chart component.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	if instance.delegate == nil {
		return fmt.Errorf("%s", uninitializedError)
	}
	return instance.delegate.Render(ctx, writer)
}

func nilComponent(component components.Component) bool {
	if component == nil {
		return true
	}
	value := reflect.ValueOf(component)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

var _ components.Component = Instance{}
