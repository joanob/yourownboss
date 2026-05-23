package http

// UpdateMeRequest is the request body for updating user profile.
type UpdateMeRequest struct {
	Timezone string `json:"timezone" validate:"required"`
}
