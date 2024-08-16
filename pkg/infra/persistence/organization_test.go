package persistence

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/entity"
)

func TestOrganizationRepository(t *testing.T) {
	t.Run("Create", func(t *testing.T) {
		fakeGDB, sqlMock, err := GetMockDB()
		require.NoError(t, err)
		repo := NewOrganizationRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		var (
			expectedID, expectedRows uint = 1, 1
			actual                        = entity.Organization{
				Name:        "mockedOrganization",
				DisplayName: "mockedDisplayName",
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
		repo := NewOrganizationRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		var expectedID uint = 1
		sqlMock.ExpectBegin()
		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).
				AddRow(1))
		sqlMock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(int64(expectedID), int64(0)))
		sqlMock.ExpectCommit()
		err = repo.Delete(context.Background(), expectedID)
		require.NoError(t, err)
	})

	t.Run("Delete not existing record", func(t *testing.T) {
		fakeGDB, sqlMock, err := GetMockDB()
		require.NoError(t, err)
		repo := NewOrganizationRepository(fakeGDB)
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
		repo := NewOrganizationRepository(fakeGDB)

		var (
			expectedID, expectedRows uint = 1, 1
			actual                        = entity.Organization{
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
		repo := NewOrganizationRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		actual := entity.Organization{
			Name: "NonExistentOrganization",
		}
		err = repo.Update(context.Background(), &actual)
		require.ErrorIs(t, err, gorm.ErrMissingWhereClause)
	})

	t.Run("Get", func(t *testing.T) {
		fakeGDB, sqlMock, err := GetMockDB()
		require.NoError(t, err)
		repo := NewOrganizationRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		var (
			expectedID          uint = 1
			expectedName             = "mockedOrganization"
			expectedDisplayName      = "mockedDisplayName"
		)
		sqlMock.ExpectQuery("SELECT .* FROM `organization`").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "display_name"}).
				AddRow(expectedID, expectedName, expectedDisplayName))

		actual, err := repo.Get(context.Background(), expectedID)
		require.NoError(t, err)
		require.Equal(t, expectedID, actual.ID)
		require.Equal(t, expectedName, actual.Name)
	})

	t.Run("List", func(t *testing.T) {
		fakeGDB, sqlMock, err := GetMockDB()
		require.NoError(t, err)
		repo := NewOrganizationRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		var (
			expectedIDFirst           uint = 1
			expectedNameFirst              = "mockedOrganization"
			expectedDisplayNameFirst       = "mockedDisplayName"
			expectedIDSecond          uint = 2
			expectedNameSecond             = "mockedOrganization2"
			expectedDisplayNameSecond      = "mockedDisplayName2"
		)
		sqlMock.ExpectQuery("SELECT .* FROM `organization`").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "name", "display_name"}).
					AddRow(expectedIDFirst, expectedNameFirst, expectedDisplayNameFirst).
					AddRow(expectedIDSecond, expectedNameSecond, expectedDisplayNameSecond))

		actual, err := repo.List(context.Background())
		require.NoError(t, err)
		require.Len(t, actual, 2)
	})
}
