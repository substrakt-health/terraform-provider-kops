package structures

import (
	"reflect"

	"k8s.io/kops/pkg/apis/kops"
)

func ExpandOpenstackMonitor(in map[string]interface{}) kops.OpenstackMonitor {
	if in == nil {
		panic("expand OpenstackMonitor failure, in is nil")
	}
	return kops.OpenstackMonitor{
		Delay: func(in interface{}) *string {
			if reflect.DeepEqual(in, reflect.Zero(reflect.TypeOf(in)).Interface()) {
				return nil
			}
			return func(in interface{}) *string {
				if in == nil {
					return nil
				}
				if _, ok := in.([]interface{}); ok && len(in.([]interface{})) == 0 {
					return nil
				}
				return func(in string) *string {
					return &in
				}(string(ExpandString(in)))
			}(in)
		}(in["delay"]),
		Timeout: func(in interface{}) *string {
			if reflect.DeepEqual(in, reflect.Zero(reflect.TypeOf(in)).Interface()) {
				return nil
			}
			return func(in interface{}) *string {
				if in == nil {
					return nil
				}
				if _, ok := in.([]interface{}); ok && len(in.([]interface{})) == 0 {
					return nil
				}
				return func(in string) *string {
					return &in
				}(string(ExpandString(in)))
			}(in)
		}(in["timeout"]),
		MaxRetries: func(in interface{}) *int {
			if reflect.DeepEqual(in, reflect.Zero(reflect.TypeOf(in)).Interface()) {
				return nil
			}
			return func(in interface{}) *int {
				if in == nil {
					return nil
				}
				if _, ok := in.([]interface{}); ok && len(in.([]interface{})) == 0 {
					return nil
				}
				return func(in int) *int {
					return &in
				}(int(ExpandInt(in)))
			}(in)
		}(in["max_retries"]),
	}
}

func FlattenOpenstackMonitor(in kops.OpenstackMonitor) map[string]interface{} {
	return map[string]interface{}{
		"delay": func(in *string) interface{} {
			return func(in *string) interface{} {
				if in == nil {
					return nil
				}
				return func(in string) interface{} {
					return FlattenString(string(in))
				}(*in)
			}(in)
		}(in.Delay),
		"timeout": func(in *string) interface{} {
			return func(in *string) interface{} {
				if in == nil {
					return nil
				}
				return func(in string) interface{} {
					return FlattenString(string(in))
				}(*in)
			}(in)
		}(in.Timeout),
		"max_retries": func(in *int) interface{} {
			return func(in *int) interface{} {
				if in == nil {
					return nil
				}
				return func(in int) interface{} {
					return FlattenInt(int(in))
				}(*in)
			}(in)
		}(in.MaxRetries),
	}
}