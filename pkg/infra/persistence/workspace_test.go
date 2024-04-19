package persistence

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/entity"
)

func TestWorkspaceRepository(t *testing.T) {
	t.Run("Create", func(t *testing.T) {
		fakeGDB, sqlMock, err := GetMockDB()
		require.NoError(t, err)
		repo := NewWorkspaceRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		var (
			expectedID, expectedRows uint = 1, 1
			actual                        = entity.Workspace{
				Name:        "mockedWorkspace",
				DisplayName: "mockedDisplayName",
				Backend:     &entity.Backend{ID: 1},
				Owners:      []string{"hua.li", "xiaoming.li"},
			}
		)
		sqlMock.ExpectBegin()
		sqlMock.ExpectExec("INSERT").
			WillReturnResult(sqlmock.NewResult(int64(expectedID), int64(expectedRows)))
		sqlMock.ExpectCommit()
		err = repo.Create(context.Background(), &actual)
		require.NoError(t, err)
		require.Equal(t, expectedID, actual.ID)
	})

	t.Run("Delete existing record", func(t *testing.T) {
		fakeGDB, sqlMock, err := GetMockDB()
		require.NoError(t, err)
		repo := NewWorkspaceRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		var expectedID, expectedRows uint = 1, 1
		sqlMock.ExpectBegin()
		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).
				AddRow(1))
		sqlMock.ExpectExec("UPDATE").
			WillReturnResult(sqlmock.NewResult(int64(expectedID), int64(expectedRows)))
		sqlMock.ExpectCommit()
		err = repo.Delete(context.Background(), expectedID)
		require.NoError(t, err)
	})

	t.Run("Delete not existing record", func(t *testing.T) {
		fakeGDB, sqlMock, err := GetMockDB()
		require.NoError(t, err)
		repo := NewWorkspaceRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		sqlMock.ExpectBegin()
		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id"}))
		err = repo.Delete(context.Background(), 1)
		require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("Update existing record", func(t *testing.T) {
		fakeGDB, sqlMock, err := GetMockDB()
		require.NoError(t, err)
		repo := NewWorkspaceRepository(fakeGDB)

		var (
			expectedID, expectedRows uint = 1, 1
			actual                        = entity.Workspace{
				ID: 1,
			}
		)
		sqlMock.ExpectExec("UPDATE").
			WillReturnResult(sqlmock.NewResult(int64(expectedID), int64(expectedRows)))
		err = repo.Update(context.Background(), &actual)
		require.NoError(t, err)
	})

	t.Run("Update not existing record", func(t *testing.T) {
		fakeGDB, sqlMock, err := GetMockDB()
		require.NoError(t, err)
		repo := NewWorkspaceRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		actual := entity.Workspace{
			Name: "NonExistentWorkspace",
		}
		err = repo.Update(context.Background(), &actual)
		require.ErrorIs(t, err, gorm.ErrMissingWhereClause)
	})

	t.Run("Get", func(t *testing.T) {
		fakeGDB, sqlMock, err := GetMockDB()
		require.NoError(t, err)
		repo := NewWorkspaceRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		var (
			expectedID   uint = 1
			expectedName      = "mockedWorkspace"
		)
		sqlMock.ExpectQuery("SELECT .* FROM `workspace`").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "Backend__id"}).
				AddRow(expectedID, expectedName, 1))

		actual, err := repo.Get(context.Background(), expectedID)
		require.NoError(t, err)
		require.Equal(t, expectedID, actual.ID)
		require.Equal(t, expectedName, actual.Name)
	})

	t.Run("List", func(t *testing.T) {
		fakeGDB, sqlMock, err := GetMockDB()
		require.NoError(t, err)
		repo := NewWorkspaceRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		var (
			expectedIDFirst    uint = 1
			expectedNameFirst       = "mockedWorkspace"
			expectedIDSecond   uint = 2
			expectedNameSecond      = "mockedWorkspace2"
		)
		sqlMock.ExpectQuery("SELECT .* FROM `workspace`").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "name", "Backend__id"}).
					AddRow(expectedIDFirst, expectedNameFirst, 1).
					AddRow(expectedIDSecond, expectedNameSecond, 2))

		actual, err := repo.List(context.Background())
		require.NoError(t, err)
		require.Len(t, actual, 2)
	})
}
