package helper

import "encoding/json"

// JSON is a json view halper
type JSON struct {
	view ViewInterface
}

// SetView sets view for halper
func (h *JSON) SetView(vi ViewInterface) error {
	h.view = vi
	return nil
}

// Encode sd
func (h *JSON) Encode(data interface{}, keepLayouts bool) ([]byte, error) {
	encoded, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	if !keepLayouts {
		//layout := Zend_Layout::getMvcInstance();
		//if ($layout instanceof Zend_Layout) {
		//	$layout->disableLayout();
		//}
	}

	return encoded, nil
}

// NewJSON creates a new JSON view halper
func NewJSON(vi ViewInterface) (Interface, error) {
	return &JSON{
		view: vi,
	}, nil
}
