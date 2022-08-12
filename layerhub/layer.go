package layerhub

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

type LayerType string

const (
	LayerStaticVector LayerType = "StaticVector"
	LayerStaticGroup  LayerType = "StaticGroup"
	LayerDynamicGroup LayerType = "DynamicGroup"
	LayerStaticPath   LayerType = "StaticPath"
	LayerDynamicPath  LayerType = "DynamicPath"
	LayerStaticImage  LayerType = "StaticImage"
	LayerStaticVideo  LayerType = "StaticVideo"
	LayerStaticAudio  LayerType = "StaticAudio"
	LayerDynamicImage LayerType = "DynamicImage"
	LayerStaticText   LayerType = "StaticText"
	LayerDynamicText  LayerType = "DynamicText"
	LayerBackground   LayerType = "Background"
	LayerFrame        LayerType = "Frame"
	LayerGroup        LayerType = "Group"
)

var (
	_ Viewable = (*Layer)(nil)
)

type Viewable interface {
	SetPreviewURL(string)
}

type BaseLayer struct {
	ID string `json:"id,omitempty" bson:"_id"`

	Name        string         `json:"name,omitempty" bson:"name,omitempty"`
	Type        LayerType      `json:"type" bson:"type"`
	Top         float64        `json:"top" bson:"top"`
	Left        float64        `json:"left" bson:"left"`
	Angle       float64        `json:"angle" bson:"angle"`
	Width       float64        `json:"width" bson:"width"`
	Height      float64        `json:"height" bson:"height"`
	OriginX     string         `json:"originX,omitempty" bson:"originX,omitempty"`
	OriginY     string         `json:"originY,omitempty" bson:"originY,omitempty"`
	ScaleX      float64        `json:"scaleX" bson:"scaleX"`
	ScaleY      float64        `json:"scaleY" bson:"scaleY"`
	Opacity     float64        `json:"opacity" bson:"opacity"`
	FlipX       bool           `json:"flipX" bson:"flipX"`
	FlipY       bool           `json:"flipY" bson:"flipY"`
	SkewX       float64        `json:"skewX" bson:"skewX"`
	SkewY       float64        `json:"skewY" bson:"skewY"`
	Stroke      string         `json:"stroke,omitempty" bson:"stroke,omitempty"`
	StrokeWidth float64        `json:"strokeWidth" bson:"strokeWidth"`
	Visible     bool           `json:"visible" bson:"visible"`
	Shadow      *Shadow        `json:"shadow,omitempty" bson:"shadow,omitempty"`
	Duration    float64        `json:"duration" bson:"duration"`
	Metadata    map[string]any `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

// Layer is a simplified representation of a Fabric.js object/layer
type Layer struct {
	BaseLayer `bson:"inline"`

	GroupMetadata `bson:"inline"`

	Props any `json:"-"`
}

type KeyValue struct {
	Key   string `json:"key" bson:"key"`
	Value string `json:"value:" bson:"value"`
}

type TimeRange struct {
	From any `json:"from" bson:"from"`
	To   any `json:"to" bson:"to"`
}

type Shadow struct {
	Color        string  `json:"color,omitempty" bson:"color,omitempty"`
	Blur         float64 `json:"blur" bson:"blur"`
	OffsetX      float64 `json:"offsetX" bson:"offsetX"`
	OffsetY      float64 `json:"offsetY" bson:"offsetY"`
	AffectStroke bool    `json:"affectStroke" bson:"affectStroke"`
	NonScaling   bool    `json:"nonScaling" bson:"nonScaling"`
	Enabled      bool    `json:"enabled" bson:"enabled"`
}

type StaticImageProps struct {
	Src   string  `json:"src" bson:"src"`
	CropX float64 `json:"cropX" bson:"cropX"`
	CropY float64 `json:"cropY" bson:"cropY"`
}

type StaticAudioProps struct {
	Src         string    `json:"src" bson:"src"`
	SpeedFactor float64   `json:"speedFactor" bson:"speedFactor"`
	Between     TimeRange `json:"between" bson:"between"`
	Cut         TimeRange `json:"cut" bson:"cut"`
}

type StaticVideoProps struct {
	Src         string    `json:"src" bson:"src"`
	SpeedFactor float64   `json:"speedFactor" bson:"speedFactor"`
	Between     TimeRange `json:"between" bson:"between"`
	Cut         TimeRange `json:"cut" bson:"cut"`
}

type DynamicImageProps struct {
	Key string `json:"key" bson:"key"`
}

type StaticTextProps struct {
	TextAlign   string      `json:"textAlign,omitempty" bson:"textAlign,omitempty"`
	FontURL     string      `json:"fontURL,omitempty" bson:"fontURL,omitempty"`
	FontFamily  string      `json:"fontFamily,omitempty" bson:"fontFamily,omitempty"`
	FontSize    float64     `json:"fontSize" bson:"fontSize"`
	FontWeight  interface{} `json:"fontWeight" bson:"fontWeight"`
	Charspacing float64     `json:"charspacing" bson:"charspacing"`
	Lineheight  float64     `json:"lineheight" bson:"lineheight"`
	Fill        string      `json:"fill,omitempty" bson:"fill,omitempty"`
	Text        string      `json:"text" bson:"text"`
}

type DynamicTextProps struct {
	KeyValues []KeyValue `json:"keyValues,omitempty" bson:"keyValues,omitempty"`
}

type StaticVectorProps struct {
	Src      string            `json:"src" bson:"src"`
	ColorMap map[string]string `json:"colorMap" bson:"colorMap"`
}

type BackgroundProps struct {
	Fill string `json:"fill,omitempty" bson:"fill,omitempty"`
}

type StaticPathProps struct {
	Path [][]any `json:"path" bson:"path"`
	Fill string  `json:"fill,omitempty" bson:"fill,omitempty"`
}

type GroupProps struct {
	Objects []*Layer `json:"objects"`
}

type GroupMetadata struct {
	Category   string   `json:"category,omitempty" bson:"category,omitempty"`
	PreviewURL string   `json:"previewURL,omitempty" bson:"previewURL,omitempty"`
	Types      []string `json:"types,omitempty" bson:"types,omitempty"`
}

func (m *GroupMetadata) SetPreviewURL(url string) {
	m.PreviewURL = url
}

func NewLayer(b BaseLayer) *Layer {
	l := &Layer{BaseLayer: b}

	switch b.Type {
	case LayerStaticText:
		l.Props = &StaticTextProps{}
	case LayerDynamicText:
		l.Props = &DynamicTextProps{}
	case LayerStaticImage:
		l.Props = &StaticImageProps{}
	case LayerStaticAudio:
		l.Props = &StaticAudioProps{}
	case LayerStaticVideo:
		l.Props = &StaticVideoProps{}
	case LayerDynamicImage:
		l.Props = &DynamicImageProps{}
	case LayerStaticVector:
		l.Props = &StaticVectorProps{}
	case LayerStaticPath:
		l.Props = &StaticPathProps{}
	case LayerBackground:
		l.Props = &BackgroundProps{}
	case LayerGroup:
		l.Props = &GroupProps{}
	default:
		panic("unknown layer type " + b.Type)
	}

	return l
}

// persistLayerResources uploads or saves layer resources that may expire, like images, videos, etc
func (c *Core) persistLayerResources(ctx context.Context, layers []*Layer) error {
	for _, layer := range layers {
		if layer.Type == LayerStaticImage {
			props, ok := layer.Props.(*StaticImageProps)
			if !ok {
				c.Logger.Errorf("layer with incorrect props: %v", layer.Props)
				continue
			}

			isThirdPartyImage := strings.Contains(props.Src, c.pexels.ImageDomain()) ||
				strings.Contains(props.Src, c.pixabay.ImageDomain())

			if !isThirdPartyImage {
				continue
			}

			c.Logger.Infof("uploading thirdparty image: %s", props.Src)

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, props.Src, nil)
			if err != nil {
				return err
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			img, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			propsURL, err := url.Parse(props.Src)
			if err != nil {
				return err
			}

			imgKey := string(UniqueID("img")) + path.Ext(propsURL.Path)
			imgURL, err := c.uploader.Upload(ctx, imgKey, img)
			if err != nil {
				return err
			}

			props.Src = imgURL
		}
	}

	return nil
}

// Unmarshals the layer and dynamically sets the props
func (l *Layer) UnmarshalJSON(data []byte) error {
	type layer Layer

	var raw layer
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	ll := NewLayer(raw.BaseLayer)
	ll.GroupMetadata = raw.GroupMetadata
	if err := json.Unmarshal(data, ll.Props); err != nil {
		return err
	}
	*l = *ll

	return nil
}

func (l *Layer) MarshalJSON() ([]byte, error) {
	type layer Layer

	var ll any
	switch p := l.Props.(type) {
	case *StaticTextProps:
		ll = struct {
			layer
			*StaticTextProps
		}{layer(*l), p}
	case *DynamicTextProps:
		ll = struct {
			layer
			*DynamicTextProps
		}{layer(*l), p}
	case *StaticImageProps:
		ll = struct {
			layer
			*StaticImageProps
		}{layer(*l), p}
	case *StaticAudioProps:
		ll = struct {
			layer
			*StaticAudioProps
		}{layer(*l), p}
	case *StaticVideoProps:
		ll = struct {
			layer
			*StaticVideoProps
		}{layer(*l), p}
	case *DynamicImageProps:
		ll = struct {
			layer
			*DynamicImageProps
		}{layer(*l), p}
	case *StaticVectorProps:
		ll = struct {
			layer
			*StaticVectorProps
		}{layer(*l), p}
	case *StaticPathProps:
		ll = struct {
			layer
			*StaticPathProps
		}{layer(*l), p}
	case *BackgroundProps:
		ll = struct {
			layer
			*BackgroundProps
		}{layer(*l), p}
	case *GroupProps:
		ll = struct {
			layer
			*GroupProps
		}{layer(*l), p}
	default:
		ll = struct {
			layer
		}{layer(*l)}
	}

	return json.Marshal(ll)
}

// Unmarshals the layer and dynamically sets the props
func (l *Layer) UnmarshalBSON(data []byte) error {
	type layer struct {
		BaseLayer     `bson:"inline"`
		GroupMetadata `bson:"inline"`
	}

	var raw layer
	if err := bson.Unmarshal(data, &raw); err != nil {
		return err
	}

	ll := NewLayer(raw.BaseLayer)
	ll.GroupMetadata = raw.GroupMetadata
	if err := bson.Unmarshal(data, ll.Props); err != nil {
		return err
	}
	*l = *ll

	return nil
}

func (l *Layer) MarshalBSON() ([]byte, error) {
	type layer struct {
		BaseLayer     `bson:"inline"`
		GroupMetadata `bson:"inline"`
		Props         map[string]any `bson:"inline"`
	}

	var ll = layer{
		BaseLayer:     l.BaseLayer,
		GroupMetadata: l.GroupMetadata,
		Props:         make(map[string]any),
	}

	if l.Props != nil {
		pb, err := bson.Marshal(l.Props)
		if err != nil {
			return nil, err
		}

		err = bson.Unmarshal(pb, &ll.Props)
		if err != nil {
			return nil, err
		}
	}

	return bson.Marshal(ll)
}
