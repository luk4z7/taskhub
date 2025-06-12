package db

import (
	"context"
	"database/sql"
	"errors"
	"regexp" // For sqlmock QueryMatcher
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/luk4z7/taskmanager/domain/task"
	"github.com/luk4z7/taskmanager/domain/user"
	"github.com/stretchr/testify/assert"
)

// TestNewTaskRepository tests the constructor for TaskRepository.
func TestNewTaskRepository(t *testing.T) {
	db, _, err := sqlmock.New() // We don't need the mock for this test, just a valid *sql.DB
	assert.NoError(t, err)
	defer db.Close()

	repo := NewTaskRepository(db)
	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

// Helper function to create a mock DB and the repository
func newMockDbAndRepo(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *TaskRepository) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	assert.NoError(t, err)
	repo := NewTaskRepository(db)
	return db, mock, repo
}

func TestTaskRepository_AddTask(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	testUser := user.User("test@example.com")
	testTask := task.MarshalTask("Test Summary", now, testUser)

	t.Run("Success", func(t *testing.T) {
		db, mock, repo := newMockDbAndRepo(t)
		defer db.Close()

		mock.ExpectPrepare(regexp.QuoteMeta("INSERT INTO task (summary, created_at, created_by) VALUES (?, ?, ?)")).
			ExpectExec().
			WithArgs(testTask.Summary, testTask.CreatedAt(), testTask.CreatedBy()).
			WillReturnResult(sqlmock.NewResult(1, 1)) // LastInsertId=1, RowsAffected=1

		err := repo.AddTask(ctx, testTask)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("PrepareFails", func(t *testing.T) {
		db, mock, repo := newMockDbAndRepo(t)
		defer db.Close()

		mock.ExpectPrepare(regexp.QuoteMeta("INSERT INTO task (summary, created_at, created_by) VALUES (?, ?, ?)")).
			WillReturnError(errors.New("prepare error"))

		err := repo.AddTask(ctx, testTask)
		assert.Error(t, err)
		assert.EqualError(t, err, "prepare error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ExecFails", func(t *testing.T) {
		db, mock, repo := newMockDbAndRepo(t)
		defer db.Close()

		mock.ExpectPrepare(regexp.QuoteMeta("INSERT INTO task (summary, created_at, created_by) VALUES (?, ?, ?)")).
			ExpectExec().
			WithArgs(testTask.Summary, testTask.CreatedAt(), testTask.CreatedBy()).
			WillReturnError(errors.New("exec error"))

		err := repo.AddTask(ctx, testTask)
		assert.Error(t, err)
		assert.EqualError(t, err, "exec error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTaskRepository_List(t *testing.T) {
	ctx := context.Background()

	t.Run("Success_ManagerRole_ReturnsTasks", func(t *testing.T) {
		db, mock, repo := newMockDbAndRepo(t)
		defer db.Close()

		query := regexp.QuoteMeta("SELECT task.id, summary, created_at, created_by FROM task")

		rows := sqlmock.NewRows([]string{"id", "summary", "created_at", "created_by"}).
			AddRow(1, "Task 1", time.Now().Format(dateFormatDefault), "user1@example.com").
			AddRow(2, "Task 2", time.Now().Add(-time.Hour).Format(dateFormatDefault), "user2@example.com")

		mock.ExpectQuery(query).WillReturnRows(rows)

		tasks, err := repo.List(ctx, user.Manager)
		assert.NoError(t, err)
		assert.Len(t, tasks, 2)
		assert.Equal(t, "Task 1", tasks[0].Summary)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success_TechnicianRole_ReturnsTasks", func(t *testing.T) {
		db, mock, repo := newMockDbAndRepo(t)
		defer db.Close()

		// Note: mountQuertByRole appends to the base query.
		// The regexp needs to account for the potentially modified query.
		baseQuery := "SELECT task.id, summary, created_at, created_by FROM task"
		expectedSQL := regexp.QuoteMeta(baseQuery) + `.*` + regexp.QuoteMeta(user.Technician.String())


		rows := sqlmock.NewRows([]string{"id", "summary", "created_at", "created_by"}).
			AddRow(1, "Task 1 Tech", time.Now().Format(dateFormatDefault), "tech1@example.com")

		mock.ExpectQuery(expectedSQL).WillReturnRows(rows)

		tasks, err := repo.List(ctx, user.Technician)
		assert.NoError(t, err)
		assert.Len(t, tasks, 1)
		assert.Equal(t, "Task 1 Tech", tasks[0].Summary)
		assert.NoError(t, mock.ExpectationsWereMet())
	})


	t.Run("Success_ReturnsEmptySlice", func(t *testing.T) {
		db, mock, repo := newMockDbAndRepo(t)
		defer db.Close()

		query := regexp.QuoteMeta("SELECT task.id, summary, created_at, created_by FROM task")
		rows := sqlmock.NewRows([]string{"id", "summary", "created_at", "created_by"}) // No rows added

		mock.ExpectQuery(query).WillReturnRows(rows)

		tasks, err := repo.List(ctx, user.Manager)
		assert.NoError(t, err)
		assert.Len(t, tasks, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("QueryFails", func(t *testing.T) {
		db, mock, repo := newMockDbAndRepo(t)
		defer db.Close()

		query := regexp.QuoteMeta("SELECT task.id, summary, created_at, created_by FROM task")
		mock.ExpectQuery(query).WillReturnError(errors.New("query error"))

		_, err := repo.List(ctx, user.Manager)
		assert.Error(t, err)
		assert.EqualError(t, err, "query error")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ScanFails", func(t *testing.T) {
		db, mock, repo := newMockDbAndRepo(t)
		defer db.Close()

		query := regexp.QuoteMeta("SELECT task.id, summary, created_at, created_by FROM task")
		rows := sqlmock.NewRows([]string{"id", "summary", "created_at", "created_by"}).
			AddRow(1, "Task 1", "invalid-date-format", "user1@example.com") // This will cause time.Parse to fail

		mock.ExpectQuery(query).WillReturnRows(rows)

		_, err := repo.List(ctx, user.Manager)
		assert.Error(t, err)
		// The actual error will be from time.Parse, so check for that.
		// Example: parsing time "invalid-date-format" as "2006-01-02 15:04:05": cannot parse "invalid-date-format" as "2006"
		assert.Contains(t, err.Error(), "cannot parse")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ScanFails_Generic", func(t *testing.T) {
		db, mock, repo := newMockDbAndRepo(t)
		defer db.Close()

		query := regexp.QuoteMeta("SELECT task.id, summary, created_at, created_by FROM task")
		// Simulate an error during scanning the first row by providing incompatible data.
		rows := sqlmock.NewRows([]string{"id", "summary", "created_at", "created_by"}).
			AddRow("not-an-int", "Task 1", time.Now().Format(dateFormatDefault), "user1@example.com")

		mock.ExpectQuery(query).WillReturnRows(rows)

		_, err := repo.List(ctx, user.Manager)
		assert.Error(t, err)
		// The error will be from the database driver trying to convert "not-an-int" to int64
		// For example: sql: Scan error on column index 0, name "id": converting driver.Value type string ("not-an-int") to a int64: invalid syntax
		assert.Contains(t, err.Error(), "converting driver.Value type string")
		assert.Contains(t, err.Error(), "to a int64")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
