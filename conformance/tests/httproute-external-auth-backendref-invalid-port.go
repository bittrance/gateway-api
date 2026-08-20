/*
Copyright 2025 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tests

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteExternalAuthBackendRefInvalidPort)
}

var HTTPRouteExternalAuthBackendRefInvalidPort = confsuite.ConformanceTest{
	ShortName:   "HTTPRouteExternalAuthBackendRefInvalidPort",
	Description: "An HTTPRoute with an ExternalAuth filter whose backendRef names an existing Service but specifies a port that does not exist on that Service should have ResolvedRefs=False/BackendNotFound and requests should return 500 or 502",
	Manifests:   []string{"tests/httproute-external-auth-backendref-invalid-port.yaml"},
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportHTTPRouteExternalAuth,
		features.SupportHTTPRouteExternalAuthHTTP,
	},
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		routeNN := types.NamespacedName{Name: "external-auth-backendref-invalid-port", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}

		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

		t.Run("HTTPRoute with ExternalAuth backendRef on a nonexistent port has ResolvedRefs=False/BackendNotFound", func(t *testing.T) {
			// A port that does not exist on the Service is treated the same as a
			// missing Service: the specific port endpoint cannot be resolved.
			kubernetes.HTTPRouteMustHaveCondition(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN, metav1.Condition{
				Type:   string(v1.RouteConditionResolvedRefs),
				Status: metav1.ConditionFalse,
				Reason: string(v1.RouteReasonBackendNotFound),
			})
		})

		t.Run("requests to a route with ExternalAuth on an invalid port return 500 or 502", func(t *testing.T) {
			// The route is misconfigured, so this should not be a 403
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request:  http.Request{Path: "/http/allowed"},
				Response: http.Response{StatusCodes: []int{500, 502}},
			})
		})
	},
}
