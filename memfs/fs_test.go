package memfs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/transientvariable/fs-go"
	"github.com/transientvariable/log-go"
	"github.com/transientvariable/support-go"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	gofs "io/fs"
)

const (
	testDataDir = "../testdata"
)

// MemFSTestSuite ...
type MemFSTestSuite struct {
	suite.Suite
	mfs       fs.FS
	files     map[string]gofs.FileInfo
	filePaths []string
}

func NewMemFSTestSuite() *MemFSTestSuite {
	return &MemFSTestSuite{}
}

func (t *MemFSTestSuite) SetupTest() {
	if err := log.SetDefault(log.New(log.WithLevel("debug"))); err != nil {
		t.T().Fatal(err)
	}

	mfs, err := New()
	if err != nil {
		t.T().Fatal(err)
	}
	t.mfs = mfs

	dir, err := os.Getwd()
	log.Info("")

	if err != nil {
		t.T().Fatal(err)
	}
	dir = filepath.Join(dir, testDataDir)

	log.Info("[memfs_test]", log.String("test_data_dir", dir))

	t.files = make(map[string]gofs.FileInfo)
	err = filepath.Walk(dir, func(path string, fi gofs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !fi.IsDir() && fi.Name() != "." {
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			filePath := strings.TrimPrefix(path, dir+"/")

			log.Info("[memfs_test] writing test file",
				log.String("file_path", filePath),
				log.Int("size", len(b)),
				log.String("source", path))

			if err := t.mfs.WriteFile(filePath, b, modePerm); err != nil {
				return err
			}
			t.files[filePath] = fi
		}
		return nil
	})
	if err != nil {
		t.T().Fatal(err)
	}

	var filePaths []string
	for p := range t.files {
		filePaths = append(filePaths, p)
	}
	t.filePaths = filePaths

	log.Info(fmt.Sprintf("[memfs_test:setup] file paths:\n%s", support.ToJSONFormatted(t.filePaths)))
}

func TestMemFSTestSuite(t *testing.T) {
	suite.Run(t, NewMemFSTestSuite())
}

func (t *MemFSTestSuite) TestFS() {
	assert.NoError(t.T(), fstest.TestFS(t.mfs, t.filePaths...))
}
