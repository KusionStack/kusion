package persistence

import (
	"context"
	"net/url"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
)

func TestProjectRepository(t *testing.T) {
	mockRemote := "https://github.com/mockorg/mockrepo"
	mockRemoteURL, err := url.Parse(mockRemote)
	require.NoError(t, err)

	t.Run("Create", func(t *testing.T) {
		fakeGDB, sqlMock, err := GetMockDB()
		require.NoError(t, err)
		repo := NewProjectRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		var (
			expectedID, expectedRows uint = 1, 1
			actual                        = entity.Project{
				Name: "mockedProject",
				Source: &entity.Source{
					ID:             1,
					SourceProvider: constant.SourceProviderTypeGithub,
					Remote:         mockRemoteURL,
				},
				Organization: &entity.Organization{
					Name:   "mockedOrg",
					ID:     1,
					Owners: []string{"hua.li", "xiaoming.li"},
				},
				Path:   "/path/to/project",
				Labels: []string{"testLabel"},
				Owners: []string{"hua.li", "xiaoming.li"},
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
		repo := NewProjectRepository(fakeGDB)
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
		repo := NewProjectRepository(fakeGDB)
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
		repo := NewProjectRepository(fakeGDB)

		var (
			expectedID, expectedRows uint = 1, 1
			actual                        = entity.Project{
				ID:   1,
				Name: "mockedProject",
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
		repo := NewProjectRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		actual := entity.Project{
			Name: "mockedProject",
		}
		err = repo.Update(context.Background(), &actual)
		require.ErrorIs(t, err, gorm.ErrMissingWhereClause)
	})

	t.Run("Get", func(t *testing.T) {
		fakeGDB, sqlMock, err := GetMockDB()
		require.NoError(t, err)
		repo := NewProjectRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		var (
			expectedID     uint = 1
			expectedName        = "mockedProject"
			expectedPath        = "/path/to/project"
			expectedOwners      = MultiString{"hua.li", "xiaoming.li"}
		)
		sqlMock.ExpectQuery("SELECT.*FROM `project`").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "path", "Organization__id", "Organization__name", "Organization__owners", "Source__id", "Source__remote", "Source__source_provider"}).
				AddRow(expectedID, expectedName, expectedPath, 1, "mockedOrg", expectedOwners, 1, "https://github.com/test/repo", constant.SourceProviderTypeGithub))

		actual, err := repo.Get(context.Background(), expectedID)
		require.NoError(t, err)
		require.Equal(t, expectedID, actual.ID)
		require.Equal(t, expectedName, actual.Name)
	})

	t.Run("List", func(t *testing.T) {
		fakeGDB, sqlMock, err := GetMockDB()
		require.NoError(t, err)
		repo := NewProjectRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		var (
			expectedID         uint = 1
			expectedName            = "mockedProject"
			expectedPath            = "/path/to/project"
			expectedOrgOwners       = MultiString{"hua.li", "xiaoming.li"}
			expectedIDSecond   uint = 2
			expectedNameSecond      = "mockedProject2"
			expectedPathSecond      = "/path/to/project/2"
		)
		sqlMock.ExpectQuery("SELECT .* FROM `project`").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "name", "path", "Organization__id", "Organization__name", "Organization__owners", "Source__id", "Source__remote", "Source__source_provider"}).
					AddRow(expectedID, expectedName, expectedPath, 1, "mockedOrg", expectedOrgOwners, 1, "https://github.com/test/repo", constant.SourceProviderTypeGithub).
					AddRow(expectedIDSecond, expectedNameSecond, expectedPathSecond, 1, "mockedOrg", expectedOrgOwners, 2, "https://github.com/test/repo2", constant.SourceProviderTypeGithub))

		actual, err := repo.List(context.Background(), &entity.ProjectFilter{})
		require.NoError(t, err)
		require.Len(t, actual, 2)
	})
}
