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
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"type": "user.created"}`))
	w := httptest.NewRecorder()

	userHandler := api.NewUserHandler(userStorageMock)
	userHandler.ClerkWebHookHandler(w, req)

	require.Equal(t, 200, w.Code)
}

func TestUserHandler_ClerkWebHookHandler_invalidOperation(t *testing.T) {
	ctrl := gomock.NewController(t)
	userStorageMock := mock.NewMockUserStorage(ctrl)

	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"type": "invalid.operation"}`))
	w := httptest.NewRecorder()

	userHandler := api.NewUserHandler(userStorageMock)
	userHandler.ClerkWebHookHandler(w, req)

	require.Equal(t, 400, w.Code)
}

func TestUserHandler_ClerkWebHookHandler_errorCreatingUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	userStorageMock := mock.NewMockUserStorage(ctrl)

	userStorageMock.EXPECT().SaveUser(gomock.Any(), gomock.Any()).Return(errors.New("error saving user"))
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"type": "user.created"}`))
	w := httptest.NewRecorder()

	userHandler := api.NewUserHandler(userStorageMock)
	userHandler.ClerkWebHookHandler(w, req)

	require.Equal(t, 500, w.Code)
}

func TestUserHandler_ClerkWebHookHandler_errorDecoding(t *testing.T) {
	ctrl := gomock.NewController(t)
	userStorageMock := mock.NewMockUserStorage(ctrl)

	req := httptest.NewRequest("POST", "/", strings.NewReader(`{`))
	w := httptest.NewRecorder()

	userHandler := api.NewUserHandler(userStorageMock)
	userHandler.ClerkWebHookHandler(w, req)

	require.Equal(t, 400, w.Code)
}

func TestUserHandler_ClerkWebHookHandler_updateUser(t *testing.T) {
	const event = `{
		"data": {
		  "birthday": "",
		  "created_at": 1654012591514,
		  "email_addresses": [
			{
			  "email_address": "example@example.org",
			  "id": "idn_29w83yL7CwVlJXylYLxcslromF1",
			  "linked_to": [],
			  "object": "email_address",
			  "reserved": true,
			  "verification": {
				"attempts": null,
				"expire_at": null,
				"status": "verified",
				"strategy": "admin"
			  }
			}
		  ],
		  "external_accounts": [],
		  "external_id": null,
		  "first_name": "Example",
		  "gender": "",
		  "id": "user_29w83sxmDNGwOuEthce5gg56FcC",
		  "image_url": "https://img.clerk.com/xxxxxx",
		  "last_name": null,
		  "last_sign_in_at": null,
		  "object": "user",
		  "password_enabled": true,
		  "phone_numbers": [],
		  "primary_email_address_id": "idn_29w83yL7CwVlJXylYLxcslromF1",
		  "primary_phone_number_id": null,
		  "primary_web3_wallet_id": null,
		  "private_metadata": {},
		  "profile_image_url": "https://www.gravatar.com/avatar?d=mp",
		  "public_metadata": {},
		  "two_factor_enabled": false,
		  "unsafe_metadata": {},
		  "updated_at": 1654012824306,
		  "username": null,
		  "web3_wallets": []
		},
		"object": "event",
		"type": "user.updated"
	  }`

	ctrl := gomock.NewController(t)
	userStorageMock := mock.NewMockUserStorage(ctrl)
	userStorageMock.EXPECT().SaveUser(gomock.Any(), gomock.Any()).Return(nil)

	req := httptest.NewRequest("POST", "/", strings.NewReader(event))
	w := httptest.NewRecorder()

	userHandler := api.NewUserHandler(userStorageMock)
	userHandler.ClerkWebHookHandler(w, req)

	require.Equal(t, 200, w.Code)
}

func TestUserHandler_ClerkWebHookHandler_deleteUser(t *testing.T) {
	const event = `{
		"data": {
			"deleted": true,
			"id": "user_29wBMCtzATuFJut8jO2VNTVekS4",
			"object": "user"
		},
		"object": "event",
		"type": "user.deleted"
	}`

	ctrl := gomock.NewController(t)
	userStorageMock := mock.NewMockUserStorage(ctrl)
	userStorageMock.EXPECT().DeleteUser(gomock.Any(), gomock.Any()).Return(nil)

	req := httptest.NewRequest("POST", "/", strings.NewReader(event))
	w := httptest.NewRecorder()

	userHandler := api.NewUserHandler(userStorageMock)
	userHandler.ClerkWebHookHandler(w, req)

	require.Equal(t, 200, w.Code)
}
