package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/gofrs/uuid/v5"
)

// Task ...
type Task struct {
	ID           string   `json:"id"`
	Description  string   `json:"description"`
	Note         string   `json:"note"`
	Applications []string `json:"applications"`
}

var tasks = map[string]Task{
	"1": {
		ID:          "1",
		Description: "Сделать финальное задание темы REST API",
		Note:        "Если сегодня сделаю, то завтра будет свободный день. Ура!",
		Applications: []string{
			"VS Code",
			"Terminal",
			"git",
		},
	},
	"2": {
		ID:          "2",
		Description: "Протестировать финальное задание с помощью Postmen",
		Note:        "Лучше это делать в процессе разработки, каждый раз, когда запускаешь сервер и проверяешь хендлер",
		Applications: []string{
			"VS Code",
			"Terminal",
			"git",
			"Postman",
		},
	},
}

// Обработчик для получения всех задач
func getTasks(w http.ResponseWriter, r *http.Request) {
	resp, err := json.Marshal(tasks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resp)
}

// Обработчик для добавления задачи
func addTask(w http.ResponseWriter, r *http.Request) {
	var task Task

	defer r.Body.Close()
	reqBody, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(reqBody, &task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if task.ID == "" {
		// ID должен являться уникальным, так как ID  имеет тип string, то ID
		// выбираю, что будет генерироваться в формате uuid v7 (Version 7 a k-sortable id based on timestamp)
		if uuid, uuidErr := uuid.NewV7(); uuidErr == nil {
			task.ID = uuid.String()
		} else {
			http.Error(w, "Cant create ID.", http.StatusBadRequest)
			return
		}
	}

	if len(task.Applications) == 0 {
		userAgent := r.UserAgent()
		if userAgent == "" {
			http.Error(w, "No Applications data in request, and cant get data from User-Agent Header", http.StatusBadRequest)
			return
		}
		task.Applications = append(task.Applications, userAgent)
	}

	if _, ok := tasks[task.ID]; ok {
		http.Error(w, fmt.Sprintf("Cant create task. ID conflict, task with ID = [%s] already exists.", task.ID), http.StatusBadRequest)
		return
	}

	tasks[task.ID] = task

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

// Обработчик для получения данных о задаче по ее идентификатору
func getTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	task, ok := tasks[id]
	if !ok {
		http.Error(w, fmt.Sprintf("No task found with id = [%s]. Nothing to show.", id), http.StatusBadRequest)
		return
	}

	resp, err := json.Marshal(task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(resp)
}

// Обработчик для удаления задачи по идентификатору
func deleteTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	_, ok := tasks[id]
	if !ok {
		http.Error(w, fmt.Sprintf("No task found with id = [%s]. Nothing to delete.", id), http.StatusBadRequest)
		return
	}

	delete(tasks, id)
	_, ok = tasks[id]
	if ok {
		http.Error(w, fmt.Sprintf("Something goes wrong when we try to delete task with id = [%s].", id), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func main() {
	r := chi.NewRouter()

	r.Get("/tasks", getTasks)
	r.Post("/tasks", addTask)
	r.Get("/tasks/{id}", getTask)
	r.Delete("/tasks/{id}", deleteTask)

	if err := http.ListenAndServe(":8080", r); err != nil {
		fmt.Printf("Ошибка при запуске сервера: %s", err.Error())
		return
	}
}
