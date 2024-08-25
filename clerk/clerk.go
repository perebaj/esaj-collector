// Package clerk gather all functions that deal with the clerk api. Including webhooks, authentication and user creation.
// Clerk is a third party service that provides authentication and user management.
package clerk

// WebHookEvent is the struct that represents the webhook event from clerk
type WebHookEvent struct {
	Data   Data   `json:"data"`
	Type   string `json:"type"`
	Object string `json:"object"`
}

// Data is the struct that represents the data from the webhook event
type Data struct {
	Birthday              string         `json:"birthday"`
	CreatedAt             int64          `json:"created_at"`
	EmailAddresses        []EmailAddress `json:"email_addresses"`
	ExternalAccounts      []any          `json:"external_accounts"`
	ExternalID            string         `json:"external_id"`
	FirstName             string         `json:"first_name"`
	Gender                string         `json:"gender"`
	ID                    string         `json:"id"`
	ImageURL              string         `json:"image_url"`
	LastName              string         `json:"last_name"`
	LastSignInAt          int64          `json:"last_sign_in_at"`
	Object                string         `json:"object"`
	PasswordEnabled       bool           `json:"password_enabled"`
	PhoneNumbers          []any          `json:"phone_numbers"`
	PrimaryEmailAddressID string         `json:"primary_email_address_id"`
	PrimaryPhoneNumberID  any            `json:"primary_phone_number_id"`
	PrimaryWeb3WalletID   any            `json:"primary_web3_wallet_id"`
	PrivateMetadata       struct {
	} `json:"private_metadata"`
	ProfileImageURL string `json:"profile_image_url"`
	PublicMetadata  struct {
	} `json:"public_metadata"`
	TwoFactorEnabled bool `json:"two_factor_enabled"`
	UnsafeMetadata   struct {
	} `json:"unsafe_metadata"`
	UpdatedAt   int64 `json:"updated_at"`
	Username    any   `json:"username"`
	Web3Wallets []any `json:"web3_wallets"`
}

// EmailAddress is the struct gathers user email address data
type EmailAddress struct {
	EmailAddress string `json:"email_address"`
	ID           string `json:"id"`
	LinkedTo     []any  `json:"linked_to"`
	Object       string `json:"object"`
	Verification struct {
		Status   string `json:"status"`
		Strategy string `json:"strategy"`
	} `json:"verification"`
}
