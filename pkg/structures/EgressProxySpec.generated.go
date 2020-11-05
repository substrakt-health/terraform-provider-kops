package structures

import (
	"k8s.io/kops/pkg/apis/kops"
)

func ExpandEgressProxySpec(in map[string]interface{}) kops.EgressProxySpec {
	if in == nil {
		panic("expand EgressProxySpec failure, in is nil")
	}
	return kops.EgressProxySpec{
		HTTPProxy: func(in interface{}) kops.HTTPProxy {
			return func(in interface{}) kops.HTTPProxy {
				if in.([]interface{})[0] == nil {
					return kops.HTTPProxy{}
				}
				return (ExpandHTTPProxy(in.([]interface{})[0].(map[string]interface{})))
			}(in)
		}(in["http_proxy"]),
		ProxyExcludes: func(in interface{}) string {
			return string(ExpandString(in))
		}(in["proxy_excludes"]),
	}
}

func FlattenEgressProxySpec(in kops.EgressProxySpec) map[string]interface{} {
	return map[string]interface{}{
		"http_proxy": func(in kops.HTTPProxy) interface{} {
			return func(in kops.HTTPProxy) []map[string]interface{} {
				return []map[string]interface{}{FlattenHTTPProxy(in)}
			}(in)
		}(in.HTTPProxy),
		"proxy_excludes": func(in string) interface{} {
			return FlattenString(string(in))
		}(in.ProxyExcludes),
	}
}