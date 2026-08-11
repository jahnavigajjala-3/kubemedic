package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	observabilityv1alpha1 "github.com/jahnavigajjala-3/kubemedic/api/v1alpha1"
)

// HealthReportSummary is the API-safe representation of a HealthReport.
type HealthReportSummary struct {
	Namespace      string             `json:"namespace,omitempty"`
	PodName        string             `json:"podName,omitempty"`
	Phase          string             `json:"phase,omitempty"`
	Diagnosis      string             `json:"diagnosis,omitempty"`
	Message        string             `json:"message,omitempty"`
	RestartCount   int32              `json:"restartCount,omitempty"`
	LastUpdated    metav1.Time        `json:"lastUpdated,omitempty"`
	Recommendation string             `json:"recommendation,omitempty"`
	Severity       string             `json:"severity,omitempty"`
	Conditions     []metav1.Condition `json:"conditions,omitempty"`
}

type listResponse struct {
	Items []HealthReportSummary `json:"items"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type Server struct {
	Client client.Reader
	server *http.Server
}

func NewServer(c client.Reader, addr string) *Server {
	s := &Server{Client: c}
	s.server = &http.Server{
		Addr:              addr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	return s
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/healthreports", s.handleHealthReports)
	mux.HandleFunc("/api/healthreports/", s.handleHealthReportByName)
	return withCORS(mux)
}

func (s *Server) ListenAndServe() error {
	if s.server == nil {
		return errors.New("api server not initialized")
	}
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

func (s *Server) handleHealthReports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	if r.URL.Path != "/api/healthreports" {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "HealthReport not found"})
		return
	}

	reports, err := s.listHealthReports(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to list HealthReports"})
		return
	}
	writeJSON(w, http.StatusOK, listResponse{Items: reports})
}

func (s *Server) handleHealthReportByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/healthreports")
	path = strings.Trim(path, "/")
	if path == "" {
		s.handleHealthReports(w, r)
		return
	}

	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "HealthReport not found"})
		return
	}

	report, err := s.getHealthReport(r.Context(), parts[0], parts[1])
	if apierrors.IsNotFound(err) {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "HealthReport not found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to read HealthReport"})
		return
	}
	writeJSON(w, http.StatusOK, toSummary(report))
}

func (s *Server) listHealthReports(ctx context.Context) ([]HealthReportSummary, error) {
	list := &observabilityv1alpha1.HealthReportList{}
	if err := s.Client.List(ctx, list); err != nil {
		return nil, err
	}

	items := make([]HealthReportSummary, 0, len(list.Items))
	for _, item := range list.Items {
		items = append(items, toSummary(item))
	}
	return items, nil
}

func (s *Server) getHealthReport(ctx context.Context, namespace, name string) (observabilityv1alpha1.HealthReport, error) {
	report := &observabilityv1alpha1.HealthReport{}
	key := client.ObjectKey{Namespace: namespace, Name: name}
	if err := s.Client.Get(ctx, key, report); err != nil {
		return observabilityv1alpha1.HealthReport{}, err
	}
	return *report, nil
}

func toSummary(report observabilityv1alpha1.HealthReport) HealthReportSummary {
	return HealthReportSummary{
		Namespace:      report.Namespace,
		PodName:        report.Spec.PodName,
		Phase:          report.Status.Phase,
		Diagnosis:      report.Status.Diagnosis,
		Message:        report.Status.Message,
		RestartCount:   report.Status.RestartCount,
		LastUpdated:    report.Status.LastUpdated,
		Recommendation: report.Status.Recommendation,
		Severity:       report.Status.Severity,
		Conditions:     report.Status.Conditions,
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if value == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(value); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode response: %v", err), http.StatusInternalServerError)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
