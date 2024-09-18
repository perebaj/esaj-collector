package api

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/perebaj/esaj/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestHandler_ProcessesByOABHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	storageMock := mock.NewMockStorage(ctrl)

	storageMock.EXPECT().ProcessBasicInfoByOAB(gomock.Any(), "123").Return(nil, nil)
	req := httptest.NewRequest("POST", "/?oab=123", strings.NewReader(`{"oab": "123"}`))
	w := httptest.NewRecorder()

	h := NewHandler(storageMock, nil)
	h.ProcessesByOABHandler(w, req)

	require.Equal(t, 200, w.Code)
}
