package model

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type NotiEvent struct {
	Dapp      common.Address `json:"dapp"`
	User      common.Address `json:"user"`
	Title     string         `json:"title"`
	Body      string         `json:"body"`
	AtTime    *big.Int       `json:"atTime"`
	SystemApp bool           `json:"systemApp"`
}

type Notification struct {
	Dapp        string `json:"dapp"`
	User        string `json:"user"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	AtTime      uint64 `json:"atTime"`
	DeviceToken string `json:"deviceToken"`
}
