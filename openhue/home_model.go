package openhue

import (
	"context"
	"github.com/spf13/cobra"
	"openhue-cli/openhue/gen"
	"openhue-cli/util/color"
)

type HomeResourceType gen.ResourceIdentifierRtype

type Resource struct {
	Id     string
	Name   string
	Type   HomeResourceType
	Parent *Resource
	// internal
	ctx *hueHomeCtx
}

//
// Home
//

type Home struct {
	Resource
	Rooms   []Room
	Devices []Device
	HueData *gen.BridgeHomeGet
}

//
// Room
//

type Room struct {
	Resource
	Devices []Device
	HueData *gen.RoomGet

	// Services
	GroupedLight *GroupedLight
}

//
// Device
//

type Device struct {
	Resource
	HueData *gen.DeviceGet

	// Services
	Light *Light
}

//
// Light
//

type SetLightOptions struct {
	Status     LightStatus
	Brightness float32
	Color      color.XY
}

func NewSetLightOptions() *SetLightOptions {
	return &SetLightOptions{
		Status:     LightStatusUndefined,
		Brightness: -1,
		Color:      color.UndefinedColor,
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
	HueData *gen.LightGet
}

func (light *Light) IsOn() bool {
	return *light.HueData.On.On
}

func (light *Light) Set(o *SetLightOptions) {
	request := &gen.UpdateLightJSONRequestBody{}

	if o.Status != LightStatusUndefined {
		request.On = &gen.On{
			On: ToBool(o.Status),
		}
	}

	if o.Brightness >= 0 && o.Brightness <= 100.0 {
		request.Dimming = &gen.Dimming{
			Brightness: &o.Brightness,
		}
	}

	if o.Color != color.UndefinedColor {
		request.Color = &gen.Color{
			Xy: &gen.GamutPosition{
				X: &o.Color.X,
				Y: &o.Color.Y,
			},
		}
	}

	_, err := light.ctx.api.UpdateLight(context.Background(), light.Id, *request)
	cobra.CheckErr(err)
}

type GroupedLight struct {
	Resource
	LightService
	HueData *gen.GroupedLightGet
}

func (groupedLight *GroupedLight) IsOn() bool {
	return *groupedLight.HueData.On.On
}

func (groupedLight *GroupedLight) Set(o *SetLightOptions) {
	request := &gen.UpdateGroupedLightJSONRequestBody{}

	if o.Status != LightStatusUndefined {
		request.On = &gen.On{
			On: ToBool(o.Status),
		}
	}

	if o.Brightness >= 0 && o.Brightness <= 100.0 {
		request.Dimming = &gen.Dimming{
			Brightness: &o.Brightness,
		}
	}

	if o.Color != color.UndefinedColor {
		request.Color = &gen.Color{
			Xy: &gen.GamutPosition{
				X: &o.Color.X,
				Y: &o.Color.Y,
			},
		}
	}

	_, err := groupedLight.ctx.api.UpdateGroupedLight(context.Background(), groupedLight.Id, *request)
	cobra.CheckErr(err)
}
