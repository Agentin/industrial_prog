package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
	"github.com/student/tech-ip-sem2/services/tasks/internal/service"
)

type PostgresTaskRepo struct {
	db *sql.DB
}

func NewPostgresTaskRepo(connStr string) (*PostgresTaskRepo, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &PostgresTaskRepo{db: db}, nil
}

func (r *PostgresTaskRepo) Close() error {
	return r.db.Close()
}

func (r *PostgresTaskRepo) Create(ctx context.Context, task service.Task) error {
	query := `INSERT INTO tasks(id, title, description, due_date, done) VALUES($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, task.ID, task.Title, task.Description, task.DueDate, task.Done)
	return err
}

func (r *PostgresTaskRepo) GetAll(ctx context.Context) ([]service.Task, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, title, description, due_date, done FROM tasks`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []service.Task
	for rows.Next() {
		var t service.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.DueDate, &t.Done); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *PostgresTaskRepo) GetByID(ctx context.Context, id string) (service.Task, error) {
	var t service.Task
	query := `SELECT id, title, description, due_date, done FROM tasks WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&t.ID, &t.Title, &t.Description, &t.DueDate, &t.Done)
	if err == sql.ErrNoRows {
		return t, fmt.Errorf("task not found")
	}
	return t, err
}

func (r *PostgresTaskRepo) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	// Динамическое построение запроса с параметрами – безопасно, так как имена полей не из пользовательского ввода
	setParts := []string{}
	args := []interface{}{}
	i := 1
	for k, v := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", k, i))
		args = append(args, v)
		i++
	}
	if len(setParts) == 0 {
		return nil
	}
	query := fmt.Sprintf("UPDATE tasks SET %s WHERE id = $%d",
		strings.Join(setParts, ", "), i)
	args = append(args, id)
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *PostgresTaskRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	return err
}

// SearchByTitle – БЕЗОПАСНАЯ версия (параметризованная)
func (r *PostgresTaskRepo) SearchByTitle(ctx context.Context, title string) ([]service.Task, error) {
	query := `SELECT id, title, description, due_date, done FROM tasks WHERE title = $1`
	rows, err := r.db.QueryContext(ctx, query, title)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []service.Task
	for rows.Next() {
		var t service.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.DueDate, &t.Done); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// !!! Демонстрация уязвимой версии (НЕ ИСПОЛЬЗОВАТЬ В ПРОДУКЦИОННОМ КОДЕ) !!!
// SearchByTitleVulnerable показывает, как НЕ НАДО делать.
func (r *PostgresTaskRepo) SearchByTitleVulnerable(ctx context.Context, title string) ([]service.Task, error) {
	// ОПАСНО: конкатенация строк!
	query := fmt.Sprintf("SELECT id, title, description, due_date, done FROM tasks WHERE title = '%s'", title)
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []service.Task
	for rows.Next() {
		var t service.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.DueDate, &t.Done); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}
