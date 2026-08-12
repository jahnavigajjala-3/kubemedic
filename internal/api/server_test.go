package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	observabilityv1alpha1 "github.com/jahnavigajjala-3/kubemedic/api/v1alpha1"
)

func TestHealthReportsList(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := observabilityv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	report := &observabilityv1alpha1.HealthReport{
		ObjectMeta: metav1.ObjectMeta{Name: "nginx", Namespace: "default"},
		Spec:       observabilityv1alpha1.HealthReportSpec{PodName: "nginx"},
		Status: observabilityv1alpha1.HealthReportStatus{
			Namespace:      "default",
			PodName:        "nginx",
			Phase:          "Running",
			Diagnosis:      "Healthy",
			Message:        "All containers healthy",
			RestartCount:   0,
			LastUpdated:    metav1.Now(),
			Recommendation: "No action required",
			Severity:       "Info",
			Conditions: []metav1.Condition{{
				Type:    "Ready",
				Status:  metav1.ConditionTrue,
				Reason:  "Healthy",
				Message: "All containers healthy",
			}},
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(report).Build()
	server := NewServer(fakeClient, ":8080")

	req := httptest.NewRequest(http.MethodGet, "/api/healthreports", nil)
	res := httptest.NewRecorder()
	server.Handler().ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 status, got %d: %s", res.Code, res.Body.String())
	}

	var payload listResponse
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Items) != 1 {
		t.Fatalf("expected one item, got %d", len(payload.Items))
	}
	if payload.Items[0].PodName != "nginx" {
		t.Fatalf("expected pod name nginx, got %q", payload.Items[0].PodName)
	}
	if payload.Items[0].Diagnosis != "Healthy" {
		t.Fatalf("expected diagnosis Healthy, got %q", payload.Items[0].Diagnosis)
	}
}

func TestHealthReportsEmptyList(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := observabilityv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	server := NewServer(fakeClient, ":8080")

	req := httptest.NewRequest(http.MethodGet, "/api/healthreports", nil)
	res := httptest.NewRecorder()
	server.Handler().ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 status, got %d: %s", res.Code, res.Body.String())
	}

	var payload listResponse
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Items) != 0 {
		t.Fatalf("expected empty list, got %d items", len(payload.Items))
	}
}

func TestGetHealthReportByNamespaceAndPod(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := observabilityv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	report := &observabilityv1alpha1.HealthReport{
		ObjectMeta: metav1.ObjectMeta{Name: "nginx", Namespace: "default"},
		Spec:       observabilityv1alpha1.HealthReportSpec{PodName: "nginx"},
		Status: observabilityv1alpha1.HealthReportStatus{
			Namespace:      "default",
			PodName:        "nginx",
			Phase:          "Running",
			Diagnosis:      "Healthy",
			Message:        "OK",
			RestartCount:   2,
			LastUpdated:    metav1.Now(),
			Recommendation: "Keep running",
			Severity:       "Info",
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(report).Build()
	server := NewServer(fakeClient, ":8080")

	req := httptest.NewRequest(http.MethodGet, "/api/healthreports/default/nginx", nil)
	res := httptest.NewRecorder()
	server.Handler().ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 status, got %d: %s", res.Code, res.Body.String())
	}

	var payload HealthReportSummary
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.PodName != "nginx" || payload.Namespace != "default" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestGetHealthReportByNamespaceAndPodNotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := observabilityv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	server := NewServer(fakeClient, ":8080")

	req := httptest.NewRequest(http.MethodGet, "/api/healthreports/default/missing", nil)
	res := httptest.NewRecorder()
	server.Handler().ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected 404 status, got %d: %s", res.Code, res.Body.String())
	}

	var payload errorResponse
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if payload.Error == "" {
		t.Fatal("expected error message")
	}
}

func TestHealthReportsClientErrorProduces500(t *testing.T) {
	server := NewServer(&failingReader{err: errors.New("boom")}, ":8080")

	for _, path := range []string{"/api/healthreports", "/api/healthreports/default/nginx"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		res := httptest.NewRecorder()
		server.Handler().ServeHTTP(res, req)
		if res.Code != http.StatusInternalServerError {
			t.Fatalf("path %s: expected 500 status, got %d: %s", path, res.Code, res.Body.String())
		}
	}
}

func TestHealthReportHandlerCors(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := observabilityv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	server := NewServer(fakeClient, ":8080")

	req := httptest.NewRequest(http.MethodOptions, "/api/healthreports", nil)
	res := httptest.NewRecorder()
	server.Handler().ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected 204 status, got %d: %s", res.Code, res.Body.String())
	}
	if got := res.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected CORS origin *, got %q", got)
	}
}

func TestHealthzEndpoint(t *testing.T) {
	server := NewServer(nil, ":8080")
	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	res := httptest.NewRecorder()
	server.Handler().ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 status for /api/healthz, got %d: %s", res.Code, res.Body.String())
	}
}

type failingReader struct {
	err error
}

func (f *failingReader) Get(_ context.Context, _ client.ObjectKey, _ client.Object, _ ...client.GetOption) error {
	return f.err
}

func (f *failingReader) List(_ context.Context, _ client.ObjectList, _ ...client.ListOption) error {
	return f.err
}

var _ client.Reader = (*failingReader)(nil)

func TestHealthReportSummaryJSONShape(t *testing.T) {
	report := observabilityv1alpha1.HealthReport{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-1", Namespace: "dev"},
		Spec:       observabilityv1alpha1.HealthReportSpec{PodName: "pod-1"},
		Status: observabilityv1alpha1.HealthReportStatus{
			Namespace:      "dev",
			PodName:        "pod-1",
			Phase:          "Running",
			Diagnosis:      "Healthy",
			Message:        "All good",
			RestartCount:   1,
			Recommendation: "No action required",
			Severity:       "Info",
			Conditions:     []metav1.Condition{{Type: "Ready", Status: metav1.ConditionTrue}},
		},
	}

	summary := toSummary(report)
	blob, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("marshal summary: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(blob, &payload); err != nil {
		t.Fatalf("unmarshal summary: %v", err)
	}
	if payload["podName"] != "pod-1" {
		t.Fatalf("expected podName field, got %#v", payload["podName"])
	}
	if payload["diagnosis"] != "Healthy" {
		t.Fatalf("expected diagnosis field, got %#v", payload["diagnosis"])
	}
	if _, ok := payload["conditions"]; !ok {
		t.Fatal("expected conditions field to be present")
	}
	if _, ok := payload["namespace"]; !ok {
		t.Fatal("expected namespace field to be present")
	}
	if _, ok := payload["restartCount"]; !ok {
		t.Fatal("expected restartCount field to be present")
	}
}
