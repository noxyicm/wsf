package dispatcher

import (
	"reflect"
	"wsf/controller/action"
	"wsf/controller/action/helper"
	"wsf/controller/context"
	"wsf/controller/request"
	"wsf/controller/response"
	"wsf/errors"
	"wsf/log"
	"wsf/registry"
)

const (
	// TYPEDefault represents default dispatcher
	TYPEDefault = "default"
)

func init() {
	Register(TYPEDefault, NewDefaultDispatcher)
}

// Default is a default dispatcher
type Default struct {
	standart
}

// Dispatch dispatches the request into the apropriet handler
func (d *Default) Dispatch(ctx context.Context, rqs request.Interface, rsp response.Interface) (bool, error) {
	md := d.modules.Module(rqs.ModuleName())
	if md == nil {
		if md = d.modules.Module(d.DefaultModule()); md == nil {
			return true, errors.Errorf("Invalid module specified ( %s )", rqs.ModuleName())
		}
	}

	ctrlType, err := md.ControllerType(rqs.ControllerName())
	if err != nil {
		if ctrlType, err = md.ControllerType(d.DefaultController()); err != nil {
			return true, err
		}
	}

	ctrl := reflect.New(ctrlType).Interface()
	if !reflect.TypeOf(ctrl).Implements(reflect.TypeOf((*action.Interface)(nil)).Elem()) {
		return true, errors.Errorf("Controller ( %s ) does not implements action.Controller", rqs.ControllerName())
	}

	d.PopulateController(ctx, ctrl, rqs, rsp, d.invokeParams)
	actctrl := ctrl.(action.Interface)
	err = actctrl.NewHelperBroker()
	if err != nil {
		return true, err
	}

	act, _ := d.GetActionMethod(rqs)
	mtd, ok := reflect.TypeOf(ctrl).MethodByName(act)
	if !ok {
		return true, errors.Errorf("Action ( %s ) does not exists", act)
	}

	if !d.ParamBool("noViewRenderer") && !actctrl.HelperBroker().HasHelper("viewRenderer") {
		vr, err := helper.NewViewRenderer()
		if err != nil {
			return true, err
		}

		vr.SetController(ctrl.(helper.ControllerInterface))
		err = actctrl.HelperBroker().SetHelper(-80, vr, nil)
		if err != nil {
			return true, err
		}
	}

	// Initiate action controller
	rqs.SetDispatched(true)
	if err = actctrl.Init(); err != nil {
		return true, err
	}

	err = actctrl.Dispatch(actctrl, mtd)
	if err != nil {
		return true, err
	}

	return true, nil
}

// NewDefaultDispatcher creates new default dispatcher
func NewDefaultDispatcher(options *Config) (di Interface, err error) {
	d := &Default{}
	d.options = options

	logResource := registry.GetResource("syslog")
	if logResource == nil {
		return nil, errors.New("[Dispatcher] Log resource is not configured")
	}

	d.logger = logResource.(*log.Log)

	return d, nil
}
