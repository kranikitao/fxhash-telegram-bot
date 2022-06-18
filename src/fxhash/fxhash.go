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
	Name          string     `json:"name"`
	Enabled       bool       `json:"enabled"`
	Id            int64      `json:"id"`
	Slug          string     `json:"slug"`
	Balance       int        `json:"balance"`
	Flag          string     `json:"flag"`
	Reserves      []*Reserve `json:"reserves"`
	Author        *Author    `json:"author"`
	Collaborators []*Author  `json:"collaborators"`
	MintOpensAt   time.Time  `json:"mintOpensAt"`
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
	bodyString := `{"query":"query Query($filters: GenerativeTokenFilter, $sort: GenerativeSortInput, $take: Int) {\n  generativeTokens(filters: $filters, sort: $sort, take: $take) {\n    author {\n      name\n      id\n      collaborators {\n        name\n        id\n      }\n      type\n    }\n    name\n    slug\n    createdAt\n    id\n    flag\n    balance\n    mintOpensAt\n    reserves {\n      amount\n    }\n    enabled\n  }\n}","variables":{"sort":{"mintOpensAt":"DESC"},"take":50}}`
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
	var result []*GenerativeToken
	if response == nil || response.Data == nil {
		return nil, errors.New("empty result", "")
	}
	for _, token := range response.Data.GenerativeTokens {
		if token.Flag == "HIDDEN" {
			continue
		}
		if !token.Enabled {
			continue
		}
		if token.Balance == 0 {
			continue
		}

		if time.Until(token.MintOpensAt) > 0 {
			continue
		}
		reserved := 0
		if token.Reserves != nil {
			for _, reserve := range token.Reserves {
				reserved += reserve.Amount
			}
		}
		if token.Balance == reserved {
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
