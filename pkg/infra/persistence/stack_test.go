package persistence

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
)

func TestStackRepository(t *testing.T) {
	mockRemote := "https://github.com/mockorg/mockrepo"
	mockRemoteURL, err := url.Parse(mockRemote)
	require.NoError(t, err)

	t.Run("Create", func(t *testing.T) {
		fakeGDB, sqlMock, err := GetMockDB()
		require.NoError(t, err)
		repo := NewStackRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		var (
			expectedID, expectedRows uint = 1, 1
			actual                        = entity.Stack{
				Name: "mockedStack",
				Project: &entity.Project{
					ID:   1,
					Name: "mockedProject",
					Path: "/path/to/project",
					Source: &entity.Source{
						ID:             1,
						SourceProvider: constant.SourceProviderTypeGithub,
						Remote:         mockRemoteURL,
					},
					Organization: &entity.Organization{
						ID: 1,
					},
				},
				Path:                 "/path/to/stack",
				DesiredVersion:       "master",
				Labels:               []string{"testLabel"},
				Owners:               []string{"hua.li", "xiaoming.li"},
				SyncState:            constant.StackStateUnSynced,
				LastAppliedTimestamp: time.Now(),
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
		repo := NewStackRepository(fakeGDB)
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
		repo := NewStackRepository(fakeGDB)
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
		repo := NewStackRepository(fakeGDB)

		var (
			expectedID, expectedRows uint = 1, 1
			actual                        = entity.Stack{
				ID:        1,
				SyncState: constant.StackStateUnSynced,
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
		repo := NewStackRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		actual := entity.Stack{
			SyncState: constant.StackStateUnSynced,
		}
		err = repo.Update(context.Background(), &actual)
		require.ErrorIs(t, err, gorm.ErrMissingWhereClause)
	})

	t.Run("Get", func(t *testing.T) {
		fakeGDB, sqlMock, err := GetMockDB()
		require.NoError(t, err)
		repo := NewStackRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		var (
			expectedID    uint = 1
			expectedPath       = "/path/to/stack"
			expectedState      = constant.StackStateUnSynced
		)
		sqlMock.ExpectQuery("SELECT .* FROM `stack`").
			WillReturnRows(sqlmock.NewRows([]string{"id", "path", "sync_state", "Project__id", "Project__name", "Project__path"}).
				AddRow(expectedID, expectedPath, string(expectedState), 1, "mockedProject", "/path/to/project"))

		actual, err := repo.Get(context.Background(), expectedID)
		require.NoError(t, err)
		require.Equal(t, expectedID, actual.ID)
		require.Equal(t, expectedState, actual.SyncState)
	})

	t.Run("List", func(t *testing.T) {
		fakeGDB, sqlMock, err := GetMockDB()
		require.NoError(t, err)
		repo := NewStackRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		var (
			expectedIDFirst         uint = 1
			expectedNameFirst            = "mockedStack"
			expectedPathFirst            = "/path/to/stack"
			expectedSyncStateFirst       = constant.StackStateUnSynced
			expectedIDSecond        uint = 2
			expectedNameSecond           = "mockedStack2"
			expectedPathSecond           = "/path/to/stack/2"
			expectedSyncStateSecond      = constant.StackStateSynced
		)
		sqlMock.ExpectQuery("SELECT .* FROM `stack`").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "name", "path", "sync_state", "Project__id", "Project__name", "Project__path"}).
					AddRow(expectedIDFirst, expectedNameFirst, expectedPathFirst, expectedSyncStateFirst, 1, "mockedProject", "path/to/project").
					AddRow(expectedIDSecond, expectedNameSecond, expectedPathSecond, expectedSyncStateSecond, 2, "mockedProject2", "path/to/project2"))

		actual, err := repo.List(context.Background(), &entity.StackFilter{})
		require.NoError(t, err)
		require.Len(t, actual, 2)
	})

	// t.Run("Get stack entity by source id and path", func(t *testing.T) {
	// 	fakeGDB, sqlMock, err := GetMockDB()
	// 	require.NoError(t, err)
	// 	repo := NewStackRepository(fakeGDB)
	// 	defer CloseDB(t, fakeGDB)
	// 	defer sqlMock.ExpectClose()

	// 	var (
	// 		expectedID        uint = 1
	// 		expectedState          = constant.StackStateUnSynced
	// 	)
	// 	sqlMock.ExpectQuery("SELECT.*FROM "stack"").
	// 		WillReturnRows(sqlmock.NewRows([]string{"id", "source_id", "path", "sync_state", "Source__source_provider"}).
	// 			AddRow(expectedID, 2, "/path/to/ws", string(expectedState), string(constant.SourceProviderTypeGithub)))
	// 	actual, err := repo.GetBy(context.Background(), 2, "/path/to/ws")
	// 	require.NoError(t, err)
	// 	require.Equal(t, expectedID, actual.ID)
	// 	require.Equal(t, expectedState, actual.State)
	// })

	// t.Run("Find", func(t *testing.T) {
	// 	fakeGDB, sqlMock, err := GetMockDB()
	// 	require.NoError(t, err)
	// 	repo := NewStackRepository(fakeGDB)
	// 	defer CloseDB(t, fakeGDB)
	// 	defer sqlMock.ExpectClose()

	// 	sqlMock.ExpectQuery("SELECT").
	// 		WillReturnRows(sqlmock.NewRows([]string{"id", "state", "framework", "Source__source_provider"}).
	// 			AddRow(1, string(constant.StackStateUnSynced), string(constant.FrameworkTypeKusion), string(constant.SourceProviderTypeRepoServer)).
	// 			AddRow(2, string(constant.StackStateUnSynced), string(constant.FrameworkTypeTerraform), string(constant.SourceProviderTypeRepoServer)))
	// 	actuals, err := repo.Find(context.Background(), repository.StackQuery{
	// 		Bound: repository.Bound{
	// 			Offset: 1,
	// 			Limit:  10,
	// 		},
	// 	})
	// 	require.NoError(t, err)
	// 	require.Equal(t, 2, len(actuals))
	// })
}
