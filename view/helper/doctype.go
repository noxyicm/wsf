package helper

import (
	"strings"

	"wsf/errors"
)

// DocType constants
const (
	XHTML11            = "XHTML11"
	XHTML1Strict       = "XHTML1_STRICT"
	XHTML1Transitional = "XHTML1_TRANSITIONAL"
	XHTML1Frameset     = "XHTML1_FRAMESET"
	XHTML1Rdfa         = "XHTML1_RDFA"
	XHTML1Rdfa11       = "XHTML1_RDFA11"
	XHTMLBasic1        = "XHTML_BASIC1"
	XHTML5             = "XHTML5"
	HTML4Strict        = "HTML4_STRICT"
	HTML4Loose         = "HTML4_LOOSE"
	HTML4Frameset      = "HTML4_FRAMESET"
	HTML5              = "HTML5"
	CUSTOMXHTML        = "CUSTOM_XHTML"
	CUSTOM             = "CUSTOM"
)

// Doctype is a DocType helper
type Doctype struct {
	registry map[string]interface{}
	view     ViewInterface
}

// SetView sets view
func (d *Doctype) SetView(vi ViewInterface) error {
	d.view = vi
	return nil
}

// SetDoctype sets a DocType
func (d *Doctype) SetDoctype(doctype string) error {
	switch doctype {
	case XHTML11, XHTML1Strict, XHTML1Transitional, XHTML1Frameset, XHTMLBasic1, XHTML1Rdfa, XHTML1Rdfa11, XHTML5, HTML4Strict, HTML4Loose, HTML4Frameset, HTML5:
		d.registry["doctype"] = doctype

	default:
		if doctype[0:9] != "<!DOCTYPE" {
			return errors.Errorf("The specified doctype is malformed")
		}

		typ := CUSTOM
		if strings.Contains(doctype, "xhtml") {
			typ = CUSTOMXHTML
		}

		d.registry["doctype"] = doctype
		d.registry["doctypes"].(map[string]string)[typ] = doctype
	}

	return nil
}

// GetDoctype returns DocType
func (d *Doctype) GetDoctype() string {
	return d.registry["doctypes"].(map[string]string)[d.registry["doctype"].(string)]
}

// IsXhtml returns true if doctype is XHTML
func (d *Doctype) IsXhtml() bool {
	return strings.Contains(d.registry["doctype"].(string), "xhtml")
}

// IsStrict returns true if doctype is strict html
func (d *Doctype) IsStrict() bool {
	switch d.registry["doctype"] {
	case XHTML1Strict, XHTML11, HTML4Strict:
		return true
	}

	return false
}

// IsHTML5 returns true if doctype is HTML5
func (d *Doctype) IsHTML5() bool {
	return strings.Contains(d.GetDoctype(), "<!DOCTYPE html>")
}

// IsRDFA returns true if doctype is RDFA
func (d *Doctype) IsRDFA() bool {
	return strings.Contains(d.registry["doctype"].(string), "rdfa")
}

// NewDoctype creates a new Doctype helper
func NewDoctype() (*Doctype, error) {
	doctypes := map[string]string{
		XHTML11:            "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.1//EN\" \"http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd\">",
		XHTML1Strict:       "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Strict//EN\" \"http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd\">",
		XHTML1Transitional: "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Transitional//EN\" \"http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd\">",
		XHTML1Frameset:     "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Frameset//EN\" \"http://www.w3.org/TR/xhtml1/DTD/xhtml1-frameset.dtd\">",
		XHTML1Rdfa:         "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML+RDFa 1.0//EN\" \"http://www.w3.org/MarkUp/DTD/xhtml-rdfa-1.dtd\">",
		XHTML1Rdfa11:       "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML+RDFa 1.1//EN\" \"http://www.w3.org/MarkUp/DTD/xhtml-rdfa-2.dtd\">",
		XHTMLBasic1:        "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML Basic 1.0//EN\" \"http://www.w3.org/TR/xhtml-basic/xhtml-basic10.dtd\">",
		XHTML5:             "<!DOCTYPE html>",
		HTML4Strict:        "<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.01//EN\" \"http://www.w3.org/TR/html4/strict.dtd\">",
		HTML4Loose:         "<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.01 Transitional//EN\" \"http://www.w3.org/TR/html4/loose.dtd\">",
		HTML4Frameset:      "<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.01 Frameset//EN\" \"http://www.w3.org/TR/html4/frameset.dtd\">",
		HTML5:              "<!DOCTYPE html>",
	}

	return &Doctype{
		registry: map[string]interface{}{"doctypes": doctypes, "doctype": HTML5},
	}, nil
}
