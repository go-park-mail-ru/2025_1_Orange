package entity

import "time"

type NotificationType string

const (
	ApplyNotificationType NotificationType = "apply"
	DownloadResumeType    NotificationType = "download_resume"
)

var AllowedNotificationTypes = map[string]NotificationType{
	"apply":           ApplyNotificationType,
	"download_resume": DownloadResumeType,
}

type UserRole string

const (
	ApplicantRole UserRole = "applicant"
	EmployerRole  UserRole = "employer"
)

var AllowedUserRoles = map[string]UserRole{
	"applicant": ApplicantRole,
	"employer":  EmployerRole,
}

// easyjson:json
type Notification struct {
	ID           int              `json:"id"`
	Type         NotificationType `json:"type"`
	SenderID     int              `json:"sender_id"`
	SenderRole   UserRole         `json:"sender_role"`
	ReceiverID   int              `json:"receiver_id"`
	ReceiverRole UserRole         `json:"receiver_role"`
	ObjectID     int              `json:"object_id"`
	ResumeID     int              `json:"resume_id"`
	IsViewed     bool             `json:"is_viewed"`
	CreatedAt    time.Time        `json:"created_at"`
}

// easyjson:json
type NotificationPreview struct {
	ID            int              `json:"id"`
	Type          NotificationType `json:"type"`
	SenderID      int              `json:"sender_id"`
	ReceiverID    int              `json:"receiver_id"`
	ObjectID      int              `json:"object_id"`
	ResumeID      int              `json:"resume_id"`
	ApplicantName string           `json:"applicant_name"`
	EmployerName  string           `json:"employer_name"`
	Title         string           `json:"title"`
	IsViewed      bool             `json:"is_viewed"`
	CreatedAt     time.Time        `json:"created_at"`
}

// easyjson:json
type NotificationsList []*NotificationPreview
