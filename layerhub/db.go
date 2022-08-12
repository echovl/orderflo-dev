package layerhub

import (
	"context"
)

type Filter struct {
	Limit  int
	Offset int

	ID               string
	ShortID          string
	RegularOrShortID string
	UserID           string
	Email            string
	ApiToken         string
	PostscriptName   string
	FontEnabled      *bool
	Visibility       FrameVisibility
	UserSource       UserSource
}

func (f *Filter) WithoutPagination() *Filter {
	fc := *f
	fc.Limit = 0
	fc.Offset = 0
	return &fc
}

type DB interface {
	PutUser(ctx context.Context, user *User) error
	FindUsers(ctx context.Context, filter *Filter) ([]User, error)

	BatchCreateFonts(ctx context.Context, fonts []Font) error
	PutFont(ctx context.Context, font *Font) error
	FindFonts(ctx context.Context, filter *Filter) ([]Font, error)
	CountFonts(ctx context.Context, filter *Filter) (int, error)
	DeleteFont(ctx context.Context, id string) error

	PutTemplate(ctx context.Context, template *Template) error
	FindTemplates(ctx context.Context, filter *Filter) ([]Template, error)
	CountTemplates(ctx context.Context, filter *Filter) (int, error)
	DeleteTemplate(ctx context.Context, id string) error

	PutProject(ctx context.Context, template *Project) error
	FindProjects(ctx context.Context, filter *Filter) ([]Project, error)
	CountProjects(ctx context.Context, filter *Filter) (int, error)
	DeleteProject(ctx context.Context, id string) error

	PutFrame(ctx context.Context, frame *Frame) error
	FindFrames(ctx context.Context, filter *Filter) ([]Frame, error)
	CountFrames(ctx context.Context, filter *Filter) (int, error)
	DeleteFrame(ctx context.Context, id string) error

	PutComponent(ctx context.Context, component *Component) error
	FindComponents(ctx context.Context, filter *Filter) ([]Component, error)
	CountComponents(ctx context.Context, filter *Filter) (int, error)
	DeleteComponent(ctx context.Context, id string) error

	PutUpload(ctx context.Context, upload *Upload) error
	FindUploads(ctx context.Context, filter *Filter) ([]Upload, error)
	CountUploads(ctx context.Context, filter *Filter) (int, error)
	DeleteUpload(ctx context.Context, id string) error

	BatchCreateEnabledFonts(ctx context.Context, fonts []*EnabledFont) error
	FindEnabledFonts(ctx context.Context, userID string) ([]EnabledFont, error)
	BatchDeleteEnabledFonts(ctx context.Context, ids []string) error

	PutSubscriptionPlan(ctx context.Context, plan *SubscriptionPlan) error
	FindSubscriptionPlans(ctx context.Context) ([]SubscriptionPlan, error)
}

type JSONDB interface {
	PutTemplate(ctx context.Context, template *Template) error
	FindTemplates(ctx context.Context, filter *Filter) ([]Template, error)
	DeleteTemplate(ctx context.Context, id string) error

	Close(ctx context.Context) error
}
