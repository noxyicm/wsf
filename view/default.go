package view

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"github.com/noxyicm/wsf/config"
	"github.com/noxyicm/wsf/context"
	"github.com/noxyicm/wsf/errors"
	"github.com/noxyicm/wsf/log"
	"github.com/noxyicm/wsf/registry"
	"github.com/noxyicm/wsf/utils"
	"github.com/noxyicm/wsf/view/helper"
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

	doctype  *helper.Doctype
	headMeta *helper.HeadMeta
	funcMap  map[string]interface{}
}

// Init a view resource
func (v *Default) Init(options *Config) (bool, error) {
	v.Options = options
	v.BaseDir = options.BaseDir
	v.ViewBasePathSpec = options.ViewBasePathSpec
	v.ViewScriptPathSpec = options.ViewScriptPathSpec
	v.ViewScriptPathNoControllerSpec = options.ViewScriptPathNoControllerSpec
	v.ViewSuffix = options.ViewSuffix

	logResource := registry.GetResource("syslog")
	if logResource == nil {
		return false, errors.New("[View] Log resource is not configured")
	}
	v.Logger = logResource.(*log.Log)

	if options.Doctype != "" {
		v.doctype.SetDoctype(strings.ToUpper(options.Doctype))
		if options.Charset != "" && v.doctype.IsHTML5() {
			v.headMeta.SetCharset(options.Charset)
		}
	}

	if options.ContentType != "" {
		v.headMeta.AppendHTTPEquiv("Content-Type", options.ContentType, map[string]string{})
	}

	if len(options.Assign) > 0 {
		v.Assign(options.Assign)
	}

	return true, nil
}

// Setup resource
func (v *Default) Setup() (bool, error) {
	err := v.PrepareTemplates()
	if err != nil {
		return false, err
	}

	err = v.PrepareLayouts()
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

	_, filename := filepath.Split(path)
	tplName := strings.Replace(filename, filepath.Ext(filename), "", -1)
	for _, tpl := range v.templates {
		_, err = tpl.New(tplName).Parse(string(tplRaw))
		if err != nil {
			return err
		}
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

	v.templates[relPath], err = template.New(v.Options.LayoutContentKey).Funcs(template.FuncMap(v.funcMap)).Parse(string(tplRaw))
	if err != nil {
		return err
	}

	return nil
}

// GetOptions returns view options
func (v *Default) GetOptions() *Config {
	return v.Options
}

// Render returns a render result of provided script
func (v *Default) Render(ctx context.Context, script string, tpl string) ([]byte, error) {
	/*wr := &bytes.Buffer{}
	if err := v.templates[script].ExecuteTemplate(wr, tpl, ctx.Data()); err == nil {
		b := make([]byte, wr.Len())
		_, err = wr.Read(b)
		if err != nil {
			return nil, err
		}

		//return b, nil
		fmt.Println(string(b))
		os.Exit(2)
	} else {
		return nil, err
	}*/

	if t, ok := v.templates[script]; ok {
		wr := &bytes.Buffer{}
		err := t.ExecuteTemplate(wr, tpl, ctx.Data())
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
	v.helpers = make(map[string]helper.Interface)
	v.templates = make(map[string]*template.Template)
	v.layouts = make(map[string]*TemplateData)
	v.template = template.New("layout")

	// for default view doctype view halper is mandatory
	dctp, err := helper.NewDoctype()
	if err != nil {
		return nil, err
	}

	v.doctype = dctp
	v.RegisterHelper("Doctype", dctp)

	// for default view headmeta view halper is mandatory
	hdmt, err := helper.NewHeadMeta()
	if err != nil {
		return nil, err
	}

	v.headMeta = hdmt
	v.RegisterHelper("HeadMeta", hdmt)

	return v, nil
}
