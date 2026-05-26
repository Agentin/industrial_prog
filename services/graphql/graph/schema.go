package graph

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/graphql-go/graphql"
)

func NewSchema(db *sql.DB) (graphql.Schema, error) {
	// Тип Task
	taskType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Task",
		Fields: graphql.Fields{
			"id":          &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
			"title":       &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"description": &graphql.Field{Type: graphql.String},
			"done":        &graphql.Field{Type: graphql.NewNonNull(graphql.Boolean)},
		},
	})

	// CreateTaskInput
	createTaskInput := graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "CreateTaskInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"title":       &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"description": &graphql.InputObjectFieldConfig{Type: graphql.String},
		},
	})

	// UpdateTaskInput
	updateTaskInput := graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "UpdateTaskInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"title":       &graphql.InputObjectFieldConfig{Type: graphql.String},
			"description": &graphql.InputObjectFieldConfig{Type: graphql.String},
			"done":        &graphql.InputObjectFieldConfig{Type: graphql.Boolean},
		},
	})

	// Query
	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"tasks": &graphql.Field{
				Type: graphql.NewList(taskType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					rows, err := db.Query("SELECT id, title, description, done FROM tasks")
					if err != nil {
						return nil, err
					}
					defer rows.Close()
					var tasks []map[string]interface{}
					for rows.Next() {
						var id, title string
						var description sql.NullString
						var done bool
						if err := rows.Scan(&id, &title, &description, &done); err != nil {
							return nil, err
						}
						task := map[string]interface{}{
							"id":    id,
							"title": title,
							"done":  done,
						}
						if description.Valid {
							task["description"] = description.String
						}
						tasks = append(tasks, task)
					}
					return tasks, nil
				},
			},
			"task": &graphql.Field{
				Type: taskType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, ok := p.Args["id"].(string)
					if !ok {
						return nil, nil
					}
					var title string
					var description sql.NullString
					var done bool
					err := db.QueryRow("SELECT title, description, done FROM tasks WHERE id=$1", id).Scan(&title, &description, &done)
					if err == sql.ErrNoRows {
						return nil, nil
					}
					if err != nil {
						return nil, err
					}
					task := map[string]interface{}{
						"id":    id,
						"title": title,
						"done":  done,
					}
					if description.Valid {
						task["description"] = description.String
					}
					return task, nil
				},
			},
		},
	})

	// Mutation
	rootMutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createTask": &graphql.Field{
				Type: taskType,
				Args: graphql.FieldConfigArgument{
					"input": &graphql.ArgumentConfig{Type: graphql.NewNonNull(createTaskInput)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					input, _ := p.Args["input"].(map[string]interface{})
					title, _ := input["title"].(string)
					description := ""
					if desc, ok := input["description"]; ok {
						description, _ = desc.(string)
					}
					id := fmt.Sprintf("t_%03d", time.Now().UnixNano()%1000)
					_, err := db.Exec(
						"INSERT INTO tasks(id, title, description, done) VALUES($1, $2, $3, $4)",
						id, title, description, false,
					)
					if err != nil {
						return nil, err
					}
					return map[string]interface{}{
						"id":          id,
						"title":       title,
						"description": description,
						"done":        false,
					}, nil
				},
			},
			"updateTask": &graphql.Field{
				Type: taskType,
				Args: graphql.FieldConfigArgument{
					"id":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"input": &graphql.ArgumentConfig{Type: graphql.NewNonNull(updateTaskInput)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, _ := p.Args["id"].(string)
					input, _ := p.Args["input"].(map[string]interface{})
					query := "UPDATE tasks SET "
					args := []interface{}{}
					argPos := 1
					if title, ok := input["title"]; ok {
						query += fmt.Sprintf("title = $%d, ", argPos)
						args = append(args, title.(string))
						argPos++
					}
					if description, ok := input["description"]; ok {
						query += fmt.Sprintf("description = $%d, ", argPos)
						args = append(args, description.(string))
						argPos++
					}
					if done, ok := input["done"]; ok {
						query += fmt.Sprintf("done = $%d, ", argPos)
						args = append(args, done.(bool))
						argPos++
					}
					if len(args) == 0 {
						// ничего не меняем, просто вернём текущую задачу из БД
						var title string
						var description sql.NullString
						var done bool
						err := db.QueryRow("SELECT title, description, done FROM tasks WHERE id=$1", id).Scan(&title, &description, &done)
						if err != nil {
							return nil, err
						}
						res := map[string]interface{}{
							"id":    id,
							"title": title,
							"done":  done,
						}
						if description.Valid {
							res["description"] = description.String
						}
						return res, nil
					}
					query = query[:len(query)-2] + fmt.Sprintf(" WHERE id = $%d", argPos)
					args = append(args, id)
					_, err := db.Exec(query, args...)
					if err != nil {
						return nil, err
					}
					// После обновления считываем обновлённые данные
					var title string
					var description sql.NullString
					var done bool
					err = db.QueryRow("SELECT title, description, done FROM tasks WHERE id=$1", id).Scan(&title, &description, &done)
					if err != nil {
						return nil, err
					}
					result := map[string]interface{}{
						"id":    id,
						"title": title,
						"done":  done,
					}
					if description.Valid {
						result["description"] = description.String
					}
					return result, nil
				},
			},
			"deleteTask": &graphql.Field{
				Type: graphql.Boolean,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, _ := p.Args["id"].(string)
					res, err := db.Exec("DELETE FROM tasks WHERE id=$1", id)
					if err != nil {
						return false, err
					}
					rows, _ := res.RowsAffected()
					return rows > 0, nil
				},
			},
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	})
}
