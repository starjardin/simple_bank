package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/starjardin/simplebank/db/sqlc"
	"github.com/starjardin/simplebank/token"
)

type transferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

func (server *Server) createTransfer(c *gin.Context) {
	var req transferRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))

		return
	}

	formAccount, valid := server.validAccount(c, req.FromAccountID, req.Currency)

	if !valid {
		return
	}

	authPayload := c.MustGet(authorizationPaylaodKey).(*token.Payload)

	if formAccount.Owner != authPayload.Username {
		err := errors.New("from account doesn't belong to the authenticated user")
		c.JSON(http.StatusUnauthorized, errorResponse(err))

		return
	}

	_, valid = server.validAccount(c, req.FromAccountID, req.Currency)

	if !valid {
		return
	}

	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	result, err := server.store.TransferTx(c, arg)

	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, result)
}

func (server *Server) validAccount(ctx *gin.Context, accountID int64, currency string) (db.Account, bool) {
	account, err := server.store.GetAccount(ctx, accountID)

	if err != nil {
		if errors.Is(err, db.ErrorRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return account, false
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return account, false
	}
	if account.Currency != currency {
		err := fmt.Errorf("account %d currency mismatch: %s vs %s", account.ID, account.Currency, currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return account, false
	}
	return account, true
}
