package database

import (
	"database/sql"
)

type TaskResponse struct {
	ID      int64  `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func (s Service) AddTask(date, title, comment, repeat string) (int64, error) {
	query := `INSERT INTO scheduler
		(date, title, comment, repeat)
		VALUES (:date, :title, :comment, :repeat)`
	result, err := s.DB.Exec(query,
		sql.Named("date", date),
		sql.Named("title", title),
		sql.Named("comment", comment),
		sql.Named("repeat", repeat))
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s Service) GetNearTask(limit int) ([]TaskResponse, error) {
	var tasks []TaskResponse
	query := `
		SELECT *
		FROM scheduler
		ORDER BY date
		LIMIT :limit
	`
	rows, err := s.DB.Query(query, sql.Named("limit", limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var task TaskResponse
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (s Service) SearchByDate(date string, limit int) ([]TaskResponse, error) {
	var tasks []TaskResponse
	query := `
		SELECT *
		FROM scheduler
		WHERE date = :date
		ORDER BY date
		LIMIT :limit
	`
	rows, err := s.DB.Query(query, sql.Named("date", date), sql.Named("limit", limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var task TaskResponse
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (s Service) SearchByTitle(search string, limit int) ([]TaskResponse, error) {
	var tasks []TaskResponse
	query := `
		SELECT * 
		FROM scheduler 
		WHERE (title LIKE :search COLLATE NOCASE OR comment LIKE :search COLLATE NOCASE)
		ORDER BY date
		LIMIT :limit
	`
	search = "%" + search + "%"
	rows, err := s.DB.Query(query, sql.Named("search", search), sql.Named("limit", limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var task TaskResponse
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}
func (s Service) GetTaskByID(id int64) (TaskResponse, error) {
	var task TaskResponse
	query := `
		SELECT *
		FROM scheduler
		WHERE id = :id
	`
	row := s.DB.QueryRow(query, sql.Named("id", id))
	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return TaskResponse{}, err
	}

	return task, nil
}

func (s Service) UpdateTask(reg TaskResponse) error {
	query := `
		UPDATE scheduler
		SET date = :date, title = :title, comment = :comment, repeat = :repeat
		WHERE id = :id
	`
	result, err := s.DB.Exec(query,
		sql.Named("id", reg.ID),
		sql.Named("date", reg.Date),
		sql.Named("title", reg.Title),
		sql.Named("comment", reg.Comment),
		sql.Named("repeat", reg.Repeat))
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s Service) DeleteTask(id int64) error {
	query := `
		DELETE FROM scheduler
		WHERE id = :id
	`
	_, err := s.DB.Exec(query, sql.Named("id", id))
	if err != nil {
		return err
	}
	return nil
}
