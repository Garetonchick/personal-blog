package articles

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"sync"
	"time"
)

type Article struct {
	ID           int
	Title        string
	CreationDate time.Time
	Content      []byte
}

const DATE_FORMAT = "02.01.2006"

var filenameRegexp *regexp.Regexp

func init() {
	filenameRegexp = regexp.MustCompile(`^(.+)-(\d+)-([^-]+).md$`)
}

func (a *Article) Filename() string {
	return a.Title + "-" + strconv.Itoa(a.ID) + "-" + a.CreationDate.Format(DATE_FORMAT) + ".md"
}

func (a *Article) Save(dir string) error {
	path := filepath.Join(dir, a.Filename())
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(a.Content)
	if err != nil {
		return err
	}

	return nil
}

func parseArticleFilename(filename string) (*Article, error) {
	var err error
	var a Article

	m := filenameRegexp.FindStringSubmatchIndex(filename)
	if len(m) != 8 {
		return nil, errors.New("wrong filename format")
	}

	a.Title = filename[m[2]:m[3]]
	idS := filename[m[4]:m[5]]
	dateS := filename[m[6]:m[7]]

	a.CreationDate, err = time.Parse(DATE_FORMAT, dateS)
	if err != nil {
		return nil, err
	}

	a.ID, err = strconv.Atoi(idS)
	if err != nil {
		return nil, err
	}

	return &a, err
}

func Load(path string) (*Article, error) {
	a, err := parseArticleFilename(filepath.Base(path))
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	a.Content = data
	return a, nil
}

type Manager struct {
	mu       sync.Mutex
	workdir  string
	articles map[int]*Article
	maxID    int
}

func (m *Manager) List() []int {
	m.mu.Lock()
	defer m.mu.Unlock()

	ids := make([]int, 0, len(m.articles))
	for id := range m.articles {
		ids = append(ids, id)
	}

	slices.Sort(ids)
	slices.Reverse(ids)
	return ids
}

func (m *Manager) Save(a *Article) (id int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.maxID += 1
	a.ID = m.maxID
	m.articles[a.ID] = a

	path := filepath.Join(m.workdir, a.Filename())
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	_, err = file.Write(a.Content)
	if err != nil {
		return 0, err
	}

	return a.ID, nil
}

func (m *Manager) Load(id int) *Article {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.articles[id]
}

func NewManager(dir string) (*Manager, error) {
	workdir := filepath.Join(dir, "articles")

	if err := os.MkdirAll(workdir, 0755); err != nil {
		return nil, err
	}

	files, err := os.ReadDir(workdir)
	if err != nil {
		return nil, err
	}

	m := Manager{workdir: workdir, articles: make(map[int]*Article)}

	for _, file := range files {
		if !file.Type().IsRegular() {
			continue
		}

		info, err := file.Info()
		if err != nil {
			return nil, err
		}

		a, err := Load(filepath.Join(workdir, info.Name()))
		if err != nil {
			return nil, err
		}

		m.articles[a.ID] = a
	}

	return &m, nil
}
