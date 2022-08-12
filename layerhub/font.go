package layerhub

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/echovl/orderflo-dev/errors"
)

type Font struct {
	ID             string `json:"id" db:"id"`
	Family         string `json:"family" db:"family"`
	FullName       string `json:"full_name" db:"full_name"`
	PostscriptName string `json:"postscript_name" db:"postscript_name"`
	Preview        string `json:"preview" db:"preview"`
	Style          string `json:"style" db:"style"`
	URL            string `json:"url" db:"url"`
	Category       string `json:"category" db:"category"`
	UserID         string `json:"user_id" db:"user_id"`
}

func NewFont() *Font {
	return &Font{
		ID: UniqueID("font"),
	}
}

type EnabledFont struct {
	ID     string `db:"id"`
	UserID string `db:"user_id"`
	FontID string `db:"font_id"`
}

func NewEnabledFont(userID, fontID string) *EnabledFont {
	return &EnabledFont{
		ID:     UniqueID("enabled_font"),
		UserID: userID,
		FontID: fontID,
	}
}

func (c *Core) EnableFonts(ctx context.Context, userID string, fontIDs []string) error {
	enabledFonts := []*EnabledFont{}
	fonts, err := c.db.FindEnabledFonts(ctx, userID)
	if err != nil {
		return err
	}

	for _, fontID := range fontIDs {
		fontExists := false
		for _, f := range fonts {
			if f.FontID == fontID {
				fontExists = true
				break
			}
		}

		if !fontExists {
			enabledFonts = append(enabledFonts, NewEnabledFont(userID, fontID))
		}
	}

	return c.db.BatchCreateEnabledFonts(ctx, enabledFonts)
}

func (c *Core) DisableFonts(ctx context.Context, userID string, fontIDs []string) error {
	fonts, err := c.db.FindEnabledFonts(ctx, userID)
	if err != nil {
		return nil
	}

	enabledFontIDs := []string{}
	for _, font := range fonts {
		for _, fontID := range fontIDs {
			if font.FontID == fontID {
				enabledFontIDs = append(enabledFontIDs, font.ID)
			}
		}
	}
	return c.db.BatchDeleteEnabledFonts(ctx, enabledFontIDs)
}

func (c *Core) PutFont(ctx context.Context, font *Font) error {
	err := c.buildFont(font)
	if err != nil {
		return err
	}
	return c.db.PutFont(ctx, font)
}

func (c *Core) GetFont(ctx context.Context, id string) (*Font, error) {
	fonts, err := c.db.FindFonts(ctx, &Filter{ID: id, Limit: 1})
	if err != nil {
		return nil, err
	}

	if len(fonts) == 0 {
		return nil, errors.NotFound(fmt.Sprintf("font '%s' not found", id))
	}

	return &fonts[0], nil
}

func (c *Core) FindFonts(ctx context.Context, filter *Filter) ([]Font, int, error) {
	fonts, err := c.db.FindFonts(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	count, err := c.db.CountFonts(ctx, filter.WithoutPagination())
	if err != nil {
		return nil, 0, err
	}

	return fonts, count, nil
}

func (c *Core) DeleteFont(ctx context.Context, id string) error {
	return c.db.DeleteFont(ctx, id)
}

func (c *Core) buildFont(font *Font) error {
	if font.URL == "" {
		return nil
	}

	fontFile, err := os.CreateTemp(os.TempDir(), "tmp")
	if err != nil {
		return err
	}
	defer fontFile.Close()
	defer os.Remove(fontFile.Name())

	resp, err := http.Get(font.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(fontFile, resp.Body)
	if err != nil {
		return err
	}

	scanOut := &bytes.Buffer{}
	fcScan := exec.Command("fc-scan", fontFile.Name())
	fcScan.Stdout = scanOut
	fcScan.Stderr = os.Stderr

	err = fcScan.Run()
	if err != nil {
		return err
	}

	scanLines := strings.Split(scanOut.String(), "\n")
	for _, line := range scanLines {
		if strings.Contains(line, "postscriptname") {
			postscriptName := regexp.MustCompile(`".*"`).FindAll([]byte(line), 1)
			if len(postscriptName) == 1 {
				font.PostscriptName = strings.Trim(string(postscriptName[0]), `"`)
			}
		}
	}

	return nil
}
