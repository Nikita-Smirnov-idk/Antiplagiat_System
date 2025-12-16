package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"api_gateway/internal/infrastructure/text_extractor"
	"api_gateway/internal/infrastructure/wordcloud"

	plagiarismpb "github.com/Nikita-Smirnov-idk/plagiarism-service/contracts/gen/go"
	storagepb "github.com/Nikita-Smirnov-idk/storage-service/contracts/gen/go"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc/status"
)

type Server struct {
	logger          *slog.Logger
	httpServer      *http.Server
	storageClient   storagepb.StorageClient
	analysisClient  plagiarismpb.PlagiarismClient
	wordCloudClient *wordcloud.QuickChartClient
	textExtractor   *text_extractor.TextExtractor
}

func NewServer(logger *slog.Logger, port int, storage storagepb.StorageClient, analysis plagiarismpb.PlagiarismClient) *Server {
	s := &Server{
		logger:          logger,
		storageClient:   storage,
		analysisClient:  analysis,
		wordCloudClient: wordcloud.NewQuickChartClient(),
		textExtractor:   text_extractor.NewTextExtractor(),
	}

	r := chi.NewRouter()
	r.Post("/api/files", s.handleGenerateUploadURL)
	r.Post("/api/files/verify", s.handleVerifyFile)
	r.Get("/api/files/{task_id}/{student_id}/download", s.handleDownloadURL)
	r.Post("/api/analysis/{task_id}", s.handleAnalyze)
	r.Get("/api/files/{task_id}/{student_id}/wordcloud", s.handleWordCloud)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}
	return s
}

func (s *Server) MustRun() {
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.httpServer.Shutdown(ctx)
}

func (s *Server) handleGenerateUploadURL(w http.ResponseWriter, r *http.Request) {
	type generateUploadRequest struct {
		TaskID    string `json:"task_id"`
		StudentID string `json:"student_id"`
	}

	var req generateUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	if req.TaskID == "" || req.StudentID == "" {
		writeError(w, http.StatusBadRequest, "task_id and student_id are required")
		return
	}

	ctx := r.Context()
	resp, err := s.storageClient.GenerateUploadURL(ctx, &storagepb.GenerateUploadURLRequest{
		StudentId: req.StudentID,
		TaskId:    req.TaskID,
	})
	if err != nil {
		writeGrpcError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"upload_url": resp.GetUrl(),
	})
}

func (s *Server) handleVerifyFile(w http.ResponseWriter, r *http.Request) {
	type verifyRequest struct {
		TaskID    string `json:"task_id"`
		StudentID string `json:"student_id"`
	}

	var req verifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	if req.TaskID == "" || req.StudentID == "" {
		writeError(w, http.StatusBadRequest, "task_id and student_id are required")
		return
	}

	ctx := r.Context()
	resp, err := s.storageClient.VerifyUploadedFile(ctx, &storagepb.VerifyUploadedFileRequest{
		StudentId: req.StudentID,
		TaskId:    req.TaskID,
	})
	if err != nil {
		writeGrpcError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"file_id": resp.GetFileId(),
	})
}

func (s *Server) handleDownloadURL(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "task_id")
	studentID := chi.URLParam(r, "student_id")
	if taskID == "" || studentID == "" {
		writeError(w, http.StatusBadRequest, "task_id and student_id are required")
		return
	}

	ctx := r.Context()
	resp, err := s.storageClient.GenerateDownloadURL(ctx, &storagepb.GenerateDownloadURLRequest{
		StudentId:  studentID,
		TaskId:     taskID,
		FromInside: false,
	})
	if err != nil {
		writeGrpcError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"url": resp.GetUrl(),
	})
}

func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "task_id")
	if taskID == "" {
		writeError(w, http.StatusBadRequest, "task_id is required")
		return
	}

	ctx := r.Context()
	resp, err := s.analysisClient.GetPlagiarismReport(ctx, &plagiarismpb.GetPlagiarismReportRequest{
		TaskId: taskID,
	})
	if err != nil {
		writeGrpcError(w, err)
		return
	}

	type report struct {
		Student                string    `json:"student"`
		StudentWithSimilarFile string    `json:"student_with_similar_file"`
		MaxSimilarity          float64   `json:"max_similarity"`
		FileHandedOverAt       time.Time `json:"file_handed_over_at,omitempty"`
	}

	var reports []report
	for _, rep := range resp.GetReports() {
		var handed time.Time
		if rep.GetFileHandedOverAt() != nil {
			handed = rep.GetFileHandedOverAt().AsTime()
		}
		reports = append(reports, report{
			Student:                rep.GetStudent(),
			StudentWithSimilarFile: rep.GetStudentWithSimilarFile(),
			MaxSimilarity:          rep.GetMaxSimilarity(),
			FileHandedOverAt:       handed,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"task_id":    taskID,
		"started_at": resp.GetStartedAt().AsTime(),
		"reports":    reports,
	})
}

func (s *Server) handleWordCloud(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "task_id")
	studentID := chi.URLParam(r, "student_id")
	if taskID == "" || studentID == "" {
		writeError(w, http.StatusBadRequest, "task_id and student_id are required")
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "png"
	}
	if format != "png" && format != "svg" {
		format = "png"
	}

	ctx := r.Context()
	downloadResp, err := s.storageClient.GenerateDownloadURL(ctx, &storagepb.GenerateDownloadURLRequest{
		StudentId:  studentID,
		TaskId:     taskID,
		FromInside: true,
	})
	if err != nil {
		writeGrpcError(w, err)
		return
	}

	text, err := s.textExtractor.ExtractFromURL(downloadResp.GetUrl())
	if err != nil {
		s.logger.Error("failed to extract text from file", "error", err, "task_id", taskID, "student_id", studentID)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to extract text from file: %v", err))
		return
	}

	if text == "" {
		writeError(w, http.StatusBadRequest, "file is empty or contains no text")
		return
	}

	wordCloudReq := wordcloud.WordCloudRequest{
		Text:            text,
		Format:          format,
		Width:           1000,
		Height:          1000,
		FontScale:       15,
		Scale:           "linear",
		MaxNumWords:     200,
		MinWordLength:   3,
		RemoveStopwords: true,
		Language:        "ru",
	}

	imageData, contentType, err := s.wordCloudClient.GenerateWordCloud(wordCloudReq)
	if err != nil {
		s.logger.Error("failed to generate word cloud", "error", err)
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to generate word cloud: %v", err))
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"wordcloud_%s_%s.%s\"", taskID, studentID, format))
	w.WriteHeader(http.StatusOK)
	w.Write(imageData)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeGrpcError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	switch st.Code() {
	case 3:
		writeError(w, http.StatusBadRequest, st.Message())
	case 5:
		writeError(w, http.StatusNotFound, st.Message())
	case 9, 10, 11, 14:
		writeError(w, http.StatusBadGateway, st.Message())
	default:
		writeError(w, http.StatusInternalServerError, st.Message())
	}
}
