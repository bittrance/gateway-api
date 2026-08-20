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
	ConformanceTests = append(ConformanceTests, HTTPRouteExternalAuthUnknownBackendRef)
}

var HTTPRouteExternalAuthUnknownBackendRef = confsuite.ConformanceTest{
	ShortName:   "HTTPRouteExternalAuthUnknownBackendRef",
	Description: "An HTTPRoute with an ExternalAuth filter whose backendRef names a Service that does not exist should have ResolvedRefs=False/BackendNotFound and requests should return 500 or 502",
	Manifests:   []string{"tests/httproute-external-auth-unknown-backendref.yaml"},
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportHTTPRouteExternalAuth,
		features.SupportHTTPRouteExternalAuthHTTP,
	},
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		routeNN := types.NamespacedName{Name: "external-auth-unknown-backendref", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}

		// The route must be accepted by the gateway even though the ExternalAuth
		// backendRef is unresolvable; the ResolvedRefs condition distinguishes the error.
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

		t.Run("HTTPRoute with nonexistent ExternalAuth backendRef has ResolvedRefs=False/BackendNotFound", func(t *testing.T) {
			kubernetes.HTTPRouteMustHaveCondition(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN, metav1.Condition{
				Type:   string(v1.RouteConditionResolvedRefs),
				Status: metav1.ConditionFalse,
				Reason: string(v1.RouteReasonBackendNotFound),
			})
		})

		t.Run("requests to a route with unresolvable ExternalAuth backend return 500 or 502", func(t *testing.T) {
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request:  http.Request{Path: "/http/allowed"},
				Response: http.Response{StatusCodes: []int{500, 502}},
			})
		})
	},
}
