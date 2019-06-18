package view

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"wsf/config"
	"wsf/errors"
	"wsf/log"
	"wsf/registry"
	"wsf/utils"
	"wsf/view/helper"
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
}

// Init a view resource
func (v *Default) Init(options *Config) (bool, error) {
	v.options = options
	v.baseDir = options.BaseDir

	logResource := registry.Get("log")
	if logResource == nil {
		return false, errors.New("[View] Log resource is not configured")
	}

	v.logger = logResource.(*log.Log)

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

	err := v.PrepareTemplates()
	if err != nil {
		return false, err
	}

	return true, nil
}

// PrepareTemplates parses a templates files
func (v *Default) PrepareTemplates() error {
	for _, tpl := range v.paths["template"] {
		err := utils.WalkDirectoryDeep(config.AppPath+tpl, config.AppPath+tpl, v.ReadTemplates)
		if err != nil {
			switch err.(type) {
			case *os.PathError:
				v.logger.Warningf("[View] Unable to read template directory: %v", nil, err.Error())

			default:
				return err
			}
		}
	}

	return nil
}

// ReadTemplates as
func (v *Default) ReadTemplates(path string, info os.FileInfo, err error) error {
	if err != nil {
		return errors.Errorf("[View] Error scanning source: %s", err)
	}

	if info.IsDir() || strings.HasPrefix(info.Name(), ".") || !strings.HasSuffix(info.Name(), ".phtml") {
		return nil
	}

	tplFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer tplFile.Close()

	tplData := make([]byte, info.Size())
	_, err = tplFile.Read(tplData)
	if err != nil {
		return err
	}

	tpl, err := template.New(info.Name()).Parse(string(tplData))
	if err != nil {
		return err
	}

	p, _ := filepath.Rel(config.AppPath, path)
	if err != nil {
		return err
	}

	//suffix := filepath.Ext(p)
	//p = p[0 : len(p)-len(suffix)]
	v.templates[p] = tpl
	return nil
}

// Render returns a render result of provided script
func (v *Default) Render(script string) ([]byte, error) {
	if v, ok := v.templates[script]; ok {
		wr := &bytes.Buffer{}
		err := v.Execute(wr, map[string]interface{}{})
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

// NewDefaultView creates new default view
func NewDefaultView(options *Config) (Interface, error) {
	v := &Default{}
	v.options = options
	v.baseDir = options.BaseDir
	v.paths = make(map[string]map[string]string)
	v.params = make(map[string]interface{})
	v.helpers = make(map[string]helper.Interface)
	v.templates = make(map[string]*template.Template)

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
