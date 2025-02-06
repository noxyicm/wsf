package view

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/filter"
	"github.com/noxyicm/wsf/filter/word"
	"github.com/noxyicm/wsf/log"
	"github.com/noxyicm/wsf/registry"
	"github.com/noxyicm/wsf/utils"
)

const (
	// TYPEDfault is a type of view
	TYPEDfault = "default"
)

func init() {
	Register(TYPEDfault, NewDefaultView)
}

// Default is a default view
type Default struct {
	view

	funcMap map[string]interface{}
}

// Init a view resource
func (v *Default) Init(options *Config) (bool, error) {
	v.Options = options
	v.BaseDir = options.BaseDir
	v.ViewBasePathSpec = options.ViewBasePathSpec
	v.ViewActionPathSpec = options.ViewActionPathSpec
	v.ViewActionPathNoControllerSpec = options.ViewActionPathNoControllerSpec
	v.ViewHelperPathSpec = options.ViewHelperPathSpec
	v.ViewSuffix = options.ViewSuffix

	logResource := registry.GetResource("syslog")
	if logResource == nil {
		return false, errors.New("[View] Log resource is not configured")
	}
	v.Logger = logResource.(*log.Log)

	if len(options.Assign) > 0 {
		v.Assign(options.Assign)
	}

	for hlprType := range buildViewHelperHandlers {
		hlpr, err := NewHelper(hlprType)
		if err != nil {
			return false, errors.Wrapf(err, "Unable to instantiate view helper of type '%s'", hlprType)
		}

		if err := v.RegisterHelper(hlpr.Name(), hlpr); err != nil {
			return false, errors.Wrapf(err, "Unable to register view helper '%s'", hlpr.Name())
		}

		if err := hlpr.Init(v, nil); err != nil {
			return false, errors.Wrapf(err, "Unable to initialize view helper '%s'", hlpr.Name())
		}
	}

	if options.Doctype != "" {
		v.Helper("Doctype").(*Doctype).SetDoctype(strings.ToUpper(options.Doctype))
		if options.Charset != "" && v.Helper("Doctype").(*Doctype).IsHTML5() {
			v.Helper("HeadMeta").(*HeadMeta).SetCharset(options.Charset)
		}
	}

	if options.ContentType != "" {
		v.Helper("HeadMeta").(*HeadMeta).AppendHTTPEquiv("Content-Type", options.ContentType, map[string]string{})
	}

	helpersPath, err := v.PathFromSpec(v.ViewHelperPathSpec, map[string]string{})
	if err != nil {
		return false, errors.Wrap(err, "Unable to parse helper path")
	}

	prefix, err := v.PrefixFromPath(helpersPath)
	if err != nil {
		return false, errors.Wrap(err, "Unable to parse helper path")
	}

	if err := v.AddHelperPath(helpersPath, prefix); err != nil {
		return false, errors.Wrap(err, "Unable to add view helper path")
	}

	return true, nil
}

// Setup resource
func (v *Default) Setup() (bool, error) {
	if err := v.AddTemplateFunc("htmlAttr", func(a string) template.HTMLAttr { return template.HTMLAttr(a) }); err != nil {
		return false, errors.Wrap(err, "Unable to add function to template")
	}

	if err := v.AddTemplateFunc("htmlText", func(a string) template.HTML { return template.HTML(a) }); err != nil {
		return false, errors.Wrap(err, "Unable to add function to template")
	}

	if err := v.AddTemplateFunc("htmlURL", func(a string) template.URL { return template.URL(a) }); err != nil {
		return false, errors.Wrap(err, "Unable to add function to template")
	}

	if err := v.AddTemplateFunc("htmlJS", func(a string) template.JS { return template.JS(a) }); err != nil {
		return false, errors.Wrap(err, "Unable to add function to template")
	}

	if err := v.AddTemplateFunc("htmlCSS", func(a string) template.CSS { return template.CSS(a) }); err != nil {
		return false, errors.Wrap(err, "Unable to add function to template")
	}

	if err := v.AddTemplateFunc("presentError", func(err error, format int) string { return err.Error() }); err != nil {
		return false, errors.Wrap(err, "Unable to add function to template")
	}

	if err := v.AddTemplateFunc("toString", func(a interface{}) string {
		switch parsed := a.(type) {
		case bool:
			return strconv.FormatBool(parsed)

		case int:
		case int8:
		case int16:
		case int32:
		case int64:
		case uint:
		case uint8:
		case uint16:
		case uint32:
		case uint64:
			return strconv.Itoa(int(parsed))

		case float32:
		case float64:
			return strconv.FormatFloat(float64(parsed), 'b', 64, 64)
		}

		return ""
	}); err != nil {
		return false, errors.Wrap(err, "Unable to add function to template")
	}

	for _, hlpr := range v.helpers {
		if err := hlpr.Setup(); err != nil {
			return false, errors.Wrapf(err, "Unable to setup view helper '%s'", hlpr.Name())
		}
	}

	err := v.PrepareTemplates()
	if err != nil {
		return false, err
	}

	// err = v.PrepareLayouts()
	// if err != nil {
	// 	return false, err
	// }

	err = v.PrepareHelpers()
	if err != nil {
		return false, err
	}

	return true, nil
}

// PrepareLayouts parses a layout templates files
func (v *Default) PrepareLayouts() error {
	for _, path := range v.paths["layouts"] {
		err := utils.WalkDirectoryDeep(filepath.Join(config.AppPath, filepath.FromSlash(path)), filepath.Join(config.AppPath, filepath.FromSlash(path)), v.ReadLayouts)
		if err != nil {
			switch err.(type) {
			case *os.PathError:
				v.Logger.Warningf("[View] Unable to read layout directory: %v", nil, err.Error())

			default:
				return err
			}
		}
	}

	return nil
}

// ReadLayouts loads and parses layout template into memory, extending
// all existing templates
func (v *Default) ReadLayouts(path string, info os.FileInfo, err error) error {
	if err != nil {
		return errors.Errorf("Scanning source '%s' failed: %v", path, err)
	}

	if info.IsDir() || strings.HasPrefix(info.Name(), ".") {
		return nil
	}

	tplFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer tplFile.Close()

	tplRaw := make([]byte, info.Size())
	_, err = tplFile.Read(tplRaw)
	if err != nil {
		return err
	}

	// _, filename := filepath.Split(path)
	// tplName := strings.Replace(filename, filepath.Ext(filename), "", -1)
	// for _, tpl := range v.templates {
	// 	_, err = tpl.New(tplName).Parse(string(tplRaw))
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	relPath, err := filepath.Rel(config.AppPath, path)
	if err != nil {
		return err
	}

	v.templates[relPath], err = template.New(v.Options.SegmentContentKey).Funcs(template.FuncMap(v.funcMap)).Parse(string(tplRaw))
	if err != nil {
		return err
	}

	return nil
}

// PrepareTemplates parses a templates files
func (v *Default) PrepareTemplates() error {
	for _, path := range v.paths["templates"] {
		err := utils.WalkDirectoryDeep(filepath.Join(config.AppPath, filepath.FromSlash(path)), filepath.Join(config.AppPath, filepath.FromSlash(path)), v.ReadTemplates)
		if err != nil {
			switch err.(type) {
			case *os.PathError:
				v.Logger.Warning(errors.Wrap(err, "[View] Unable to read template directory"), nil)

			default:
				return errors.Wrap(err, "[View] Unable to read template directory")
			}
		}
	}

	return nil
}

// ReadTemplates loads and parses template into memory
func (v *Default) ReadTemplates(path string, info os.FileInfo, err error) error {
	if err != nil {
		return errors.Errorf("Scanning source '%s' failed: %v", path, err)
	}

	if info.IsDir() || strings.HasPrefix(info.Name(), ".") {
		return nil
	}

	tplFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer tplFile.Close()

	tplRaw := make([]byte, info.Size())
	_, err = tplFile.Read(tplRaw)
	if err != nil {
		return err
	}

	relPath, err := filepath.Rel(config.AppPath, path)
	if err != nil {
		return err
	}

	v.templates[relPath], err = template.New(v.Options.SegmentContentKey).Funcs(template.FuncMap(v.funcMap)).Parse(string(tplRaw))
	if err != nil {
		return err
	}

	return nil
}

// PrepareHelpers parses a templates files
func (v *Default) PrepareHelpers() error {
	for helperPath, _ := range v.paths["helpers"] {
		err := utils.WalkDirectoryDeep(filepath.Join(config.AppPath, filepath.FromSlash(helperPath)), filepath.Join(config.AppPath, filepath.FromSlash(helperPath)), v.ReadHelpers)
		if err != nil {
			switch err.(type) {
			case *os.PathError:
				v.Logger.Warning(errors.Wrap(err, "[View] Unable to read template directory"), nil)

			default:
				return errors.Wrap(err, "[View] Unable to read template directory")
			}
		}
	}

	return nil
}

// ReadHelpers loads and parses template into memory
func (v *Default) ReadHelpers(path string, info os.FileInfo, err error) error {
	if err != nil {
		return errors.Errorf("Scanning source '%s' failed: %v", path, err)
	}

	if info.IsDir() || strings.HasPrefix(info.Name(), ".") {
		return nil
	}

	tplFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer tplFile.Close()

	tplRaw := make([]byte, info.Size())
	_, err = tplFile.Read(tplRaw)
	if err != nil {
		return err
	}

	relPath, err := filepath.Rel(config.AppPath, path)
	if err != nil {
		return err
	}

	//_, fileName := filepath.Split(tplFile.Name())
	//tplName := strings.Replace(fileName, filepath.Ext(fileName), "", -1)
	prefix, err := v.PrefixFromPath(strings.Replace(relPath, "."+v.ViewSuffix, "", 1))
	if err != nil {
		return errors.Wrap(err, "Unable to parse helper path")
	}

	//funcMap := utils.MapSCopy(v.funcMap)
	//delete(funcMap, )
	v.templates[prefix], err = template.New(v.Options.SegmentContentKey).Funcs(template.FuncMap(v.funcMap)).Parse(string(tplRaw))
	if err != nil {
		return err
	}

	return nil
}

// GetOptions returns view options
func (v *Default) GetOptions() *Config {
	return v.Options
}

// PathFromSpec transforms spec into path
func (v *Default) PathFromSpec(spec string, params map[string]string) (string, error) {
	inflector, err := filter.NewInflector()
	if err != nil {
		return "", errors.Wrap(err, "Unable to determine path from spec")
	}

	uts, err := word.NewUnderscoreToSeparator("/")
	if err != nil {
		return "", errors.Wrap(err, "Unable to determine path from spec")
	}

	rrc, err := filter.NewRegexpReplace(`\.`, "-")
	if err != nil {
		return "", errors.Wrap(err, "Unable to determine path from spec")
	}

	inflector.AddRules(map[string]interface{}{
		":module":     []interface{}{"Word_CamelCaseToDash", "StringToLower"},
		":controller": []interface{}{"Word_CamelCaseToDash", uts, "StringToLower", rrc},
	})
	inflector.SetTarget(spec)

	path, err := inflector.Filter(params)
	if err != nil {
		return "", errors.Wrap(err, "Unable to determine path from spec")
	}

	return path.(string), nil
}

// PrefixFromPath transforms path into prefix
func (v *Default) PrefixFromPath(path string) (string, error) {
	stu, err := word.NewSeparatorToUnderscore("/")
	if err != nil {
		return "", errors.Wrap(err, "Unable to determine prefix from path")
	}

	prefix, err := stu.Filter(path)
	if err != nil {
		return "", errors.Wrap(err, "Unable to determine prefix from path")
	}

	return prefix.(string), nil
}

// Render returns a render result of provided script
func (v *Default) Render(data map[string]interface{}, script string, tpl string) ([]byte, error) {
	if t, ok := v.templates[script]; ok {
		wr := &bytes.Buffer{}
		err := t.ExecuteTemplate(wr, tpl, data)
		if err != nil {
			return nil, err
		}

		b := make([]byte, wr.Len())
		_, err = wr.Read(b)
		if err != nil {
			return nil, err
		}

		return b, nil
	}

	return nil, errors.Errorf("[View] Template by name '%s' not found", script)
}

// AddTemplateFunc add a function to template
func (v *Default) AddTemplateFunc(name string, fn interface{}) error {
	if _, ok := v.funcMap[name]; ok {
		return errors.Errorf("Template function by name '%s' is already registered", name)
	}

	v.funcMap[name] = fn
	return nil
}

// SetTemplateFunc adds or replaces existsing function in template
func (v *Default) SetTemplateFunc(name string, fn interface{}) error {
	v.funcMap[name] = fn
	return nil
}

// RemoveTemplateFunc removes function from template
func (v *Default) RemoveTemplateFunc(name string) error {
	if _, ok := v.funcMap[name]; !ok {
		return errors.Errorf("Template function by name '%s' is not registered", name)
	}

	delete(v.funcMap, name)
	return nil
}

// TemplateFunctions returns map of registered template functions
func (v *Default) TemplateFunctions() map[string]interface{} {
	return v.funcMap
}

// GetTemplate sa
func (v *Default) GetTemplate(path string) *template.Template {
	//return v.template.Lookup(name)
	if v, ok := v.templates[path]; ok {
		return v
	}

	return nil
}

// NewDefaultView creates new default view
func NewDefaultView(options *Config) (Interface, error) {
	v := &Default{
		funcMap: make(map[string]interface{}),
	}
	v.Options = options
	v.paths = make(map[string]map[string]string)
	v.params = make(map[string]interface{})
	v.helpers = make(map[string]HelperInterface)
	v.templates = make(map[string]*template.Template)
	v.layouts = make(map[string]*template.Template)
	v.template = template.New("layout")

	return v, nil
}
