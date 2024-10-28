package translate

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/file/reader"
	"github.com/noxyicm/wsf/locale"
	"github.com/noxyicm/wsf/utils"
)

const (
	// TYPEAdapterCSV resource type
	TYPEAdapterCSV = "csv"
)

func init() {
	RegisterAdapter(TYPEAdapterCSV, NewCSVAdapter)
}

// CSVAdapter is a translate adapter fro csv files
type CSVAdapter struct {
	*DefaultAdapter

	Length    int
	Delimiter rune
	Enclosure rune
	Comment   rune
	temp      map[string]*Entry
}

// AddTranslationData adds new translations
// This may be a new language or additional content for an existing language
// If the key 'clear' is true, then translations for the specified
// language will be replaced and added otherwise
func (a *CSVAdapter) AddTranslationData(data map[string]*Entry, loc string) error {
	var err error
	loc, err = locale.FindLocale(loc)
	if err != nil {
		return errors.Wrapf(err, "The given Language '%s' does not exist", loc)
	}

	if _, ok := a.translate[loc]; !ok || a.Options.Clear {
		a.translate[loc] = data
	} else {
		for key, value := range data {
			a.translate[loc][key] = value
		}
	}

	return nil
}

// LoadTranslationData loads translation data into memory
func (a *CSVAdapter) LoadTranslationData(loc string, options map[string]interface{}) (map[string]*Entry, error) {
	a.temp = make(map[string]*Entry)
	if a.Options.Content == nil {
		return a.temp, errors.New("Required option 'content' is missing")
	}

	read := true
	if a.Options.UseCache {
		hash := md5.Sum([]byte(loc))
		id := "WSF_Translate_" + hex.EncodeToString(hash[:]) + "_CSV"
		if cached, ok := a.Cache.Load(id, false); ok {
			if err := json.Unmarshal(cached, &a.temp); err != nil {
				read = false
			}
		}
	}

	if a.Options.Reload {
		read = true
	}

	if read {
		switch content := a.Options.Content.(type) {
		case string:
			err := utils.WalkDirectoryDeep(filepath.Join(config.AppPath, filepath.FromSlash(content)), filepath.Join(config.AppPath, filepath.FromSlash(content)), a.traverseFile)
			if err != nil {
				switch err.(type) {
				case *os.PathError:
					return a.temp, errors.Wrap(err, "Unable to read translation file")

				default:
					return a.temp, errors.Wrap(err, "Unable to read translation file")
				}
			}
			break

		case map[string][]string:
			for key, values := range content {
				switch len(values) {
				case 1:
					a.temp[key] = &Entry{
						Single:   values[0],
						Double:   "",
						Plural:   "",
						Multiple: false,
					}
					break

				case 2:
					a.temp[key] = &Entry{
						Single:   values[0],
						Double:   "",
						Plural:   values[1],
						Multiple: true,
					}
					break

				case 3:
					a.temp[key] = &Entry{
						Single:   values[0],
						Double:   values[1],
						Plural:   values[2],
						Multiple: true,
					}
					break

				default:
					a.temp[key] = &Entry{
						Single:   values[0],
						Double:   "",
						Plural:   "",
						Multiple: false,
					}
				}
			}
			break

		case map[string]interface{}:
			for key, vals := range content {
				switch values := vals.(type) {
				case string:
					a.temp[key] = &Entry{
						Single:   values,
						Double:   "",
						Plural:   "",
						Multiple: false,
					}
					break

				case []string:
					switch len(values) {
					case 1:
						a.temp[key] = &Entry{
							Single:   values[0],
							Double:   "",
							Plural:   "",
							Multiple: false,
						}
						break

					case 2:
						a.temp[key] = &Entry{
							Single:   values[0],
							Double:   "",
							Plural:   values[1],
							Multiple: true,
						}
						break

					case 3:
						a.temp[key] = &Entry{
							Single:   values[0],
							Double:   values[1],
							Plural:   values[2],
							Multiple: true,
						}
						break

					default:
						a.temp[key] = &Entry{
							Single:   values[0],
							Double:   "",
							Plural:   "",
							Multiple: false,
						}
					}
					break
				}
			}
			break
		}
	}

	return a.temp, nil
}

func (a *CSVAdapter) traverseFile(path string, info fs.FileInfo, err error) error {
	filename := info.Name()
	for key, ignore := range a.Options.Ignore {
		if strings.Contains(key, "regex") {
			rg := regexp.MustCompile(ignore)
			if rg.MatchString(filename) {
				continue
			}
		} else if strings.Contains(path, filepath.FromSlash(ignore)) {
			// ignore files matching first characters from option 'ignore' and all files below
			continue
		}
	}

	loc := ""
	prev := ""
	if info.IsDir() {
		// pathname as locale
		if a.Options.Scan == a.LocaleDirectory && locale.IsLocale(filename, true) {
			loc = filename
			prev = filename
		}
	} else {
		// filename as locale
		if a.Options.Scan == a.LocaleFilename {
			filename = filename[:len(filename)-len(filepath.Ext(filename))]
			if locale.IsLocale(filename, true) {
				loc = filename
			} else {
				parts := strings.Split(filename, ".")
				parts2 := make([]string, 0)
				for _, token := range parts {
					parts2 = append(parts2, strings.Split(token, "_")...)
				}
				parts = append(parts, parts2...)
				parts2 = make([]string, 0)
				for _, token := range parts {
					parts2 = append(parts2, strings.Split(token, "-")...)
				}
				parts = append(parts, parts2...)
				parts = utils.UniqueSSlice(parts)
				for _, token := range parts {
					if locale.IsLocale(token, true) {
						if len(prev) <= len(token) {
							loc = token
							prev = token
						}
					}
				}
			}
		}
	}

	fmt.Printf("loc : %v\n", loc)
	if loc != "" {
		data, err := a.readFile(path)
		if err != nil {
			return errors.Wrapf(err, "Unable to read file '%s'", path)
		}

		for key, value := range data {
			a.temp[key] = value
		}
	}

	return nil
}

func (a *CSVAdapter) readFile(filename string) (map[string]*Entry, error) {
	fd, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "[CSVAdapter] Unable to open file '%s'", filename)
	}

	m := make(map[string]*Entry)
	r := reader.NewCSVReader(fd)
	r.Comma = a.Delimiter
	r.Comment = a.Comment

	row := 0
	columns := make([]string, 0)

	record, err := r.Read()
	if err == io.EOF {
		fd.Close()
		return m, nil
	} else if err != nil {
		fd.Close()
		return nil, errors.Wrapf(err, "[CSVAdapter] Unable to read record")
	}

	for _, col := range record {
		columns = append(columns, col)
	}
	parts := len(columns)

ParseLoop:
	for {
		record, err = r.Read()
		if err == io.EOF {
			break ParseLoop
		} else if err != nil {
			continue
		}

		switch parts {
		case 2:
			m[record[0]] = &Entry{
				Single:   record[1],
				Double:   "",
				Plural:   "",
				Multiple: false,
			}
			break

		case 3:
			m[record[0]] = &Entry{
				Single:   record[1],
				Double:   "",
				Plural:   record[2],
				Multiple: true,
			}
			break

		case 4:
			m[record[0]] = &Entry{
				Single:   record[1],
				Double:   record[2],
				Plural:   record[3],
				Multiple: true,
			}
			break

		default:
			m[record[0]] = &Entry{
				Single:   record[1],
				Double:   "",
				Plural:   "",
				Multiple: false,
			}
		}

		row++
	}

	fd.Close()
	return m, nil
}

// NewCSVAdapter creates a new translate adapter of type CSVAdapter
func NewCSVAdapter(cfg *AdapterConfig) (Adapter, error) {
	a := &CSVAdapter{
		Length:    0,
		Delimiter: ';',
		Enclosure: '"',
		Comment:   '#',
	}

	var err error
	a.DefaultAdapter, err = NewDefaultAdapter(cfg)
	if err == nil {
		return nil, errors.Wrap(err, "Unable to create translate adapter")
	}

	if v, ok := cfg.Params["length"]; ok {
		a.Length = int(v.(float64))
	}

	if v, ok := cfg.Params["delimiter"]; ok {
		a.Delimiter = []rune(v.(string))[0]
	}

	if v, ok := cfg.Params["enclosure"]; ok {
		a.Enclosure = []rune(v.(string))[0]
	}

	if v, ok := cfg.Params["comment"]; ok {
		a.Comment = []rune(v.(string))[0]
	}

	return a, nil
}
