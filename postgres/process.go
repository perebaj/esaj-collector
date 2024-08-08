package postgres

import (
	"context"
	"fmt"
	"time"
)

type Process struct {
	// Identifier of the process on database.
	ID uint64 `db:"id"`
	// The identifier of the process. It follows the parttern 0000000-00.2024.8.26.0000.
	Number string `db:"number"`
	// Is the title of the process on esaj (?)
	Title string `db:"title"`
	// The status of the process. Possible values are 'active' and 'inactive'.
	Status ProcessStatus `db:"status"`
	// Also known as "Cliente/Pasta"
	CustomerName string `db:"customer_name"`
	// Also known as "Ação"
	Action string `db:"action"`
	// Also known as "Foro"
	Forum string `db:"forum"`
	// Date of last movimentation on the process. The time part is 00:00:00 UTC.
	LastMovimentationDate time.Time `db:"last_movimentation_date"`
}

type ProcessStatus string

const (
	ActiveStatus   ProcessStatus = "active"
	Inactivestatus ProcessStatus = "inactive"
)

type ProcessRepository interface {
	Get(ctx context.Context, id uint64) (Process, error)
	Create(ctx context.Context, process Process) (Process, error)
	ByStatus(ctx context.Context, status ProcessStatus) ([]Process, error)
}

func (s Storage) Get(ctx context.Context, id uint64) (Process, error) {
	var process Process
	row := s.db.QueryRowContext(ctx,
		`SELECT id, number, title, status, customer_name, action, forum, last_movimentation_date
	FROM processes WHERE id = $1`,
		id)

	err := row.Scan(&process.ID, &process.Number, &process.Title, &process.Status,
		&process.CustomerName, &process.Action, &process.Forum, &process.LastMovimentationDate)
	if err != nil {
		return Process{}, fmt.Errorf("querying process: %w", err)
	}
	return process, nil
}

func (s Storage) Create(ctx context.Context, process Process) (Process, error) {
	row, err := s.db.ExecContext(ctx,
		`INSERT INTO processes
		(number, title, status, customer_name, action, forum, last_movimentation_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`)
	if err != nil {
		return Process{}, fmt.Errorf("inserting process: %w", err)
	}

	createdProcess := process
	id, err := row.LastInsertId()
	if err != nil {
		return Process{}, fmt.Errorf("retrieving process id: %w", err)
	}

	process.ID = uint64(id)
	return createdProcess, nil
}

func (s Storage) ByStatus(ctx context.Context, status ProcessStatus, page, size uint) ([]Process, error) {
	processes := []Process{}
	err := s.db.SelectContext(ctx,
		&processes,
		`SELECT
			id, number, title, status, customer_name, action, forum, last_movimentation_date
		FROM processes
		ORDER BY id
		LIMIT $1
		OFFSET $2`,
		size, page*size)

	if err != nil {
		return nil, fmt.Errorf("querying processes: %w", err)
	}
	return processes, nil
}
