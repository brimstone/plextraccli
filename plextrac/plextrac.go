// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package plextrac

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pquerna/otp/totp"
)

type OnRenewFunc func(string, time.Time) error

type UserAgent struct {
	authToken      string
	expires        time.Time
	onRenewFunc    OnRenewFunc
	tenantID       int
	tenantURL      string
	authTokenMutex sync.Mutex
	tags           []tenantTag
}

type tenantTag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type jwtPayload struct {
	Username string `json:"username"`
	TenantID int    `json:"tenantId"`
	Iat      int    `json:"iat"`
	Exp      int64  `json:"exp"`
}

type NewOptions struct {
	InstanceURL string
	Username    string
	Password    string
	MFAToken    string
	MFASeed     string
	AuthToken   string
	OnRenewFunc OnRenewFunc
}

func New(o NewOptions) (*UserAgent, []error, error) {
	var err error

	var warnings []error

	if o.InstanceURL == "" {
		return nil, warnings, errors.New("must have instanceurl")
	}

	if (o.Username == "" || o.Password == "") && o.AuthToken == "" {
		return nil, warnings, errors.New("must have username/password or authtoken")
	}

	ua := UserAgent{
		tenantURL:   o.InstanceURL,
		authToken:   o.AuthToken,
		onRenewFunc: o.OnRenewFunc,
	}

	if o.AuthToken != "" {
		ua.expires, err = getExpirationFromToken(ua.authToken)
		if err == nil {
			ua.authToken = o.AuthToken
		}
	}

	if time.Now().After(ua.expires) {
		warnings, err = ua.Login(o.Username, o.Password, o.MFAToken, o.MFASeed)
		if err != nil {
			return nil, warnings, err
		}
	}

	return &ua, warnings, nil
}

func (ua *UserAgent) GetTenantID() int {
	return ua.tenantID
}

func (ua *UserAgent) Login(u, p, token, seed string) ([]error, error) {
	var warnings []error

	authPayload := struct {
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

	authPayloadBytes, err := json.Marshal(authPayload)
	if err != nil {
		return warnings, err
	}

	req, err := http.NewRequest(http.MethodPost, "https://"+ua.tenantURL+"/api/v1/authenticate", bytes.NewReader(authPayloadBytes))
	if err != nil {
		return warnings, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return warnings, err
	}
	defer must(resp.Body.Close)

	var authResponse AuthenticationResponse

	err = json.NewDecoder(resp.Body).Decode(&authResponse)
	if err != nil {
		return warnings, err
	}

	if authResponse.Status == "error" {
		return warnings, errors.New(authResponse.Message)
	}

	ua.tenantID = authResponse.TenantID

	if authResponse.MfaEnabled {
		if token == "" {
			token, err = totp.GenerateCode(seed, time.Now())
			if err != nil {
				return warnings, err
			}
		}

		data := struct {
			Code  string `json:"code"`
			Token string `json:"token"`
		}{
			Code:  authResponse.Code,
			Token: token,
		}

		authPayloadBytes, err := json.Marshal(data)
		if err != nil {
			return warnings, err
		}

		req, err := http.NewRequest(http.MethodPost, "https://"+ua.tenantURL+"/api/v1/authenticate/mfa", bytes.NewReader(authPayloadBytes))
		if err != nil {
			return warnings, err
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return warnings, err
		}
		defer must(resp.Body.Close)

		err = json.NewDecoder(resp.Body).Decode(&authResponse)
		if err != nil {
			return warnings, err
		}

		if authResponse.Status == "error" {
			return warnings, errors.New(authResponse.Message)
		}
	}

	ua.authToken = authResponse.Token

	ua.expires, err = getExpirationFromToken(ua.authToken)

	if err != nil {
		return warnings, fmt.Errorf("unable to extract expiration from token: %w", err)
	}

	if ua.onRenewFunc != nil {
		// TODO catch this error?
		err = ua.onRenewFunc(ua.authToken, ua.expires)
		if err != nil {
			warnings = append(warnings, fmt.Errorf("error calling OnRenewFunc: %w", err))
		}
	}

	return warnings, nil
}

func getExpirationFromToken(token string) (time.Time, error) {
	// Strip the payload out of the JWT and decode it
	var jwt jwtPayload

	payload := strings.Split(token, ".")[1]
	// Fix the padding so it decodes correctly
	payload += strings.Repeat("=", ((len(payload)/3)+1)*3-len(payload))

	dst := make([]byte, base64.StdEncoding.DecodedLen(len(payload)))

	n, err := base64.StdEncoding.Decode(dst, []byte(payload))
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to decode jwt payload: %w", err)
	}

	dst = dst[:n]

	err = json.Unmarshal(dst, &jwt)
	if err != nil {
		return time.Time{}, fmt.Errorf("auth token jwt isn't expected format: %w", err)
	}

	expires := time.Unix(jwt.Exp, 0)
	slog.Debug("Token",
		"expires", jwt.Exp,
		"expires", expires,
	)

	return expires, nil
}

func (ua *UserAgent) checkExpired() ([]error, error) {
	var warnings []error

	var err error
	if ua.expires.IsZero() {
		// This is probably due to being in the Login process, so abort now
		return warnings, err
	}

	startRenew := ua.expires.Add(-2 * time.Minute)
	if time.Now().Before(startRenew) { // not time to renew yet
		return warnings, err
	}

	if time.Now().After(ua.expires) { // too late, can't renew now
		return warnings, err
	}

	slog.Debug("Renewing token",
		"expires", ua.expires,
	)

	response := struct {
		Status   string `json:"status"`
		TenantID int    `json:"tenant_id"`
		Token    string `json:"token"`
		Cookie   string `json:"cookie"`
	}{}
	// PUT to /v1/token/refresh to get an updated token

	// Simplified version of apiCall
	fullpath := "https://" + ua.tenantURL + "/api/" + "v1/token/refresh"

	req, err := http.NewRequest(http.MethodPut, fullpath, nil)
	if err != nil {
		return warnings, err
	}

	req.Header.Set("Authorization", "Bearer "+ua.authToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return warnings, err
	}

	defer must(resp.Body.Close)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return warnings, err
	}

	err = json.Unmarshal(body, &response)

	if err != nil {
		warnings = append(warnings, fmt.Errorf("unable to refresh token: %w", err))

		return warnings, err
	}

	// Save the token to the useragent

	ua.authTokenMutex.Lock()
	ua.authToken = response.Token
	ua.authTokenMutex.Unlock()

	// Save the expiration from the token to the useragent

	ua.expires, err = getExpirationFromToken(ua.authToken)
	if err != nil {
		return warnings, err
	}

	if ua.onRenewFunc != nil {
		// TODO catch this error?
		warning := ua.onRenewFunc(ua.authToken, ua.expires)
		if warning != nil {
			warnings = append(warnings, warning)
		}
	}

	return warnings, err
}

func (ua *UserAgent) apiGet(path string, response interface{}) (string, error) {
	_, err := ua.checkExpired()
	if err != nil {
		return "", err
	}

	ua.authTokenMutex.Lock()
	defer ua.authTokenMutex.Unlock()

	fullpath := "https://" + ua.tenantURL + "/api/" + path
	slog.Debug("Getting from API",
		"url", fullpath,
	)

	req, err := http.NewRequest(http.MethodGet, fullpath, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+ua.authToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer must(resp.Body.Close)

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

func (ua *UserAgent) apiCall(method, path string, body interface{}, response interface{}) (string, error) {
	_, err := ua.checkExpired()
	if err != nil {
		return "", err
	}

	ua.authTokenMutex.Lock()
	defer ua.authTokenMutex.Unlock()

	fullpath := "https://" + ua.tenantURL + "/api/" + path
	slog.Debug("Posting to API",
		"url", fullpath,
	)

	reqBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(method, fullpath, bytes.NewReader(reqBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+ua.authToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer must(resp.Body.Close)

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if response != nil {
		err = json.Unmarshal(bodyResp, response)
	}

	return string(bodyResp), err
}

func must(f func() error) {
	if err := f(); err != nil {
		panic(err)
	}
}
