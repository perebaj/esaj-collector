package api_test

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/perebaj/esaj/api"
	"github.com/perebaj/esaj/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUserHandler_ClerkWebHookHandler_createUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	userStorageMock := mock.NewMockUserStorage(ctrl)

	userStorageMock.EXPECT().SaveUser(gomock.Any(), gomock.Any()).Return(nil)
	req := httptest.NewRequest("POST", "/random", strings.NewReader(`{"type": "user.created"}`))
	w := httptest.NewRecorder()

	userHandler := api.NewUserHandler(userStorageMock)
	userHandler.ClerkWebHookHandler(w, req)

	require.Equal(t, 200, w.Code)
}

func TestUserHandler_ClerkWebHookHandler_invalidOperation(t *testing.T) {
	ctrl := gomock.NewController(t)
	userStorageMock := mock.NewMockUserStorage(ctrl)

	req := httptest.NewRequest("POST", "/random", strings.NewReader(`{"type": "invalid.operation"}`))
	w := httptest.NewRecorder()

	userHandler := api.NewUserHandler(userStorageMock)
	userHandler.ClerkWebHookHandler(w, req)

	require.Equal(t, 400, w.Code)
}

func TestUserHandler_ClerkWebHookHandler_errorCreatingUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	userStorageMock := mock.NewMockUserStorage(ctrl)

	userStorageMock.EXPECT().SaveUser(gomock.Any(), gomock.Any()).Return(errors.New("error saving user"))
	req := httptest.NewRequest("POST", "/random", strings.NewReader(`{"type": "user.created"}`))
	w := httptest.NewRecorder()

	userHandler := api.NewUserHandler(userStorageMock)
	userHandler.ClerkWebHookHandler(w, req)

	require.Equal(t, 500, w.Code)
}

func TestUserHandler_ClerkWebHookHandler_errorDecoding(t *testing.T) {
	ctrl := gomock.NewController(t)
	userStorageMock := mock.NewMockUserStorage(ctrl)

	req := httptest.NewRequest("POST", "/random", strings.NewReader(`{`))
	w := httptest.NewRecorder()

	userHandler := api.NewUserHandler(userStorageMock)
	userHandler.ClerkWebHookHandler(w, req)

	require.Equal(t, 400, w.Code)
}
