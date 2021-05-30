package jsonschema

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/refto/server/service/data"
	"github.com/xeipuuv/gojsonschema"
)

type Schema struct {
	// I need Path only to display informative error messages
	Path   string
	Schema *gojsonschema.Schema
}

type File struct {
	// I need Path only to display informative error messages
	Path string
	Data []byte
}

type schemas map[data.Type]Schema
type filesByType map[data.Type][]File

type ValidateResult struct {
	SchemaCount     int
	DataCount       int
	DataCountByType map[data.Type]int
}

type ErrStack []string

func (e *ErrStack) Add(err error, wrapOpt ...string) {
	if len(wrapOpt) > 0 {
		err = errors.New(wrapOpt[0] + ": " + err.Error())
	}

	*e = append(*e, err.Error())
}
func (e ErrStack) Error() string {
	return strings.Join(e, "\n")
}

// Validate validates each YAML doc against JSON schema of it's type
func Validate(dirPath string) (resp ValidateResult, err error) {
	_, err = os.Stat(dirPath)
	if err != nil {
		return
	}

	// add trailing slash, so base name of file that not have it
	// when displaying errors and messages
	if !strings.HasSuffix(dirPath, string(filepath.Separator)) {
		dirPath += string(filepath.Separator)
	}

	resp = ValidateResult{
		SchemaCount:     0,
		DataCount:       0,
		DataCountByType: map[data.Type]int{},
	}

	schemasRepo := schemas{}
	dataRepo := filesByType{}
	errs := ErrStack{}

	// collect schemas along with data (to not walk dirs seconds time)
	err = filepath.Walk(dirPath, func(path string, f os.FileInfo, wErr error) (err error) {
		if wErr != nil {
			return wErr
		}

		if data.IsSchemaFile(f.Name()) {
			err = registerSchema(dirPath, path, schemasRepo)
			if err != nil {
				errs.Add(err, "register schema: "+relPath(dirPath, path))
				return nil
			}
			resp.SchemaCount++
			return nil
		}

		// skip dirs, samples and non-yaml files
		if !data.IsDataFile(f) {
			return
		}

		resp.DataCount++
		jsonBytes, err := data.JSONBytesFromYAMLFile(path)
		if err != nil {
			errs.Add(err, relPath(dirPath, path))
			return nil
		}

		t := data.TypeFromFilename(path)
		_, ok := resp.DataCountByType[t]
		if !ok {
			resp.DataCountByType[t] = 0
		}
		resp.DataCountByType[t]++
		files, ok := dataRepo[t]
		if !ok {
			files = []File{}
		}
		files = append(files, File{
			Path: path,
			Data: jsonBytes,
		})
		dataRepo[t] = files

		return nil
	})

	if len(errs) > 0 {
		err = errs
		return
	}

	// iterate over collected data and validate each
	for t, files := range dataRepo {
		for _, v := range files {
			schema, ok := schemasRepo[t]
			if !ok {
				errs.Add(fmt.Errorf("schema of type '%s' is not exists (source '%s')", t, relPath(dirPath, v.Path)))
				break
			}

			loader := gojsonschema.NewBytesLoader(v.Data)
			result, err := schema.Schema.Validate(loader)
			if err != nil {
				errs.Add(err, "validate "+relPath(dirPath, v.Path))
				continue
			}

			if !result.Valid() {
				errs.Add(errors.New("validation failed for " + relPath(dirPath, v.Path)))
				for _, e := range result.Errors() {
					errs.Add(errors.New("\t - " + e.String()))
				}
			}
		}
	}

	if len(errs) > 0 {
		err = errs
	}

	return
}

// registerSchema loads YAML schema from fPath converts it to JSON and adds it to the repo
// returns error if schema of given type already been registered
func registerSchema(dirPath, filePath string, repo schemas) (err error) {
	t := data.TypeFromSchemaFilename(filePath)

	// check if this schema type already registered
	existing, ok := repo[t]
	if ok {
		err = fmt.Errorf(
			"schema type '%s' from '%s' already registered in '%s'",
			t, relPath(dirPath, filePath), relPath(dirPath, existing.Path),
		)
		return
	}

	jsonBytes, err := data.JSONBytesFromYAMLFile(filePath)
	if err != nil {
		return
	}

	sl := gojsonschema.NewSchemaLoader()
	loader := gojsonschema.NewBytesLoader(jsonBytes)
	schema, err := sl.Compile(loader)
	if err != nil {
		return
	}

	repo[t] = Schema{
		Path:   filePath,
		Schema: schema,
	}

	return
}

func relPath(dirPath, filePath string) string {
	return strings.TrimPrefix(filePath, dirPath)
}
