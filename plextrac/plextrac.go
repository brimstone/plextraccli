// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package plextrac

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"plextraccli/utils"
	"time"

	"github.com/pquerna/otp/totp"
)

type UserAgent struct {
	username  string
	password  string
	mfaseed   string
	tenantID  int
	tenantURL string
	authToken string
}

func New(u, p, token, seed string) (*UserAgent, error) {
	c := UserAgent{
		tenantURL: "idealintegrations",
		username:  u,
		password:  p,
		mfaseed:   seed,
		authToken: token,
	}

	err := c.Login(u, p, token, seed)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (ua *UserAgent) GetTenantID() int {
	return ua.tenantID
}

func (ua *UserAgent) Login(u, p, token, seed string) error {
	// fmt.Printf("Trying to login with %s\n", u) // TODO debug
	payload := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: u,
		Password: p,
	}

	type AuthenticationResponse struct {
		Status   string `json:"status"`
		TenantID int    `json:"tenant_id"`
		Cookie   string `json:"cookie"`
		Token    string `json:"token"`
		// If MFA, then these will be set
		MfaEnabled bool   `json:"mfa_enabled"`
		Code       string `json:"code"`
		// Errors
		StatusCode int    `json:"statusCode"`
		Error      string `json:"error"`
		Message    string `json:"message"`
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, "https://"+ua.tenantURL+".plextrac.com/api/v1/authenticate", bytes.NewReader(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer utils.Must(resp.Body.Close)

	var authResponse AuthenticationResponse

	err = json.NewDecoder(resp.Body).Decode(&authResponse)
	if err != nil {
		return err
	}

	if authResponse.Status == "error" {
		return errors.New(authResponse.Message)
	}

	ua.tenantID = authResponse.TenantID

	if authResponse.MfaEnabled {
		if token == "" {
			token, err = totp.GenerateCode(seed, time.Now())
			if err != nil {
				return err
			}
		}

		data := struct {
			Code  string `json:"code"`
			Token string `json:"token"`
		}{
			Code:  authResponse.Code,
			Token: token,
		}

		payloadBytes, err := json.Marshal(data)
		if err != nil {
			return err
		}

		req, err := http.NewRequest(http.MethodPost, "https://"+ua.tenantURL+".plextrac.com/api/v1/authenticate/mfa", bytes.NewReader(payloadBytes))
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer utils.Must(resp.Body.Close)

		err = json.NewDecoder(resp.Body).Decode(&authResponse)
		if err != nil {
			return err
		}

		if authResponse.Status == "error" {
			return errors.New(authResponse.Message)
		}
	}

	ua.authToken = authResponse.Token

	return nil
}

func (ua *UserAgent) apiGet(path string, response interface{}) (string, error) {
	fullpath := "https://" + ua.tenantURL + ".plextrac.com/api/" + path
	req, err := http.NewRequest(http.MethodGet, fullpath, nil)
	req.Header.Set("Authorization", "Bearer "+ua.authToken)
	// fmt.Printf("-H \"Authorization: Bearer %s\" %s\n", ua.authToken, fullpath)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer utils.Must(resp.Body.Close)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if response != nil {
		err = json.Unmarshal(body, response)
	}

	/*
		if err != nil {
			return string(body), err
		}

		// OR
		//err = json.NewDecoder(resp.Body).Decode(response)
		//if err != nil {
		//	return err
		//}

		return string(body), nil
	*/
	return string(body), err
}
