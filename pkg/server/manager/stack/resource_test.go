package stack

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/domain/entity"
	resourcepersistence "kusionstack.io/kusion/pkg/infra/persistence"
)

func TestStackManager_WriteResources(t *testing.T) {
	ctx := context.Background()
	release := &v1.Release{
		State: &v1.State{
			Resources: []v1.Resource{
				{
					ID:         "apps/v1:Deployment:my-namespace:my-deployment",
					Type:       v1.Kubernetes,
					Attributes: map[string]interface{}{"key1": "value1"},
				},
				{
					ID:         "apps/v1:Deployment:my-namespace:my-deployment",
					Type:       v1.Kubernetes,
					Attributes: map[string]interface{}{"key2": "value2"},
				},
			},
		},
		ModifiedTime: time.Now(),
	}
	stack := &entity.Stack{}
	specID := "spec-1"

	t.Run("WriteResources", func(t *testing.T) {
		fakeGDB, sqlMock, err := resourcepersistence.GetMockDB()
		require.NoError(t, err)
		repo := resourcepersistence.NewResourceRepository(fakeGDB)
		defer resourcepersistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		m := &StackManager{
			resourceRepo: repo,
		}

		sqlMock.ExpectBegin()
		sqlMock.ExpectExec("INSERT").
			WillReturnResult(sqlmock.NewResult(int64(1), int64(2)))
		sqlMock.ExpectCommit()
		err = m.WriteResources(ctx, release, stack, specID)
		require.NoError(t, err)

		var (
			expectedIDFirst    uint = 1
			expectedTypeFirst       = "Kubernetes"
			expectedIDSecond   uint = 2
			expectedTypeSecond      = "Terraform"
		)
		sqlMock.ExpectQuery("SELECT .* FROM `resource`").
			WillReturnRows(
				sqlmock.NewRows([]string{"id", "resource_type"}).
					AddRow(expectedIDFirst, expectedTypeFirst).
					AddRow(expectedIDSecond, expectedTypeSecond))

		actual, err := repo.List(context.Background(), &entity.ResourceFilter{})
		require.NoError(t, err)
		require.Len(t, actual, 2)
	})

	t.Run("MarkResourcesAsDeleted", func(t *testing.T) {
		fakeGDB, sqlMock, err := resourcepersistence.GetMockDB()
		require.NoError(t, err)
		repo := resourcepersistence.NewResourceRepository(fakeGDB)
		defer resourcepersistence.CloseDB(t, fakeGDB)
		defer sqlMock.ExpectClose()

		m := &StackManager{
			resourceRepo: repo,
		}

		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).
				AddRow(1))
		sqlMock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(int64(1), int64(0)))
		sqlMock.ExpectQuery("SELECT").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).
				AddRow(1))
		sqlMock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(int64(1), int64(0)))
		err = m.MarkResourcesAsDeleted(ctx, release)
		require.NoError(t, err)
	})
}
