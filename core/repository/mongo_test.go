package repository_test

import (
	"context"
	"testing"

	"git.neds.sh/technology/pricekinetics/tools/codetest/core/repository"
	"git.neds.sh/technology/pricekinetics/tools/codetest/internal/modeltest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func Test_mongoRepo_UpdateEvent(t *testing.T) {
	repo, err := repository.NewMongoRepository(context.Background(), "mongodb://localhost:27017", "codetest")
	require.NoError(t, err)
	defer repo.Close(context.Background())

	in := modeltest.PopulatedEvent()
	defer repo.DeleteEventByID(context.Background(), in.ID)
	require.NoError(t, repo.UpdateEvent(context.Background(), in))

	out, err := repo.GetEventByID(context.Background(), in.ID)
	require.NoError(t, err)
	assert.True(t, proto.Equal(in, out), "round trip lost a field:\n want %v\n got  %v", in, out)
}

func Test_mongoRepo_GetEventByID_NotFound(t *testing.T) {
	repo, err := repository.NewMongoRepository(context.Background(), "mongodb://localhost:27017", "codetest")
	require.NoError(t, err)
	defer repo.Close(context.Background())

	out, err := repo.GetEventByID(context.Background(), "does-not-exist")
	assert.NoError(t, err)
	assert.Nil(t, out)
}
