package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	mockdb "github.com/starjardin/simplebank/db/mock"
	db "github.com/starjardin/simplebank/db/sqlc"
	"github.com/starjardin/simplebank/utils"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type eqCreateUserParamsMatcher struct {
	arg      db.CreateUserParams
	password string
}

func (e eqCreateUserParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	err := utils.CheckPassword(e.password, arg.HashedPassword)
	if err != nil {
		return false
	}

	e.arg.HashedPassword = arg.HashedPassword
	return reflect.DeepEqual(arg, arg)
}

func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("Matches arg %v and password %v", e.arg, e.password)
}

func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{arg, password}
}

func TestCreateUserAPI(t *testing.T) {
	user, password := randomUser(t)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"user_name": user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.CreateUserParams{
					Username: user.Username,
					FullName: user.FullName,
					Email:    user.Email,
				}

				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, user)
			},
		},
		// {
		// 	name:      "NotFound",
		// 	accountID: account.ID,
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		store.EXPECT().
		// 			GetAccount(gomock.Any(), gomock.Eq(account.ID)).
		// 			Times(1).
		// 			Return(db.Account{}, sql.ErrNoRows)
		// 	},
		// 	checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
		// 		require.Equal(t, http.StatusNotFound, recorder.Code)
		// 	},
		// },
		// {
		// 	name:      "InternalError",
		// 	accountID: account.ID,
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		store.EXPECT().
		// 			GetAccount(gomock.Any(), gomock.Eq(account.ID)).
		// 			Times(1).
		// 			Return(db.Account{}, sql.ErrConnDone)
		// 	},
		// 	checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
		// 		require.Equal(t, http.StatusInternalServerError, recorder.Code)
		// 	},
		// },
		// {
		// 	name:      "InvalidID",
		// 	accountID: 0,
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		store.EXPECT().
		// 			GetAccount(gomock.Any(), gomock.Any).
		// 			Times(0)
		// 	},
		// 	checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
		// 		require.Equal(t, http.StatusBadRequest, recorder.Code)
		// 	},
		// },
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctr := gomock.NewController(t)

			defer ctr.Finish()

			store := mockdb.NewMockStore(ctr)
			tc.buildStubs(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)

			require.NoError(t, err)

			url := "/users"

			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))

			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(recorder)
		})
	}
}

func randomUser(t *testing.T) (user db.User, password string) {
	password = utils.RandomString(6)
	hashedPassword, err := utils.HashedPassword(password)
	require.NoError(t, err)
	return db.User{

		Username:       utils.RandomOwner(),
		Email:          utils.RandomEmail(),
		HashedPassword: hashedPassword,
		FullName:       utils.RandomOwner(),
	}, password
}

func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	var gotUser struct {
		Username string `json:"username"`
		FullName string `json:"full_name"`
		Email    string `json:"email"`
	}

	err = json.Unmarshal(data, &gotUser)
	require.NoError(t, err)

	require.Equal(t, user.Username, gotUser.Username)
	require.Equal(t, user.FullName, gotUser.FullName)
	require.Equal(t, user.Email, gotUser.Email)
}
