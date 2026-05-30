package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/student/tech-ip-sem2/shared/rabbit"
	"go.uber.org/zap"
)

type createJobRequest struct {
	TaskID string `json:"task_id"`
}

func CreateJobHandler(publisher *rabbit.Publisher, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setInstanceHeader(w)

		var req createJobRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if req.TaskID == "" {
			http.Error(w, "task_id is required", http.StatusBadRequest)
			return
		}

		job := map[string]interface{}{
			"job":        "process_task",
			"task_id":    req.TaskID,
			"attempt":    1,
			"message_id": uuid.New().String(),
			"created_at": time.Now().UTC().Format(time.RFC3339),
		}

		if err := publisher.PublishJSON(job); err != nil {
			logger.Error("failed to publish job", zap.Error(err))
			http.Error(w, `{"error":"failed to enqueue job"}`, http.StatusInternalServerError)
			return
		}

		logger.Info("job enqueued", zap.String("task_id", req.TaskID))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"status": "accepted", "task_id": req.TaskID})
	}
}
