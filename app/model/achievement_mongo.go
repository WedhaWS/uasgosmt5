package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Collection: achievements
type Achievement struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	StudentID       string             `bson:"studentId" json:"studentId"` // UUID Reference
	AchievementType string             `bson:"achievementType" json:"achievementType"`
	Title           string             `bson:"title" json:"title"`
	Description     string             `bson:"description" json:"description"`

	// Field dinamis
	Details AchievementDetails `bson:"details" json:"details"`

	Attachments []AchievementAttachment `bson:"attachments" json:"attachments"`
	Tags        []string                `bson:"tags" json:"tags"`
	Points      int                     `bson:"points" json:"points"`

	// Soft delete fields
	IsDeleted bool       `bson:"isDeleted,omitempty" json:"isDeleted,omitempty"`
	DeletedAt *time.Time `bson:"deletedAt,omitempty" json:"deletedAt,omitempty"`

	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}

// Detail Prestasi (SRS Halaman 6)
type AchievementDetails struct {
	// Competition
	CompetitionName  string `bson:"competitionName,omitempty" json:"competitionName,omitempty"`
	CompetitionLevel string `bson:"competitionLevel,omitempty" json:"competitionLevel,omitempty"`
	Rank             int    `bson:"rank,omitempty" json:"rank,omitempty"`
	MedalType        string `bson:"medalType,omitempty" json:"medalType,omitempty"`

	// Publication
	PublicationType  string   `bson:"publicationType,omitempty" json:"publicationType,omitempty"`
	PublicationTitle string   `bson:"publicationTitle,omitempty" json:"publicationTitle,omitempty"`
	Authors          []string `bson:"authors,omitempty" json:"authors,omitempty"`
	Publisher        string   `bson:"publisher,omitempty" json:"publisher,omitempty"`
	ISSN             string   `bson:"issn,omitempty" json:"issn,omitempty"`

	// Organization
	OrganizationName string         `bson:"organizationName,omitempty" json:"organizationName,omitempty"`
	Position         string         `bson:"position,omitempty" json:"position,omitempty"`
	Period           *PeriodDetails `bson:"period,omitempty" json:"period,omitempty"`

	// Certification
	CertificationName   string     `bson:"certificationName,omitempty" json:"certificationName,omitempty"`
	IssuedBy            string     `bson:"issuedBy,omitempty" json:"issuedBy,omitempty"`
	CertificationNumber string     `bson:"certificationNumber,omitempty" json:"certificationNumber,omitempty"`
	ValidUntil          *time.Time `bson:"validUntil,omitempty" json:"validUntil,omitempty"`

	// General / Other
	EventDate    *time.Time             `bson:"eventDate,omitempty" json:"eventDate,omitempty"`
	Location     string                 `bson:"location,omitempty" json:"location,omitempty"`
	Organizer    string                 `bson:"organizer,omitempty" json:"organizer,omitempty"`
	Score        float64                `bson:"score,omitempty" json:"score,omitempty"`
	CustomFields map[string]interface{} `bson:"customFields,omitempty" json:"customFields,omitempty"`
}

type PeriodDetails struct {
	Start time.Time `bson:"start" json:"start"`
	End   time.Time `bson:"end" json:"end"`
}

type AchievementAttachment struct {
	FileName   string    `bson:"fileName" json:"fileName"`
	FileURL    string    `bson:"fileUrl" json:"fileUrl"`
	FileType   string    `bson:"fileType" json:"fileType"`
	UploadedAt time.Time `bson:"uploadedAt" json:"uploadedAt"`
}
