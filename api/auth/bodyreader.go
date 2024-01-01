package auth

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"

	"github.com/friendsofgo/errors"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/tedyst/licenta/api/auth/authbosswebauthn"
	"github.com/volatiletech/authboss/v3"
	"github.com/volatiletech/authboss/v3/defaults"
)

// FormValue types
const (
	FormValueEmail    = "email"
	FormValuePassword = "password"
	FormValueUsername = "username"

	FormValueConfirm      = "cnf"
	FormValueToken        = "token"
	FormValueCode         = "code"
	FormValueRecoveryCode = "recovery_code"
	FormValuePhoneNumber  = "phone_number"
)

// WebauthnValues from the login form
type WebauthnValues struct {
	defaults.HTTPFormValidator

	PID string

	CreationCredential  *protocol.ParsedCredentialCreationData
	CredentialAssertion *protocol.ParsedCredentialAssertionData
}

// GetPID from the values
func (u WebauthnValues) GetPID() string {
	return u.PID
}

func (u WebauthnValues) GetCreationCredential() protocol.ParsedCredentialCreationData {
	return *u.CreationCredential
}

func (u WebauthnValues) GetCredentialAssertion() protocol.ParsedCredentialAssertionData {
	return *u.CredentialAssertion
}

func (u WebauthnValues) GetShouldRemember() bool {
	rm, ok := u.Values[authboss.CookieRemember]
	return ok && rm == "true"
}

// authbossBodyReader reads forms from various pages and decodes
// them.
type authbossBodyReader struct {
	Rulesets  map[string][]defaults.Rules
	Confirms  map[string][]string
	Whitelist map[string][]string
}

// newAuthbossBodyReader creates a form reader with default validation rules
// and fields for each page. If no defaults are required, simply construct
// this using the struct members itself for more control.
func newAuthbossBodyReader() *authbossBodyReader {
	pid := "username"
	pidRules := defaults.Rules{
		FieldName: pid, Required: true,
		MatchError: "Usernames must only start with letters, and contain letters and numbers",
		MustMatch:  regexp.MustCompile(`(?i)[a-z][a-z0-9]?`),
	}

	passwordRule := defaults.Rules{
		FieldName:  "password",
		MinLength:  8,
		MinNumeric: 1,
		MinSymbols: 1,
		MinUpper:   1,
		MinLower:   1,
	}

	emailRule := defaults.Rules{
		FieldName:  "email",
		MinLength:  3,
		MinNumeric: 0,
		MinSymbols: 0,
		MinUpper:   0,
		MinLower:   0,
		Required:   true,
	}

	return &authbossBodyReader{
		Rulesets: map[string][]defaults.Rules{
			"login":                {pidRules},
			"register":             {pidRules, passwordRule, emailRule},
			"confirm":              {defaults.Rules{FieldName: FormValueConfirm, Required: true}},
			"recover_start":        {pidRules},
			"recover_end":          {passwordRule},
			"twofactor_verify_end": {defaults.Rules{FieldName: FormValueToken, Required: true}},
		},
		Confirms: map[string][]string{
			"register":    {FormValuePassword, authboss.ConfirmPrefix + FormValuePassword},
			"recover_end": {FormValuePassword, authboss.ConfirmPrefix + FormValuePassword},
		},
		Whitelist: map[string][]string{
			"register": {FormValueEmail, FormValuePassword},
		},
	}
}

// Read the form pages
func (h authbossBodyReader) Read(page string, r *http.Request) (authboss.Validator, error) {
	if page == authbosswebauthn.PageWebauthnSetupFinish {
		var values protocol.CredentialCreationResponse

		b, err := io.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			return nil, errors.Wrap(err, "failed to read http body")
		}

		if err = json.Unmarshal(b, &values); err != nil {
			return nil, errors.Wrap(err, "failed to parse json http body")
		}

		creds, err := values.Parse()
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse credential creation response")
		}

		return WebauthnValues{
			CreationCredential: creds,
		}, nil
	} else if page == authbosswebauthn.PageWebauthnLoginFinish {
		var values protocol.CredentialAssertionResponse

		b, err := io.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			return nil, errors.Wrap(err, "failed to read http body")
		}

		if err = json.Unmarshal(b, &values); err != nil {
			return nil, errors.Wrap(err, "failed to parse json http body")
		}

		creds, err := values.Parse()
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse credential creation response")
		}

		return WebauthnValues{
			CredentialAssertion: creds,
		}, nil
	}

	var values map[string]string

	b, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read http body")
	}

	if err = json.Unmarshal(b, &values); err != nil {
		return nil, errors.Wrap(err, "failed to parse json http body")
	}

	rules := h.Rulesets[page]
	confirms := h.Confirms[page]
	whitelist := h.Whitelist[page]

	switch page {
	case "confirm":
		return defaults.ConfirmValues{
			HTTPFormValidator: defaults.HTTPFormValidator{Values: values, Ruleset: rules},
			Token:             values[FormValueConfirm],
		}, nil
	case "login":
		return defaults.UserValues{
			HTTPFormValidator: defaults.HTTPFormValidator{Values: values, Ruleset: rules, ConfirmFields: confirms},
			PID:               values[FormValueUsername],
			Password:          values[FormValuePassword],
		}, nil
	case "recover_start":
		return defaults.RecoverStartValues{
			HTTPFormValidator: defaults.HTTPFormValidator{Values: values, Ruleset: rules, ConfirmFields: confirms},
			PID:               values[FormValueUsername],
		}, nil
	case "recover_middle":
		return defaults.RecoverMiddleValues{
			HTTPFormValidator: defaults.HTTPFormValidator{Values: values, Ruleset: rules, ConfirmFields: confirms},
			Token:             values[FormValueToken],
		}, nil
	case "recover_end":
		return defaults.RecoverEndValues{
			HTTPFormValidator: defaults.HTTPFormValidator{Values: values, Ruleset: rules, ConfirmFields: confirms},
			Token:             values[FormValueToken],
			NewPassword:       values[FormValuePassword],
		}, nil
	case "twofactor_verify_end":
		// Reuse ConfirmValues here, it's the same values we need
		return defaults.ConfirmValues{
			HTTPFormValidator: defaults.HTTPFormValidator{Values: values, Ruleset: rules, ConfirmFields: confirms},
			Token:             values[FormValueToken],
		}, nil
	case "totp2fa_confirm", "totp2fa_remove", "totp2fa_validate":
		return defaults.TwoFA{
			HTTPFormValidator: defaults.HTTPFormValidator{Values: values, Ruleset: rules, ConfirmFields: confirms},
			Code:              values[FormValueCode],
			RecoveryCode:      values[FormValueRecoveryCode],
		}, nil
	case "sms2fa_setup", "sms2fa_remove", "sms2fa_confirm", "sms2fa_validate":
		return defaults.SMSTwoFA{
			HTTPFormValidator: defaults.HTTPFormValidator{Values: values, Ruleset: rules, ConfirmFields: confirms},
			Code:              values[FormValueCode],
			PhoneNumber:       values[FormValuePhoneNumber],
			RecoveryCode:      values[FormValueRecoveryCode],
		}, nil
	case "register":
		arbitrary := make(map[string]string)

		for k, v := range values {
			for _, w := range whitelist {
				if k == w {
					arbitrary[k] = v
					break
				}
			}
		}

		return defaults.UserValues{
			HTTPFormValidator: defaults.HTTPFormValidator{Values: values, Ruleset: rules, ConfirmFields: confirms},
			PID:               values[FormValueUsername],
			Password:          values[FormValuePassword],
			Arbitrary:         arbitrary,
		}, nil
	case "webauthn_login":
		return WebauthnValues{
			HTTPFormValidator: defaults.HTTPFormValidator{Values: values, Ruleset: rules, ConfirmFields: confirms},
			PID:               values[FormValueUsername],
		}, nil
	default:
		return nil, errors.Errorf("failed to parse unknown page's form: %s", page)
	}
}

// URLValuesToMap helps create a map from url.Values
func URLValuesToMap(form url.Values) map[string]string {
	values := make(map[string]string)

	for k, v := range form {
		if len(v) != 0 {
			values[k] = v[0]
		}
	}

	return values
}
