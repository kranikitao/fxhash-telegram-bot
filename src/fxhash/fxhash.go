package fxhash

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kranikitao/fxhash-telegram-bot/src/errors"
)

type FxHash struct {
}

const (
	ErrTypeUserNotFound = "fx_hash_user_not_found"
)

func New() *FxHash {
	return &FxHash{}
}

type GenerativeTokensResponse struct {
	Data *GenerativeTokensDataResponse `json:"data"`
}

type GenerativeTokensDataResponse struct {
	GenerativeTokens []*GenerativeToken `json:"generativeTokens"`
}
type GenerativeToken struct {
	Name                string               `json:"name"`
	Enabled             bool                 `json:"enabled"`
	Id                  int64                `json:"id"`
	Slug                string               `json:"slug"`
	Balance             int                  `json:"balance"`
	Supply              int                  `json:"supply"`
	ObjktsCount         int                  `json:"objktsCount"`
	Flag                string               `json:"flag"`
	Reserves            []*Reserve           `json:"reserves"`
	Author              *Author              `json:"author"`
	Collaborators       []*Author            `json:"collaborators"`
	MintOpensAt         time.Time            `json:"mintOpensAt"`
	PricingFixed        *PricingFixed        `json:"pricingFixed"`
	PricingDutchAuction *PricingDutchAuction `json:"pricingDutchAuction"`
}

type PricingFixed struct {
	Price int `json:"price"`
}
type PricingDutchAuction struct {
	RestingPrice int `json:"restingPrice"`
}

type Reserve struct {
	Name   string `json:"name"`
	Amount int    `json:"amount"`
}
type Author struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

func (fxHash *FxHash) GetLastGeneratives() ([]*GenerativeToken, *errors.Error) {
	bodyString := `{"query":"query Query($filters: GenerativeTokenFilter, $sort: GenerativeSortInput, $take: Int) {\n  generativeTokens(filters: $filters, sort: $sort, take: $take) {\n    author {\n      name\n      id\n      collaborators {\n        name\n        id\n      }\n      type\n    }\n    name\n    slug\n    createdAt\n    id\n    flag\n    balance\n    objktsCount\n    supply\n    mintOpensAt\n    reserves {\n      amount\n    }\n    enabled\n  }\n}","variables":{"sort":{"mintOpensAt":"DESC"},"take":50}}`

	response, err := fxHash.request(bodyString)
	if err != nil {
		return nil, err
	}

	var result []*GenerativeToken
	for _, token := range response.Data.GenerativeTokens {
		if !fxHash.isAvailableToMint(token) {
			continue
		}
		result = append(result, token)
	}
	if len(result) == 0 {
		return nil, errors.New("not found generatives", "")
	}
	return result, nil
}

func (*FxHash) request(bodyString string) (*GenerativeTokensResponse, *errors.Error) {
	postBody := []byte(bodyString)
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post("https://api.fxhash.xyz/graphql", "application/json", responseBody)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	response := &GenerativeTokensResponse{}
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	if response == nil || response.Data == nil {
		return nil, errors.New("empty result", "")
	}

	return response, nil
}

func (*FxHash) isAvailableToMint(token *GenerativeToken) bool {
	if token.Flag == "HIDDEN" {
		return false
	}
	if !token.Enabled {
		return false
	}
	if token.Balance == 0 {
		return false
	}

	if time.Until(token.MintOpensAt) > 0 {
		return false
	}
	reserved := 0
	if token.Reserves != nil {
		for _, reserve := range token.Reserves {
			reserved += reserve.Amount
		}
	}
	if reserved >= token.Balance {
		return false
	}

	return true
}

func (fxHash *FxHash) GetFreeGeneratives() ([]*GenerativeToken, *errors.Error) {
	bodyString := `{"query":"query Query($filters: GenerativeTokenFilter, $sort: GenerativeSortInput, $take: Int) {\n  generativeTokens(filters: $filters, sort: $sort, take: $take) {\n    author {\n      name\n      id\n      collaborators {\n        name\n        id\n      }\n      type\n    }\n    name\n    slug\n    createdAt\n    id\n    flag\n    balance\n    objktsCount\n    supply\n    mintOpensAt\n    reserves {\n      amount\n    }\n    enabled\n    pricingFixed {\n      price\n    }\n    pricingDutchAuction {\n      finalPrice\n      restingPrice\n      levels\n      decrementDuration\n      opensAt\n    }\n  }\n}","variables":{"sort":{"mintOpensAt":"DESC"},"take":50,"filters":{"price_lte":1}}}`

	response, err := fxHash.request(bodyString)
	if err != nil {
		return nil, err
	}

	var result []*GenerativeToken
	for _, token := range response.Data.GenerativeTokens {
		if !fxHash.isAvailableToMint(token) {
			continue
		}

		hasZeroCost := token.PricingFixed != nil && token.PricingFixed.Price == 0
		// (token.PricingDutchAuction != nil && token.PricingDutchAuction.RestingPrice == 0)
		if !hasZeroCost {
			continue
		}

		result = append(result, token)
	}
	if len(result) == 0 {
		return nil, errors.New("not found generatives", "")
	}
	return result, nil
}

type UserResponse struct {
	Data *UserDataResponse `json:"data"`
}

type UserDataResponse struct {
	User *User `json:"user"`
}
type User struct {
	Name string `json:"name"`
	Id   string `json:"id"`
	Flag string `json:"flag"`
}

func (fxHash *FxHash) GetFxHashUser(fxHashUserName string) (*User, *errors.Error) {
	bodyString := fmt.Sprintf(`{"query":"query User($name: String) {\n  user(name: $name) {\n    name\n    id\n    flag\n  }\n}","variables":{"name":"%s"}}`, fxHashUserName)
	postBody := []byte(bodyString)
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post("https://api.fxhash.xyz/graphql", "application/json", responseBody)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	response := &UserResponse{}
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	if response.Data.User == nil {
		return nil, errors.New("User not found", ErrTypeUserNotFound)
	}
	return response.Data.User, nil
}
