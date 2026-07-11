package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role values. "advocate" is the lawyer/owner role created via public
// signup; "client" is only ever created by a lawyer through the Client
// module (see ClientService.Create) — there is no public client signup.
// "admin" (RoleAdmin) is never assignable from the app at all — there is no
// signup/creation path anywhere in the API for it. An admin account can only
// come from a direct database insert/update, e.g.:
//
//	UPDATE users SET role = 'admin' WHERE email = 'you@example.com';
const (
	RoleAdvocate = "advocate"
	RoleClient   = "client"
	RoleAdmin    = "admin"
)

type User struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string    `json:"name" gorm:"not null"`
	Email     string    `json:"email" gorm:"unique;not null"`
	Password  string    `json:"-" gorm:"not null"`
	Phone     string    `json:"phone"`
	Role      string    `json:"role" gorm:"default:advocate"`
	AvatarURL string    `json:"avatar_url"`
	Bio       string    `json:"bio"`
	LawFirm   string    `json:"law_firm"`
	BarNumber string    `json:"bar_number"`
	// GSTNumber, if set, is printed on every generated invoice PDF.
	GSTNumber     string `json:"gst_number"`
	SignaturePath string `json:"-"`
	// Payment Details (Lawyer Profile/Settings) — the source of truth for
	// where a client's UPI payment on an invoice should go. UpiID is the
	// only field required to actually accept a payment; the rest are
	// informational/for the lawyer's own records.
	AccountHolderName string `json:"account_holder_name"`
	BankName          string `json:"bank_name"`
	AccountNumber     string `json:"account_number"`
	IFSCCode          string `json:"ifsc_code"`
	UpiID             string `json:"upi_id"`
	// Presence: IsOnline is true from the moment of login until logout (or
	// an explicit disconnect) — a real login-session concept, deliberately
	// NOT tied to whether a WebSocket happens to be connected at any given
	// instant (that connection drops/reconnects constantly on Android as
	// the app backgrounds, which made "online" appear false almost always).
	// LastSeenAt is updated alongside every IsOnline transition.
	IsOnline   bool       `json:"is_online" gorm:"default:false;index"`
	LastSeenAt *time.Time `json:"last_seen_at"`
	// MustChangePassword is set true whenever a login password is assigned by
	// someone other than the account owner (a client's temporary password,
	// set by their lawyer at Client creation/reset) and cleared the moment
	// the user successfully changes their own password. The client app
	// blocks dashboard access and forces the Change Password screen while
	// this is true — see AuthService.ChangePassword and login handling.
	MustChangePassword bool `json:"must_change_password" gorm:"default:false"`
	// DeviceToken is this user's current Firebase Cloud Messaging token —
	// single most-recent device only (matches the existing single-value
	// IsOnline/LastSeenAt style). Re-registered every app launch/login, so a
	// user switching devices simply overwrites it; a stale token just fails
	// silently at send time (see PushService).
	DeviceToken string         `json:"-"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// UpdateDeviceTokenRequest registers/refreshes this user's FCM token for
// push notifications. An empty string is valid — it's how the app clears
// the token on logout so a signed-out device stops receiving pushes.
type UpdateDeviceTokenRequest struct {
	DeviceToken string `json:"device_token"`
}

type SignupRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Phone    string `json:"phone"`
}

type LoginRequest struct {
	// Not restricted to email format: a client's Login ID (set by their
	// lawyer) may be a mobile number instead of an email address.
	Email      string `json:"email" binding:"required"`
	Password   string `json:"password" binding:"required"`
	RememberMe bool   `json:"remember_me"`
}

type UpdateProfileRequest struct {
	Name      string `json:"name" binding:"required"`
	Phone     string `json:"phone"`
	AvatarURL string `json:"avatar_url"`
	Bio       string `json:"bio"`
	LawFirm   string `json:"law_firm"`
	BarNumber string `json:"bar_number"`
	// GSTNumber and Payment Details — pointers so a caller that omits them
	// (e.g. the plain Edit Profile screen, which only sends
	// name/phone/bio/etc., or the Payment Details screen, which never sends
	// GSTNumber) leaves the lawyer's existing values untouched instead of
	// wiping them with empty strings.
	GSTNumber         *string `json:"gst_number"`
	AccountHolderName *string `json:"account_holder_name"`
	BankName          *string `json:"bank_name"`
	AccountNumber     *string `json:"account_number"`
	IFSCCode          *string `json:"ifsc_code"`
	UpiID             *string `json:"upi_id"`
}

type AuthResponse struct {
	Message      string `json:"message"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}
