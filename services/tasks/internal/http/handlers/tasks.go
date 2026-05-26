package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/student/tech-ip-sem2/services/tasks/internal/repository"
	"github.com/student/tech-ip-sem2/services/tasks/internal/service"
)

// sanitizeString удаляет все HTML-теги из строки (простая защита от XSS)
func sanitizeString(input string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(input, "")
}

// getInstanceID возвращает ID текущего инстанса из переменной окружения
func getInstanceID() string {
	id := os.Getenv("INSTANCE_ID")
	if id == "" {
		id = "unknown"
	}
	return id
}

// setInstanceHeader устанавливает заголовок X-Instance-ID
func setInstanceHeader(w http.ResponseWriter) {
	w.Header().Set("X-Instance-ID", getInstanceID())
}

type createTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
}

func CreateTaskHandler(repo repository.TaskRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setInstanceHeader(w)

		var req createTaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if req.Title == "" {
			http.Error(w, "title is required", http.StatusBadRequest)
			return
		}
		safeTitle := sanitizeString(req.Title)
		safeDescription := sanitizeString(req.Description)

		task := service.Task{
			ID:          fmt.Sprintf("t_%03d", time.Now().UnixNano()%1000),
			Title:       safeTitle,
			Description: safeDescription,
			DueDate:     req.DueDate,
			Done:        false,
		}
		if err := repo.Create(r.Context(), task); err != nil {
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)
	}
}

func GetTasksHandler(repo repository.TaskRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setInstanceHeader(w)

		tasks, err := repo.GetAll(r.Context())
		if err != nil {
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasks)
	}
}

func GetTaskHandler(repo repository.TaskRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setInstanceHeader(w)

		id := strings.TrimPrefix(r.URL.Path, "/v1/tasks/")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		task, err := repo.GetByID(r.Context(), id)
		if err != nil {
			if err.Error() == "task not found" {
				http.Error(w, "task not found", http.StatusNotFound)
			} else {
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	}
}

func UpdateTaskHandler(repo repository.TaskRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setInstanceHeader(w)

		id := strings.TrimPrefix(r.URL.Path, "/v1/tasks/")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if title, ok := updates["title"]; ok {
			if str, ok := title.(string); ok {
				updates["title"] = sanitizeString(str)
			}
		}
		if desc, ok := updates["description"]; ok {
			if str, ok := desc.(string); ok {
				updates["description"] = sanitizeString(str)
			}
		}
		if err := repo.Update(r.Context(), id, updates); err != nil {
			if err.Error() == "task not found" {
				http.Error(w, "task not found", http.StatusNotFound)
			} else {
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			}
			return
		}
		task, err := repo.GetByID(r.Context(), id)
		if err != nil {
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	}
}

func DeleteTaskHandler(repo repository.TaskRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setInstanceHeader(w)

		id := strings.TrimPrefix(r.URL.Path, "/v1/tasks/")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		if err := repo.Delete(r.Context(), id); err != nil {
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func SearchTasksHandler(repo repository.TaskRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		setInstanceHeader(w)

		title := r.URL.Query().Get("title")
		if title == "" {
			http.Error(w, "missing title parameter", http.StatusBadRequest)
			return
		}
		tasks, err := repo.SearchByTitle(r.Context(), title)
		if err != nil {
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasks)
	}
}

// HealthHandler – эндпоинт для проверки готовности и живости
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	setInstanceHeader(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
