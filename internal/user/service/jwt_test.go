package user_service

import (
	"Makhkets/internal/configs"
	mock "Makhkets/internal/user/mocks"
	user "Makhkets/internal/user/repository"
	"Makhkets/pkg/logging"
	"fmt"
	mock2 "github.com/stretchr/testify/mock"
	"net/http"
	"testing"
)

func getMockService() *service {
	logger := logging.GetLogger()
	config := configs.GetConfig()
	repositroy := mock.NewMockRepository()

	return &service{
		repository: repositroy,
		config:     config,
		logger:     &logger,
	}
}

func TestGenerateAccessToken(t *testing.T) {
	l := logging.GetLogger()
	cfg := configs.GetConfig()
	srvc := service{
		config: cfg,
		logger: &l,
	}

	var cases = []struct {
		name    string
		isError bool
		user    user.UserDTO
	}{
		{
			name:    "Zero ID",
			isError: true,
			user: user.UserDTO{
				Id:       "0",
				Username: "aliev",
				Password: "123321asSs",
				Status:   user.StatusUser,
				IsBanned: false,
			},
		},
		{
			name:    "Invalid type ID",
			isError: true,
			user: user.UserDTO{
				Id:       "x",
				Username: "aliev",
				Password: "123321asSs",
				Status:   user.StatusUser,
				IsBanned: false,
			},
		},
		{
			name:    "Invalid status",
			isError: true,
			user: user.UserDTO{
				Id:       "1",
				Username: "aliev",
				Password: "123321asSs",
				Status:   "InvalidStatus",
				IsBanned: true,
			},
		},
		{
			name:    "working version",
			isError: false,
			user: user.UserDTO{
				Id:       "1",
				Username: "aliev",
				Password: "123321asSs",
				Status:   user.StatusAdmin,
				IsBanned: true,
			},
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			token, err := srvc.GenerateAccessToken(&tCase.user)
			if !(err != nil == tCase.isError) {
				t.Errorf("error found in isError")
				return
			}

			if !tCase.isError {
				if len([]rune(token)) < 12 {
					t.Errorf("Unvalid token: %s", token)
					return
				}
			}
		})
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	l := logging.GetLogger()
	cfg := configs.GetConfig()
	srvc := service{
		config: cfg,
		logger: &l,
	}

	var cases = []struct {
		name    string
		isError bool
		user    user.UserDTO
	}{
		{
			name:    "Zero ID",
			isError: true,
			user: user.UserDTO{
				Id:       "0",
				Username: "aliev",
				Password: "123321asSs",
				Status:   user.StatusUser,
				IsBanned: false,
			},
		},
		{
			name:    "Invalid type ID",
			isError: true,
			user: user.UserDTO{
				Id:       "x",
				Username: "aliev",
				Password: "123321asSs",
				Status:   user.StatusUser,
				IsBanned: false,
			},
		},
		{
			name:    "Working version",
			isError: false,
			user: user.UserDTO{
				Id:       "1",
				Username: "aliev",
				Password: "123321asSs",
				Status:   user.StatusUser,
				IsBanned: false,
			},
		},
	}

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			token, id, err := srvc.GenerateRefreshToken(&tCase.user)
			if !(err != nil == tCase.isError) {
				t.Errorf("error found in isError")
				return
			}

			if !tCase.isError {
				if len([]rune(token)) < 12 {
					t.Errorf("Unvalid token: %s", token)
					return
				}
			}

			if id < 0 {
				t.Error("Zero id")
				return
			}
		})
	}
}

func TestCreateTokenPair(t *testing.T) {
	cases := []struct {
		name    string
		isError bool
		srvc    *service
		user    user.UserDTO
		saveErr error
	}{
		{
			name:    "Zero ID",
			isError: true,
			srvc:    getMockService(),
			user: user.UserDTO{
				Id:       "0",
				Username: "aliev",
				Password: "123321asSs",
				Status:   user.StatusUser,
				IsBanned: false,
			},
		},
		{
			name:    "Invalid type ID",
			isError: true,
			srvc:    getMockService(),
			user: user.UserDTO{
				Id:       "x",
				Username: "aliev",
				Password: "123321asSs",
				Status:   user.StatusUser,
				IsBanned: false,
			},
		},
		{
			name:    "Working version",
			isError: false,
			srvc:    getMockService(),
			user: user.UserDTO{
				Id:       "1",
				Username: "aliev",
				Password: "123321asSs",
				Status:   user.StatusUser,
				IsBanned: false,
			},
			saveErr: nil,
		},
		{
			name:    "Error saving session",
			isError: true,
			srvc:    getMockService(),
			user: user.UserDTO{
				Id:       "1",
				Username: "aliev",
				Password: "123321asSs",
				Status:   user.StatusUser,
				IsBanned: false,
			},
			saveErr: fmt.Errorf("error"),
		},
	}

	req := http.Request{RemoteAddr: "127.0.0.1"}
	req.Header = make(http.Header)
	req.Header.Set(
		"User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "+
			"AppleWebKit/537.36 (KHTML, like Gecko) "+
			"Chrome/114.0.0.0 Safari/537.36",
	)

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {
			tCase.srvc.repository.(*mock.MockRepository).
				On("SaveRefreshSession", mock2.Anything, mock2.Anything).Return(tCase.saveErr)

			_, _, err := tCase.srvc.CreateTokenPair(&tCase.user, &req)
			if !(err != nil == tCase.isError) {
				t.Errorf("unexpected error status for test case '%s': got %v, expected %v", tCase.name, err, tCase.isError)
				return
			}

			if !tCase.isError {
				tCase.srvc.repository.(*mock.MockRepository).AssertCalled(t, "SaveRefreshSession", mock2.Anything, mock2.Anything)
			}
		})
	}
}

func TestParseToken(t *testing.T) {
	// WARNING: если тесты не будут проходить, то возможно сменили secret_key в config's
	cases := []struct {
		name     string
		token    string
		isAccess bool
		isError  bool
	}{
		{
			name: "working version",
		},
	}

	//

	for _, tCase := range cases {
		t.Run(tCase.name, func(t *testing.T) {

		})
	}
}
