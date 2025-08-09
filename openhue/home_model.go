package openhue

import (
	"openhue-cli/util/color"

	"github.com/openhue/openhue-go"
	"github.com/spf13/cobra"
)

type HomeResourceType openhue.ResourceIdentifierRtype

type Resource struct {
	Id     string
	Name   string
	Type   HomeResourceType
	Parent *Resource
	// internal
	ctx *hueHomeCtx
}

// matchesNameOrId returns true of the given parameter equals either the Resource Name or Id.
// If the parameter is empty, true is returned.
func (r *Resource) matchesNameOrId(nameOrId string) bool {
	if len(nameOrId) == 0 {
		return true
	}
	return r.Name == nameOrId || r.Id == nameOrId
}

//
// HomeModel
//

type HomeModel struct {
	Resource
	Rooms   []Room
	Devices []Device
	HueData *openhue.BridgeHomeGet
}

//
// Room
//

type Room struct {
	Resource
	Devices []Device
	Scenes  []Scene
	HueData *openhue.RoomGet

	// Services
	GroupedLight *GroupedLight
}

//
// Device
//

type Device struct {
	Resource
	HueData *openhue.DeviceGet

	// Services
	Light *Light
}

//
// Light
//

type SetLightOptions struct {
	Status         LightStatus
	Brightness     float32
	Color          color.XY
	Temperature    int
	TransitionTime int
}

func NewSetLightOptions() *SetLightOptions {
	return &SetLightOptions{
		Status:      LightStatusUndefined,
		Brightness:  -1,
		Color:       color.UndefinedColor,
		Temperature: -1,
	}
}

type LightStatus string

const (
	LightStatusOn        LightStatus = "on"
	LightStatusOff       LightStatus = "off"
	LightStatusUndefined LightStatus = "undefined"
)

// ToBool converts a LightStatus to its bool value. LightStatusUndefined is considered as false.
func ToBool(status LightStatus) *bool {

	onValue, offValue := true, false

	if status == LightStatusOn {
		return &onValue
	} else if status == LightStatusOff {
		return &offValue
	} else {
		return &offValue
	}
}

type LightService interface {
	IsOn() bool
	Set(options SetLightOptions)
}

type Light struct {
	Resource
	LightService
	HueData *openhue.LightGet
}

func (light *Light) IsOn() bool {
	return *light.HueData.On.On
}

func (light *Light) Set(o *SetLightOptions) {
	request := &openhue.UpdateLightJSONRequestBody{}

	if o.Status != LightStatusUndefined {
		request.On = &openhue.On{
			On: ToBool(o.Status),
		}
	}

	if o.Brightness >= 0 && o.Brightness <= 100.0 {
		request.Dimming = &openhue.Dimming{
			Brightness: &o.Brightness,
		}
	}

	if o.Temperature >= 153 && o.Temperature <= 500 {
		request.ColorTemperature = &openhue.ColorTemperature{
			Mirek: &o.Temperature,
		}
	}

	if o.Color != color.UndefinedColor {
		request.Color = &openhue.Color{
			Xy: &openhue.GamutPosition{
				X: &o.Color.X,
				Y: &o.Color.Y,
			},
		}
	}

	if o.TransitionTime > 0 {
		request.Dynamics = &openhue.LightDynamics{
			Duration: &o.TransitionTime,
		}
	}

	err := light.ctx.h.UpdateLight(light.Id, *request)
	cobra.CheckErr(err)
}

type GroupedLight struct {
	Resource
	LightService
	HueData *openhue.GroupedLightGet
}

func (groupedLight *GroupedLight) IsOn() bool {
	return *groupedLight.HueData.On.On
}

func (groupedLight *GroupedLight) Set(o *SetLightOptions) {
	request := &openhue.UpdateGroupedLightJSONRequestBody{}

	if o.Status != LightStatusUndefined {
		request.On = &openhue.On{
			On: ToBool(o.Status),
		}
	}

	if o.Brightness >= 0 && o.Brightness <= 100.0 {
		request.Dimming = &openhue.Dimming{
			Brightness: &o.Brightness,
		}
	}

	if o.Temperature >= 153 && o.Temperature <= 500 {
		request.ColorTemperature = &openhue.ColorTemperature{
			Mirek: &o.Temperature,
		}
	}

	if o.Color != color.UndefinedColor {
		request.Color = &openhue.Color{
			Xy: &openhue.GamutPosition{
				X: &o.Color.X,
				Y: &o.Color.Y,
			},
		}
	}

	if o.TransitionTime > 0 {
		request.Dynamics = &openhue.Dynamics{
			Duration: &o.TransitionTime,
		}
	}

	err := groupedLight.ctx.h.UpdateGroupedLight(groupedLight.Id, *request)
	cobra.CheckErr(err)
}

//
// Scene
//

type SceneService interface {
	Activate()
}

type Scene struct {
	Resource
	SceneService
	HueData *openhue.SceneGet
}

func (scene *Scene) Activate(action openhue.SceneRecallAction) error {
	body := openhue.UpdateSceneJSONRequestBody{
		Recall: &openhue.SceneRecall{
			Action: &action,
		},
	}

	err := scene.ctx.h.UpdateScene(scene.Id, body)
	cobra.CheckErr(err)

	return nil
}
