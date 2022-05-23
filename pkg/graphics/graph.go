package graphics

import (
	"fmt"
	"golang.org/x/xerrors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"tinkoff-invest-bot/internal/config"

	"go.uber.org/zap"
)

const (
	graphsPath string = "./graphs/"
	detailPath string = "/detail/"
)

type graphHandler struct {
	logger *zap.Logger
}

// NewGraphHandler создаёт обработчик запросов для отображения графиков во время торговли
func NewGraphHandler(logger *zap.Logger) *graphHandler {
	return &graphHandler{
		logger: logger,
	}
}

func (h *graphHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error

	if len(r.URL.Path) >= len(detailPath) && r.URL.Path[:len(detailPath)] == detailPath {
		err = h.handleDetail(w, r)
	} else {
		err = h.handleMain(w)
	}

	if err != nil {
		h.logger.Error("Error in handling http request", zap.Error(err))
	}
}

// Возвращает страницу с выбранным графиком
func (h *graphHandler) handleDetail(w http.ResponseWriter, r *http.Request) error {
	err := config.CreateDirIfNotExist(graphsPath)
	if err != nil {
		return xerrors.Errorf("Can't create dir for graphics: %w", err)
	}

	path := r.URL.Path[len(detailPath):]
	path = path[:len(path)-len(".html")]
	path = strings.ReplaceAll(path, ".", "_")
	path = strings.ReplaceAll(path, "*", "_")
	path = strings.ReplaceAll(path, "/", "_")

	file, err := os.Open(graphsPath + path + ".html")
	if err != nil {
		_, _ = fmt.Fprintln(w, "<body>")
		_, _ = fmt.Fprintln(w, "No graph for this request")
		_, _ = fmt.Fprintln(w, "</body>")

		return xerrors.Errorf("cant find graphic with path: %w", err)
	}

	b, err := ioutil.ReadAll(file)
	if err != nil {
		if err = file.Close(); err != nil {
			return xerrors.Errorf("Can't close file with graphic: %w", err)
		}
		return xerrors.Errorf("Can't read file with graphic: %w", err)
	}
	_, _ = fmt.Fprintln(w, string(b))

	if err = file.Close(); err != nil {
		return xerrors.Errorf("Can't close file with graphic: %w", err)
	}
	return nil
}

// отображает страницу со списком доступных графиков
func (h *graphHandler) handleMain(w http.ResponseWriter) error {
	err := config.CreateDirIfNotExist(graphsPath)
	if err != nil {
		return xerrors.Errorf("Can't create dir for graphics: %w", err)
	}

	_, _ = fmt.Fprintln(w, "<body>")
	_, _ = fmt.Fprintln(w, "Welcome to main page<br/><br/>Please select trading bot config name for monitoring:<br/><br/>")

	files, err := ioutil.ReadDir(graphsPath)
	if err != nil {
		return xerrors.Errorf("Can't read directory with graphics: %w", err)
	}

	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".html") {
			s := fmt.Sprintf("<a href='http://localhost:8080/detail/%s'>%s</a><br/><br/>", f.Name(), f.Name())
			_, _ = fmt.Fprintln(w, s)
		}
	}
	_, err = fmt.Fprintln(w, "</body>")
	return err
}
