package models

import "time"

// UserProfile is the editable profile used by recommendations and the profile
// page. Skills and Experience are stored as ordered string lists.
type UserProfile struct {
	UserID     string    `json:"user_id"`
	Bio        string    `json:"bio"`
	Skills     []string  `json:"skills"`
	Experience []string  `json:"experience"`
	ResumeURL  string    `json:"resume_url"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// UserProfilePatch is a partial update. Nil fields are left unchanged, which
// lets callers update skills, experience, or bio independently without
// reading and rewriting the whole profile.
type UserProfilePatch struct {
	Bio        *string   `json:"bio,omitempty"`
	Skills     *[]string `json:"skills,omitempty"`
	Experience *[]string `json:"experience,omitempty"`
	ResumeURL  *string   `json:"resume_url,omitempty"`
}

// Empty reports whether the patch contains no requested changes.
func (p UserProfilePatch) Empty() bool {
	return p.Bio == nil && p.Skills == nil && p.Experience == nil && p.ResumeURL == nil
}
