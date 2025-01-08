package server

import (
	"GO_FINAL_PROJECT/internal/database"
	"GO_FINAL_PROJECT/internal/task"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TaskRequest struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type TaskResponse struct {
	ID    int64  `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

const (
	// Limit the number of tasks returned
	limit = 15
)

// SignInHandler handles the sign-in request
func SignInHandler(w http.ResponseWriter, r *http.Request, password string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the JSON body
	var request struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Check if the password matches the one in the environment variable
	if password == "" || request.Password != password {
		http.Error(w, `{"error":"Invalid password"}`, http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := generateJWTToken(password)
	if err != nil {
		http.Error(w, `{"error":"Failed to generate token"}`, http.StatusInternalServerError)
		return
	}
	fmt.Println("Print token for test: ", token)
	// Respond with the token
	response := map[string]string{"token": token}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func AuthMiddleware(next http.HandlerFunc, pass string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Если пароль не задан в переменной окружения, пропускаем аутентификацию
		if len(pass) > 0 {
			// Получаем токен из куки
			cookie, err := r.Cookie("token")
			if err != nil {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Токен из куки
			jwtToken := cookie.Value

			// Проверяем, что токен валиден и не истек
			token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
				// Проверяем, что токен использует правильный метод подписи
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, http.ErrUseLastResponse
				}
				return JWTSecretKey, nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}
			// Вытаскиваем хэш пароля из токена
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				fmt.Println("!!!", claims)
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			// Хэшируем текущий пароль из переменной окружения
			hash := sha256.New()
			hash.Write([]byte(pass))
			expectedHash := hex.EncodeToString(hash.Sum(nil))

			// Проверяем хэш пароля в токене
			if claims["iss"] != expectedHash {
				http.Error(w, "Password changed, token no longer valid", http.StatusUnauthorized)
				return
			}
		}

		// Если токен прошел все проверки, продолжаем обработку запроса
		next(w, r)
	}
}

// nextDateHandler processes requests to /api/nextdate
func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	nowStr := r.URL.Query().Get("now")
	dateStr := r.URL.Query().Get("date")
	repeatStr := r.URL.Query().Get("repeat")

	// Validate required parameters
	if nowStr == "" || dateStr == "" || repeatStr == "" {
		http.Error(w, "Missing required parameters: now, date, or repeat", http.StatusBadRequest)
		return
	}

	// Parse 'now' as time.Time
	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		http.Error(w, "Invalid 'now' parameter: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Call NextDate function
	nextDate, err := task.NextDate(now, dateStr, repeatStr)
	if err != nil {
		http.Error(w, "Error calculating next date: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Return the result
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(nextDate))
}

// TaskHandler processes requests to /api/task
func TaskHandler(w http.ResponseWriter, r *http.Request, s *database.Service) {
	// Check the request method
	switch r.Method {
	case http.MethodPost:
		TaskHandlerPost(w, r, s)
	case http.MethodGet:
		TaskHandlerGet(w, r, s)
	case http.MethodPut:
		TaskHandelerPut(w, r, s)
	case http.MethodDelete:
		TaskHandlerDelete(w, r, s)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

}

// TasksHandler processes requests to /api/tasks
func TasksHandler(w http.ResponseWriter, r *http.Request, s *database.Service) {
	if search := r.URL.Query().Get("search"); search != "" {
		TasksHandlerSearch(w, r, s)
	} else {
		TasksHandlerGetTasks(w, r, s)
	}
}

func TaskHandlerPost(w http.ResponseWriter, r *http.Request, s *database.Service) {
	// Set the response content type
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Parse the request body
	var req TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(TaskResponse{Error: "Invalid JSON: " + err.Error()})
		return
	}

	// Validate the request
	if req.Title == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(TaskResponse{Error: "Title is required"})
		return
	}

	// Validate the date
	if req.Date == "" {
		req.Date = time.Now().Format("20060102")
	} else {
		_, err := time.Parse("20060102", req.Date)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(TaskResponse{Error: "Invalid date format: " + err.Error()})
			return
		}
	}

	// Calculate the next date if repeat is set
	var date string
	if req.Date < time.Now().Format("20060102") {
		if req.Repeat != "" {
			var err error
			date, err = task.NextDate(time.Now(), req.Date, req.Repeat)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(TaskResponse{Error: "Invalid repeat format: " + err.Error()})
				return
			}
		} else {
			date = time.Now().Format("20060102")
		}
	} else {
		date = req.Date
	}
	// Add the task to the database
	id, err := s.AddTask(date, req.Title, req.Comment, req.Repeat)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(TaskResponse{Error: "Failed to add task: " + err.Error()})
		return
	}

	// Return the task ID
	resp := TaskResponse{ID: id}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func TaskHandlerGet(w http.ResponseWriter, r *http.Request, s *database.Service) {
	// Set the response content type
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Get the task ID from the URL
	id := r.URL.Query().Get("id")

	// Validate the ID
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(TaskResponse{Error: "ID is required"})
		return
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(TaskResponse{Error: "Invalid ID: " + err.Error()})
		return
	}

	// Get the task from the database
	task, err := s.GetTaskByID(int64(idInt))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(TaskResponse{Error: "Failed to get task: " + err.Error()})
		return
	}

	// Prepare the response
	response := map[string]string{}

	if task.ID != 0 {
		response = map[string]string{
			"id":      fmt.Sprintf("%d", task.ID),
			"date":    task.Date,
			"title":   task.Title,
			"comment": task.Comment,
			"repeat":  task.Repeat,
		}
	}

	// Return the task
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func TaskHandelerPut(w http.ResponseWriter, r *http.Request, s *database.Service) {
	// Set the response content type
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Parse the request body
	var req TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(TaskResponse{Error: "Invalid JSON: " + err.Error()})
		return
	}

	// Validate the request
	if req.ID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(TaskResponse{Error: "ID is required"})
		return
	}

	// Validate the title
	if req.Title == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(TaskResponse{Error: "Title is required"})
		return
	}

	// Validate the date
	if req.Date == "" {
		req.Date = time.Now().Format("20060102")
	} else {
		_, err := time.Parse("20060102", req.Date)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(TaskResponse{Error: "Invalid date format: " + err.Error()})
			return
		}
	}

	// Calculate the next date if repeat is set
	var date string
	if req.Date < time.Now().Format("20060102") {
		if req.Repeat != "" {
			var err error
			date, err = task.NextDate(time.Now(), req.Date, req.Repeat)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(TaskResponse{Error: "Invalid repeat format: " + err.Error()})
				return
			}
		} else {
			date = time.Now().Format("20060102")
		}
	} else {
		date = req.Date
	}
	id, err := strconv.Atoi(req.ID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(TaskResponse{Error: "Invalid ID: " + err.Error()})
		return
	}
	// Update the task in the database
	err = s.UpdateTask(database.TaskResponse{ID: int64(id), Date: date, Title: req.Title, Comment: req.Comment, Repeat: req.Repeat})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(TaskResponse{Error: "Failed to update task: " + err.Error()})
		return
	}

	// Return empty response
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(TaskResponse{})
}

func TaskHandlerDelete(w http.ResponseWriter, r *http.Request, s *database.Service) {
	// Set the response content type
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Get the task ID from the URL
	id := r.URL.Query().Get("id")

	// Validate the ID
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(TaskResponse{Error: "ID is required"})
		return
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(TaskResponse{Error: "Invalid ID: " + err.Error()})
		return
	}

	// Delete the task from the database
	err = s.DeleteTask(int64(idInt))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(TaskResponse{Error: "Failed to delete task: " + err.Error()})
		return
	}

	// Return empty response
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(TaskResponse{})
}

func TasksHandlerGetTasks(w http.ResponseWriter, r *http.Request, s *database.Service) {
	// Set the response content type
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Get the tasks from the database
	tasks, err := s.GetNearTask(limit)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(TaskResponse{Error: "Failed to get tasks: " + err.Error()})
		return
	}

	// Prepare the response
	response := map[string]interface{}{
		"tasks": []map[string]string{},
	}

	if tasks != nil {
		for _, task := range tasks {
			response["tasks"] = append(response["tasks"].([]map[string]string), map[string]string{
				"id":      fmt.Sprintf("%d", task.ID),
				"date":    task.Date,
				"title":   task.Title,
				"comment": task.Comment,
				"repeat":  task.Repeat,
			})
		}
	} else {
		response["tasks"] = []map[string]string{}
	}

	// Return the tasks
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func TasksHandlerSearch(w http.ResponseWriter, r *http.Request, s *database.Service) {
	// Set the response content type
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Get the search query
	search := r.URL.Query().Get("search")

	// Get the tasks from the database
	var tasks []database.TaskResponse
	var err error
	if parsedDate, parseErr := time.Parse("02.01.2006", search); parseErr == nil {
		// Если это дата, ищем задачи по дате
		tasks, err = s.SearchByDate(parsedDate.Format("20060102"), limit)
	} else {
		// Если это текст, ищем задачи по тексту
		tasks, err = s.SearchByTitle(search, limit)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to search tasks: " + err.Error()})
		return
	}

	// Prepare the response
	response := map[string]interface{}{
		"tasks": []map[string]string{},
	}

	if tasks != nil {
		for _, task := range tasks {
			response["tasks"] = append(response["tasks"].([]map[string]string), map[string]string{
				"id":      fmt.Sprintf("%d", task.ID),
				"date":    task.Date,
				"title":   task.Title,
				"comment": task.Comment,
				"repeat":  task.Repeat,
			})
		}
	} else {
		response["tasks"] = []map[string]string{}
	}

	// Return the tasks
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func TaskHandlerDone(w http.ResponseWriter, r *http.Request, s *database.Service) {
	// Set the response content type
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the task ID from the URL
	id := r.URL.Query().Get("id")

	// Validate the ID
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(TaskResponse{Error: "ID is required"})
		return
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(TaskResponse{Error: "Invalid ID: " + err.Error()})
		return
	}

	//Get the task from the database
	TaskResp, err := s.GetTaskByID(int64(idInt))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(TaskResponse{Error: "Failed to get task: " + err.Error()})
		return
	}

	if TaskResp.ID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(TaskResponse{Error: "Task not found"})
		return
	}

	// Check repeat task
	if TaskResp.Repeat != "" {
		// Calculate the next date
		nextDate, err := task.NextDate(time.Now(), TaskResp.Date, TaskResp.Repeat)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(TaskResponse{Error: "Failed to calculate next date: " + err.Error()})
			return
		}
		// update the task in the database
		err = s.UpdateTask(database.TaskResponse{ID: int64(idInt), Date: nextDate, Title: TaskResp.Title, Comment: TaskResp.Comment, Repeat: TaskResp.Repeat})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(TaskResponse{Error: "Failed to update task: " + err.Error()})
			return
		}

	} else {
		// Delete the task from the database
		err = s.DeleteTask(int64(idInt))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(TaskResponse{Error: "Failed to delete task: " + err.Error()})
			return
		}
	}
	// Return empty response
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(TaskResponse{})
}
