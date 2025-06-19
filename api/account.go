package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/starjardin/simplebank/db/sqlc"
	"github.com/starjardin/simplebank/token"
)

type createAccountRequest struct {
	Currency string `json:"currency" binding:"required,currency"`
}

func (server *Server) createAccount(c *gin.Context) {
	var req createAccountRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))

		return
	}

	authPayload := c.MustGet(authorizationPaylaodKey).(*token.Payload)

	arg := db.CreateAccountParams{
		Owner:    authPayload.Username,
		Currency: req.Currency,
		Balance:  0,
	}

	account, err := server.store.CreateAccount(c, arg)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, account)
}

type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getAccount(c *gin.Context) {
	var req getAccountRequest

	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))

		return
	}

	account, err := server.store.GetAccount(c, req.ID)

	if err != nil {

		if errors.Is(err, db.ErrorRecordNotFound) {
			c.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	authPayload := c.MustGet(authorizationPaylaodKey).(*token.Payload)

	if authPayload.Username != account.Owner {
		err := errors.New("account does not belong to the uathenticated user")

		c.JSON(http.StatusUnauthorized, errorResponse(err))
		return

	}

	c.JSON(http.StatusOK, account)
}

type listAccountRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listAccount(c *gin.Context) {
	var req listAccountRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))

		return
	}

	authPayload := c.MustGet(authorizationPaylaodKey).(*token.Payload)

	arg := db.ListAccountsParams{
		Owner:  authPayload.Username,
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	account, err := server.store.ListAccounts(c, arg)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, account)
}
