// Copyright (c) 2025 Matt Robinson brimstone@the.narro.ws

package plextrac

import (
	"fmt"
	"strconv"
)

type User struct {
	ID    int64
	Name  string
	Email string
}
type userResponse struct {
	Data struct {
		ActivatedAt            int64         `json:"activatedAt"`
		AuthenticationProvider string        `json:"authentication_provider"`
		CreatedAt              int64         `json:"createdAt"`
		Cuid                   string        `json:"cuid"`
		DateFormat             string        `json:"dateFormat"`
		DefaultGroup           bool          `json:"default_group"`
		Disabled               bool          `json:"disabled"`
		DocType                string        `json:"doc_type"`
		Email                  string        `json:"email"`
		FailedLogins           int64         `json:"failedLogins"`
		First                  string        `json:"first"`
		FullName               string        `json:"fullName"`
		IsPaidUser             bool          `json:"isPaidUser"`
		Language               string        `json:"language"`
		Last                   string        `json:"last"`
		LastFailedLogin        int64         `json:"lastFailedLogin"`
		LastLogin              int64         `json:"lastLogin"`
		LicenseKeys            []interface{} `json:"licenseKeys"`
		Name                   struct {
			First string `json:"first"`
			Last  string `json:"last"`
		} `json:"name"`
		Roles                     []string `json:"roles"`
		TenantClassificationLevel string   `json:"tenantClassificationLevel"`
		TenantID                  int64    `json:"tenant_id"`
		UpdatedAt                 int64    `json:"updatedAt"`
		UserID                    int64    `json:"user_id"`
	} `json:"data"`
	DocID []int64 `json:"doc_id"`
	ID    string  `json:"id"`
}

func (ua *UserAgent) Users() ([]User, error) {
	var userResp []userResponse

	var users []User

	_, err := ua.apiGet(fmt.Sprintf("v1/tenant/%d/user/list", ua.GetTenantID()), &userResp)
	if err != nil {
		return nil, fmt.Errorf("unable to get users: %w", err)
	}

	for _, u := range userResp {
		var user User
		if s, err := strconv.Atoi(u.ID); err == nil {
			user.ID = int64(s)
		}

		user.Name = u.Data.FullName
		user.Email = u.Data.Email
		users = append(users, user)
	}

	return users, nil
}
