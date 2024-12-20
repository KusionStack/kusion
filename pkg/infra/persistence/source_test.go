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

func TestSourceRepository(t *testing.T) {
	mockRemote := "https://github.com/mockorg/mockrepo"
	mockRemoteURL, err := url.Parse(mockRemote)
	require.NoError(t, err)

	t.Run("Create", func(t *testing.T) {
		fakeGDB, sqlMock, err := GetMockDB()
		require.NoError(t, err)
		repo := NewSourceRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		var (
			expectedID, expectedRows uint = 1, 1
			actual                        = entity.Source{
				Name:           "mockedSource",
				SourceProvider: constant.SourceProviderTypeOCI,
				Remote:         mockRemoteURL,
				Description:    "i am a description",
				Labels:         []string{"testLabel"},
				Owners:         []string{"hua.li", "xiaoming.li"},
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
		repo := NewSourceRepository(fakeGDB)
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
		repo := NewSourceRepository(fakeGDB)
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
		repo := NewSourceRepository(fakeGDB)

		var (
			expectedID, expectedRows uint = 1, 1
			actual                        = entity.Source{
				ID:             1,
				SourceProvider: constant.SourceProviderTypeGithub,
				Remote:         mockRemoteURL,
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
		repo := NewSourceRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		actual := entity.Source{
			SourceProvider: constant.SourceProviderTypeGithub,
			Remote:         mockRemoteURL,
		}
		err = repo.Update(context.Background(), &actual)
		require.ErrorIs(t, err, gorm.ErrMissingWhereClause)
	})

	t.Run("Get", func(t *testing.T) {
		fakeGDB, sqlMock, err := GetMockDB()
		require.NoError(t, err)
		repo := NewSourceRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		var (
			expectedID                 uint = 1
			expectedSourceProviderType      = constant.SourceProviderTypeGithub
			expectedRemote                  = mockRemote
		)
		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "source_provider", "remote"}).
				AddRow(expectedID, string(expectedSourceProviderType), expectedRemote))
		actual, err := repo.Get(context.Background(), expectedID)
		require.NoError(t, err)
		require.Equal(t, expectedID, actual.ID)
		require.Equal(t, expectedSourceProviderType, actual.SourceProvider)
		require.Equal(t, expectedRemote, actual.Remote.String())
	})

	t.Run("Get source entity by remote", func(t *testing.T) {
		fakeGDB, sqlMock, err := GetMockDB()
		require.NoError(t, err)
		repo := NewSourceRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		var (
			expectedID                 uint = 1
			expectedSourceProviderType      = constant.SourceProviderTypeOCI
			expectedRemote                  = mockRemote
		)
		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id", "source_provider", "remote"}).
				AddRow(expectedID, string(expectedSourceProviderType), expectedRemote))
		actual, err := repo.GetByRemote(context.Background(), expectedRemote)
		require.NoError(t, err)
		require.Equal(t, expectedID, actual.ID)
		require.Equal(t, expectedSourceProviderType, actual.SourceProvider)
		require.Equal(t, expectedRemote, actual.Remote.String())
	})

	t.Run("List", func(t *testing.T) {
		fakeGDB, sqlMock, err := GetMockDB()
		require.NoError(t, err)
		repo := NewSourceRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		var (
			expectedIDFirst              uint = 1
			expectedRemoteFirst               = "https://remote/Mocked/Source"
			expectedSourceProviderFirst       = constant.SourceProviderTypeGithub
			expectedIDSecond             uint = 2
			expectedRemoteSecond              = "local://mockedSource"
			expectedSourceProviderSecond      = constant.SourceProviderTypeGithub
		)
		sqlMock.ExpectQuery("SELECT .* FROM `source`").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "remote", "source_provider"}).
					AddRow(expectedIDFirst, expectedRemoteFirst, expectedSourceProviderFirst).
					AddRow(expectedIDSecond, expectedRemoteSecond, expectedSourceProviderSecond))

		actual, err := repo.List(context.Background(), &entity.SourceFilter{})
		require.NoError(t, err)
		require.Len(t, actual, 2)
	})
}
