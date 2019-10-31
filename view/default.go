package view

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"wsf/config"
	"wsf/controller/context"
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

	return true, nil
}

// PrepareTemplates parses a templates files
func (v *Default) PrepareTemplates() error {
	for _, tpl := range v.paths["template"] {
		err := utils.WalkDirectoryDeep(config.AppPath+tpl, config.AppPath+tpl, v.ReadTemplates)
		if err != nil {
			switch err.(type) {
			case *os.PathError:
				v.Logger.Warningf("[View] Unable to read template directory: %v", nil, err.Error())

			default:
				return err
			}
		}
	}

	return nil
}

// ReadTemplates loads and parses template into memory
func (v *Default) ReadTemplates(path string, info os.FileInfo, err error) error {
	if err != nil {
		return errors.Errorf("[View] Error scanning source: %s", err)
	}

	if info.IsDir() || strings.HasPrefix(info.Name(), ".") {
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

	tpl, err := template.New(path).Parse(string(tplData))
	if err != nil {
		return err
	}

	p, err := filepath.Rel(config.AppPath, path)
	if err != nil {
		return err
	}

	v.templates[p] = tpl
	v.template.New(p).Parse(string(tplData))
	return nil
}

// Render returns a render result of provided script
func (v *Default) Render(ctx context.Context, script string) ([]byte, error) {
	/* 	wr := &bytes.Buffer{}
	   	if err := v.template.ExecuteTemplate(wr, script, ctx.Data()); err == nil {
	   		b := make([]byte, wr.Len())
	   		_, err = wr.Read(b)
	   		if err != nil {
	   			return nil, err
	   		}

	   		return b, nil
	   		//fmt.Println(string(b))
	   		//os.Exit(2)
	   	} else {
	   		return nil, err
	   	}
	*/
	if t, ok := v.templates[script]; ok {
		wr := &bytes.Buffer{}
		err := t.Execute(wr, ctx.Data())
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

// GetTemplate sa
func (v *Default) GetTemplate(name string) *template.Template {
	return v.template.Lookup(name)
	//return v.templates[name]
}

// NewDefaultView creates new default view
func NewDefaultView(options *Config) (Interface, error) {
	v := &Default{}
	v.Options = options
	v.paths = make(map[string]map[string]string)
	v.params = make(map[string]interface{})
	v.helpers = make(map[string]helper.Interface)
	v.templates = make(map[string]*template.Template)
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
