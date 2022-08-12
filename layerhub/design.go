package layerhub

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/echovl/orderflo-dev/errors"
)

type Design interface {
	Key() string
}

// Template is a simplified representation of a Fabric.js canvas
type Template struct {
	ID          string    `json:"id" bson:"_id"`
	ShortID     string    `json:"short_id" bson:"short_id" db:"short_id"`
	Type        string    `json:"type" bson:"type" db:"type"`
	Name        string    `json:"name" bson:"name" db:"name"`
	Description string    `json:"description" bson:"description"`
	Published   bool      `json:"published" bson:"published" db:"published"`
	Tags        []string  `json:"tags" bson:"tags"`
	Colors      []string  `json:"colors" bson:"colors"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at" db:"updated_at"`

	// Layers is a collection of layers like StaticImage, StaticPath, etc.
	Layers []*Layer `json:"layers" bson:"layers"`

	Frame Frame `json:"frame" bson:"frame"`

	Metadata Metadata `json:"metadata" bson:"metadata"`

	// Preview is the rendered template's URL
	Preview string `json:"preview" bson:"preview" db:"preview"`
}

func NewTemplate() *Template {
	now := Now()
	return &Template{
		ID:        UniqueID("temp"),
		ShortID:   UniqueShortID(),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (t *Template) Key() string {
	return string(t.ID) + ".layerhub"
}

func (c *Core) uploadDesign(ctx context.Context, dsg Design) {
	json, err := json.Marshal(dsg)
	if err != nil {
		c.Logger.Errorf("design upload: %s", err)
		return
	}
	_, err = c.uploader.Upload(ctx, dsg.Key(), json)
	if err != nil {
		c.Logger.Errorf("design upload: %s", err)
		return
	}
}

func (c *Core) PutTemplate(ctx context.Context, template *Template) error {
	err := c.persistLayerResources(ctx, template.Layers)
	if err != nil {
		return err
	}

	url, err := c.renderer.Render(ctx, template, nil)
	if err != nil {
		return err
	}
	template.Preview = url

	if err := c.db.PutTemplate(ctx, template); err != nil {
		return err
	}

	go c.uploadDesign(ctx, template)

	return nil
}

func (c *Core) FindTemplates(ctx context.Context, filter *Filter) ([]Template, int, error) {
	templates, err := c.db.FindTemplates(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	count, err := c.db.CountTemplates(ctx, filter.WithoutPagination())
	if err != nil {
		return nil, 0, err
	}

	return templates, count, nil
}

func (c *Core) GetTemplate(ctx context.Context, id string) (*Template, error) {
	templates, err := c.db.FindTemplates(ctx, &Filter{RegularOrShortID: id, Limit: 1})
	if err != nil {
		return nil, err
	}

	if len(templates) == 0 {
		return nil, errors.NotFound(fmt.Sprintf("template '%s' not found", id))
	}

	content, err := c.uploader.Download(ctx, templates[0].Key())
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(content, &templates[0])
	if err != nil {
		return nil, err
	}

	return &templates[0], nil
}

func (c *Core) DeleteTemplate(ctx context.Context, id string) error {
	// TODO: Delete template from s3?
	err := c.db.DeleteTemplate(ctx, id)
	if err != nil {
		return err
	}
	err = c.db.DeleteFrame(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

// Project is a simplified representation of a Fabric.js canvas
type Project struct {
	ID          string    `json:"id"`
	ShortID     string    `json:"short_id" db:"short_id"`
	Type        string    `json:"type" db:"type"`
	Name        string    `json:"name" db:"name"`
	UserID      string    `json:"user_id" db:"user_id"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// Layers is a collection of layers like StaticImage, StaticPath, etc.
	Layers []*Layer `json:"layers"`

	Frame Frame `json:"frame" bson:"frame"`

	// Preview is the rendered template's URL
	Preview string `json:"preview" bson:"preview" db:"preview"`
}

func NewProject() *Project {
	now := Now()
	return &Project{
		ID:        UniqueID("proj"),
		ShortID:   UniqueShortID(),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (p *Project) Key() string {
	return string(p.ID) + ".layerhub"
}

func (c *Core) PutProject(ctx context.Context, project *Project) error {
	t1 := time.Now()

	err := c.persistLayerResources(ctx, project.Layers)
	if err != nil {
		return err
	}

	url, err := c.renderer.Render(ctx, project, nil)
	if err != nil {
		return err
	}
	project.Preview = url

	t2 := time.Now()

	if err := c.db.PutProject(ctx, project); err != nil {
		return err
	}

	t3 := time.Now()

	go c.uploadDesign(ctx, project)

	c.Logger.Infof("render: %v", t2.Sub(t1).Milliseconds())
	c.Logger.Infof("db: %v", t3.Sub(t2).Milliseconds())

	return nil
}

func (c *Core) FindProjects(ctx context.Context, filter *Filter) ([]Project, int, error) {
	projects, err := c.db.FindProjects(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	count, err := c.db.CountProjects(ctx, filter.WithoutPagination())
	if err != nil {
		return nil, 0, err
	}

	return projects, count, nil
}

func (c *Core) GetProject(ctx context.Context, id string) (*Project, error) {
	projects, err := c.db.FindProjects(ctx, &Filter{RegularOrShortID: id, Limit: 1})
	if err != nil {
		return nil, err
	}

	if len(projects) == 0 {
		return nil, errors.NotFound(fmt.Sprintf("project '%s' not found", id))
	}

	content, err := c.uploader.Download(ctx, projects[0].Key())
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(content, &projects[0])
	if err != nil {
		return nil, err
	}

	return &projects[0], nil
}

func (c *Core) DeleteProject(ctx context.Context, id string) error {
	// TODO: Delete project from s3?
	err := c.db.DeleteProject(ctx, id)
	if err != nil {
		return err
	}
	err = c.db.DeleteFrame(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

type Metadata struct {
	ID          string `json:"id,omitempty" db:"id"`
	License     string `json:"license" db:"license"`
	Orientation string `json:"orientation" db:"orientation"`
}

type FrameVisibility string

const (
	FramePublic  FrameVisibility = "public"
	FramePrivate FrameVisibility = "private"
)

type FrameUnit string

const (
	Centimeters FrameUnit = "cm"
	Pixels      FrameUnit = "px"
	Inches      FrameUnit = "in"
)

type Frame struct {
	ID         string          `json:"id,omitempty" db:"id"`
	Name       string          `json:"name,omitempty" db:"name"`
	Visibility FrameVisibility `json:"visibility,omitempty" db:"visibility"`
	Width      float64         `json:"width" db:"width"`
	Height     float64         `json:"height" db:"height"`
	Unit       FrameUnit       `json:"unit" db:"unit"`
	Preview    string          `json:"preview" db:"preview"`
}

func NewFrame() *Frame {
	return &Frame{
		ID: UniqueID("frame"),
	}
}

func (c *Core) PutFrame(ctx context.Context, frame *Frame) error {
	return c.db.PutFrame(ctx, frame)
}

func (c *Core) GetFrame(ctx context.Context, id string) (*Frame, error) {
	frames, err := c.db.FindFrames(ctx, &Filter{ID: id, Limit: 1})
	if err != nil {
		return nil, err
	}

	if len(frames) == 0 {
		return nil, errors.NotFound(fmt.Sprintf("frame '%s' not found", id))
	}

	return &frames[0], nil
}

func (c *Core) FindFrames(ctx context.Context, filter *Filter) ([]Frame, int, error) {
	frames, err := c.db.FindFrames(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	count, err := c.db.CountFrames(ctx, filter.WithoutPagination())
	if err != nil {
		return nil, 0, err
	}

	return frames, count, nil
}

func (c *Core) DeleteFrame(ctx context.Context, id string) error {
	return c.db.DeleteFrame(ctx, id)
}

// A wrapper for Scenify layers
type Component struct {
	ID        string    `json:"id" bson:"_id"`
	Name      string    `json:"name" db:"name"`
	Preview   string    `json:"preview" db:"preview"`
	UserID    string    `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Layers is a collection of layers like StaticImage, StaticPath, etc.
	Layers []*Layer `json:"layers" bson:"layers"`

	Metadata map[string]any `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

func NewComponent() *Component {
	now := Now()
	return &Component{
		ID:        UniqueID("comp"),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (c *Component) Key() string {
	return string(c.ID) + ".layerhub"
}

func (c *Core) PutComponent(ctx context.Context, comp *Component) error {
	template := &Template{
		ID: comp.ID,
		Frame: Frame{
			Width:  comp.Layers[0].Width,
			Height: comp.Layers[0].Height,
		},
		Layers: comp.Layers,
	}

	preview, err := c.renderer.Render(ctx, template, nil)
	if err != nil {
		return err
	}
	comp.Preview = preview
	if err := c.db.PutComponent(ctx, comp); err != nil {
		return err
	}

	go c.uploadDesign(ctx, comp)

	return nil
}

func (c *Core) GetComponent(ctx context.Context, id string) (*Component, error) {
	comps, err := c.db.FindComponents(ctx, &Filter{ID: id})
	if err != nil {
		return nil, err
	}

	if len(comps) == 0 {
		return nil, errors.NotFound(fmt.Sprintf("components '%s' not found", id))
	}

	content, err := c.uploader.Download(ctx, comps[0].Key())
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(content, &comps[0])
	if err != nil {
		return nil, err
	}

	return &comps[0], nil
}

func (c *Core) FindComponents(ctx context.Context, filter *Filter) ([]Component, int, error) {
	comps, err := c.db.FindComponents(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	count, err := c.db.CountComponents(ctx, filter.WithoutPagination())
	if err != nil {
		return nil, 0, err
	}

	return comps, count, nil
}

func (c *Core) DeleteComponent(ctx context.Context, id string) error {
	err := c.db.DeleteComponent(ctx, id)
	if err != nil {
		return err
	}
	return nil
}
