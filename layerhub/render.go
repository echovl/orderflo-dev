package layerhub

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net"
	"os"
	"os/exec"

	"github.com/echovl/orderflo-dev/errors"
	"github.com/echovl/orderflo-dev/upload"
	"go.uber.org/zap"
)

type Renderer interface {
	Render(ctx context.Context, sch any, params map[string]any) (string, error)
	RawRender(ctx context.Context, sch any, params map[string]any) ([]byte, error)
}

type renderer struct {
	socket   string
	logger   *zap.SugaredLogger
	uploader upload.Uploader
}

func NewRenderer(socket string, logger *zap.SugaredLogger, uploader upload.Uploader) Renderer {
	return &renderer{
		socket:   socket,
		logger:   logger,
		uploader: uploader,
	}
}

func (r *renderer) Render(ctx context.Context, sch any, params map[string]any) (string, error) {
	key := UniqueID("preview") + ".png"
	img, err := r.RawRender(ctx, sch, params)
	if err != nil {
		return "", errors.Errorf("renderer: %s", err)
	}

	url, err := r.uploader.Upload(ctx, key, img)
	if err != nil {
		return "", errors.Errorf("renderer: %s", err)
	}

	return url, nil
}

func (r *renderer) RawRender(ctx context.Context, sch any, params map[string]any) ([]byte, error) {
	type request struct {
		Template any            `json:"template"`
		Params   map[string]any `json:"params"`
	}

	type response struct {
		Error string `json:"error"`
		Image string `json:"image"`
	}

	body, err := json.Marshal(request{sch, params})
	if err != nil {
		return nil, errors.Errorf("renderer: %s", err)
	}

	r.logger.Debugf("rendering template: %s", string(body))

	conn, err := net.Dial("unix", r.socket)
	if err != nil {
		return nil, errors.Errorf("renderer: %s", err)
	}
	defer conn.Close()

	_, err = conn.Write(body)
	if err != nil {
		return nil, errors.Errorf("renderer: %s", err)
	}

	if err := conn.(*net.UnixConn).CloseWrite(); err != nil {
		return nil, errors.Errorf("renderer: %s", err)
	}

	var resp response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return nil, errors.Errorf("renderer(json): %s", err)
	}

	if resp.Error != "" {
		return nil, errors.Errorf("renderer: %s", resp.Error)
	}

	img, err := base64.StdEncoding.DecodeString(resp.Image)
	if err != nil {
		return nil, errors.Errorf("renderer: %s", err)
	}

	return img, nil
}

func (c *Core) Render(ctx context.Context, sch any, params map[string]any) ([]byte, error) {
	return c.renderer.RawRender(ctx, sch, params)
}

type logWriter struct {
	l   *zap.SugaredLogger
	err bool
}

func (lw *logWriter) Write(p []byte) (int, error) {
	if lw.err {
		lw.l.Error(string(p))
	} else {
		lw.l.Info(string(p))
	}
	return len(p), nil
}

func RunRenderer(logger *zap.SugaredLogger) {
	baseDir := "./renderer/"

	os.Mkdir(baseDir+"fonts", 0777)

	infoLogger := &logWriter{logger, false}
	errLogger := &logWriter{logger, true}

	rend := exec.Command("node", "build/index.js")
	rend.Dir = baseDir
	rend.Stdout = infoLogger
	rend.Stderr = errLogger

	go func() {
		build := exec.Command("npm", "run", "build")
		build.Dir = baseDir

		err := build.Run()
		if err != nil {
			log.Panic(err)
		}

		err = rend.Run()
		if err != nil {
			log.Panic(err)
		}
	}()
}
