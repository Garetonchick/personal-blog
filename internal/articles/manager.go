package articles

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"
)

type Meta struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	CreationDate string `json:"creation-date"`
}

type Article struct {
	Meta
	Content []byte
}

const DATE_FORMAT = "02.01.2006"

var ErrNotExist = errors.New("article doesn't exist")

type ArticleDoesNotExist struct {
	ArticleID string
}

func (err *ArticleDoesNotExist) Error() string {
	return fmt.Sprintf("article %q doesn't exist", string(err.ArticleID))
}

func (err *ArticleDoesNotExist) Unwrap() error {
	return ErrNotExist
}

func loadMeta(path string) ([]Meta, error) {
	metaRaw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var metas []Meta
	err = json.Unmarshal(metaRaw, &metas)
	return metas, err
}

func loadArticle(dir string, meta *Meta) (*Article, error) {
	data, err := os.ReadFile(filepath.Join(dir, meta.ID+".md"))
	if err != nil {
		return nil, err
	}

	var a Article
	a.Meta = *meta
	a.Content = data

	return &a, nil
}

type Manager struct {
	mu       sync.Mutex
	workdir  string
	articles map[string]Article
}

func (m *Manager) List() []Article {
	m.mu.Lock()
	defer m.mu.Unlock()

	ars := make([]Article, 0, len(m.articles))
	for _, a := range m.articles {
		ars = append(ars, a)
	}

	slices.SortFunc(ars, func(a1, a2 Article) int {
		d1, err := time.Parse(DATE_FORMAT, a1.CreationDate)
		if err != nil {
			panic(err)
		}
		d2, err := time.Parse(DATE_FORMAT, a2.CreationDate)
		if err != nil {
			panic(err)
		}
		if d1 == d2 {
			return cmp.Compare(a1.ID, a2.ID)
		}
		if d1.Before(d2) {
			return 1
		}
		return -1
	})
	return ars
}

func (m *Manager) updateMeta(a *Article) error {
	metas, err := loadMeta(filepath.Join(m.workdir, "meta.json"))
	if err != nil {
		return err
	}

	found := false

	for i, meta := range metas {
		if meta.ID == a.ID {
			metas[i].CreationDate = a.CreationDate
			metas[i].Title = a.Title
			found = true
			break
		}
	}

	if !found {
		metas = append(metas, Meta{
			ID:           a.ID,
			CreationDate: a.CreationDate,
			Title:        a.Title,
		})
	}

	newMeta, err := json.Marshal(metas)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(m.workdir, "meta.json"), newMeta, 0644)
}

func (m *Manager) Save(a *Article) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.articles[a.ID] = *a

	path := filepath.Join(m.workdir, a.ID+".md")
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(a.Content)
	if err != nil {
		return err
	}

	return m.updateMeta(a)
}

func (m *Manager) Load(id string) (Article, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, ok := m.articles[id]
	if !ok {
		return Article{}, &ArticleDoesNotExist{ArticleID: id}
	}
	return a, nil
}

func NewManager(dir string) (*Manager, error) {
	workdir := filepath.Join(dir, "articles")

	if err := os.MkdirAll(workdir, 0755); err != nil {
		return nil, err
	}

	m := Manager{workdir: workdir, articles: make(map[string]Article)}

	metas, err := loadMeta(filepath.Join(workdir, "meta.json"))
	if errors.Is(err, os.ErrNotExist) {
		f, err := os.Create(filepath.Join(workdir, "meta.json"))
		if err != nil {
			return nil, err
		}
		f.Close()
	} else if err != nil {
		return nil, err
	}

	var errs []error

	for _, meta := range metas {
		a, err := loadArticle(workdir, &meta)
		if err != nil {
			errs = append(errs, err)
		}
		m.articles[a.ID] = *a
	}

	if len(errs) != 0 {
		return &m, errs[0]
	}

	return &m, nil
}
