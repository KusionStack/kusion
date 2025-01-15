//nolint:dupl
package persistence

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"kusionstack.io/kusion/pkg/domain/constant"
	"kusionstack.io/kusion/pkg/domain/entity"
)

func TestResourceRepository(t *testing.T) {
	t.Run("Get", func(t *testing.T) {
		fakeGDB, sqlMock, err := GetMockDB()
		require.NoError(t, err)
		repo := NewResourceRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		var (
			expectedID   uint = 1
			expectedType      = "Kubernetes"
		)
		sqlMock.ExpectQuery("SELECT .* FROM `resource`").
			WillReturnRows(sqlmock.NewRows([]string{"id", "resource_type"}).
				AddRow(expectedID, expectedType))

		actual, err := repo.Get(context.Background(), expectedID)
		require.NoError(t, err)
		require.Equal(t, expectedID, actual.ID)
		require.Equal(t, expectedType, actual.ResourceType)
	})

	t.Run("GetByKusionResourceID", func(t *testing.T) {
		fakeGDB, sqlMock, err := GetMockDB()
		require.NoError(t, err)
		repo := NewResourceRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		var (
			expectedID                uint = 1
			expectedType                   = "Kubernetes"
			expectedKusionResourceID       = "apps/v1:Deployment:my-namespace:my-deployment"
			expectedKusionResourceURN      = "project:stack:workspace:apps/v1:Deployment:my-namespace:my-deployment"
		)
		sqlMock.ExpectQuery("SELECT .* FROM `resource`").
			WillReturnRows(sqlmock.NewRows([]string{"id", "resource_type", "kusion_resource_id", "resource_urn"}).
				AddRow(expectedID, expectedType, expectedKusionResourceID, expectedKusionResourceURN))

		actual, err := repo.GetByKusionResourceURN(context.Background(), expectedKusionResourceURN)
		require.NoError(t, err)
		require.Equal(t, expectedID, actual.ID)
		require.Equal(t, expectedType, actual.ResourceType)
	})

	t.Run("List", func(t *testing.T) {
		fakeGDB, sqlMock, err := GetMockDB()
		require.NoError(t, err)
		repo := NewResourceRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		var (
			expectedIDFirst    uint = 1
			expectedTypeFirst       = "Kubernetes"
			expectedIDSecond   uint = 2
			expectedTypeSecond      = "Terraform"
		)

		sqlMock.ExpectQuery("SELECT count(.*) FROM `resource`").
			WillReturnRows(
				sqlmock.NewRows([]string{"count"}).
					AddRow(2))

		sqlMock.ExpectQuery("SELECT .* FROM `resource` .* IS NULL LIMIT").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "resource_type"}).
					AddRow(expectedIDFirst, expectedTypeFirst).
					AddRow(expectedIDSecond, expectedTypeSecond))

		actual, err := repo.List(context.Background(), &entity.ResourceFilter{
			Pagination: &entity.Pagination{
				Page:     constant.CommonPageDefault,
				PageSize: constant.CommonPageSizeDefault,
			},
		})
		require.NoError(t, err)
		require.Len(t, actual.Resources, 2)
	})

	t.Run("Delete existing record", func(t *testing.T) {
		fakeGDB, sqlMock, err := GetMockDB()
		require.NoError(t, err)
		repo := NewResourceRepository(fakeGDB)
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

	t.Run("Batch delete existing record", func(t *testing.T) {
		fakeGDB, sqlMock, err := GetMockDB()
		require.NoError(t, err)
		repo := NewResourceRepository(fakeGDB)
		defer CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		expectedResources := []*entity.Resource{
			{
				ID:           1,
				ResourceType: "Kubernetes",
				Stack:        &entity.Stack{},
			},
			{
				ID:           2,
				ResourceType: "Terraform",
				Stack:        &entity.Stack{},
			},
		}
		sqlMock.ExpectBegin()
		sqlMock.ExpectExec("DELETE").WillReturnResult(sqlmock.NewResult(int64(1), int64(1)))
		sqlMock.ExpectCommit()
		err = repo.BatchDelete(context.Background(), expectedResources)
		require.NoError(t, err)
	})

	t.Run("Update existing record", func(t *testing.T) {
		fakeGDB, sqlMock, err := GetMockDB()
		require.NoError(t, err)
		repo := NewResourceRepository(fakeGDB)

		var (
			expectedID, expectedRows uint = 1, 1
			actual                        = entity.Resource{
				ID:           1,
				ResourceType: "Kubernetes",
			}
		)
		sqlMock.ExpectExec("UPDATE").
			WillReturnResult(sqlmock.NewResult(int64(expectedID), int64(expectedRows)))
		err = repo.Update(context.Background(), &actual)
		require.NoError(t, err)
	})
}
