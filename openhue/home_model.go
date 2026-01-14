package openhue

import (
	"openhue-cli/util/color"
	"time"

	"github.com/openhue/openhue-go"
)

// HomeResourceType represents the type of a Hue resource (light, room, scene, etc.)
type HomeResourceType openhue.ResourceIdentifierRtype

// Resource is the base type for all Hue resources, containing common fields
// like ID, Name, Type, and a reference to the parent resource in the hierarchy.
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

// HomeModel represents the root of the Hue home hierarchy, containing all rooms and devices.
type HomeModel struct {
	Resource
	Rooms   []Room
	Devices []Device
	HueData *openhue.BridgeHomeGet
}

// Room represents a room in the Hue system, containing devices and scenes.
type Room struct {
	Resource
	Devices []Device
	Scenes  []Scene
	HueData *openhue.RoomGet

	// Services
	GroupedLight *GroupedLight
}

// Device represents a physical Hue device (bulb, sensor, etc.)
type Device struct {
	Resource
	HueData *openhue.DeviceGet

	// Services
	Light *Light
}

// SetLightOptions contains the parameters for controlling a light or group of lights.
// Values set to their "undefined" state (-1 for numbers, UndefinedColor for color) are ignored.
type SetLightOptions struct {
	Status         LightStatus
	Brightness     float32       // 0-100 percentage, -1 means unchanged
	Color          color.XY      // CIE color space coordinates
	Temperature    int           // Color temperature in Mirek (153-500), -1 means unchanged
	TransitionTime time.Duration // Transition duration for the change
}

// NewSetLightOptions creates a SetLightOptions with all values set to undefined/unchanged.
func NewSetLightOptions() *SetLightOptions {
	return &SetLightOptions{
		Status:      LightStatusUndefined,
		Brightness:  -1,
		Color:       color.UndefinedColor,
		Temperature: -1,
	}
}

// LightStatus represents the on/off state of a light
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

// LightService defines the common interface for controllable lights
type LightService interface {
	IsOn() bool
	Set(options *SetLightOptions) error
}

// Light represents a single light device
type Light struct {
	Resource
	LightService
	HueData *openhue.LightGet
}

func (light *Light) IsOn() bool {
	if light.HueData == nil || light.HueData.On == nil || light.HueData.On.On == nil {
		return false
	}
	return *light.HueData.On.On
}

func (light *Light) Set(o *SetLightOptions) error {
	request := &openhue.UpdateLightJSONRequestBody{
		On:               buildOnConfig(o),
		Dimming:          buildDimmingConfig(o),
		ColorTemperature: buildColorTemperatureConfig(o),
		Color:            buildColorConfig(o),
	}

	if o.TransitionTime > 0 {
		duration := int(o.TransitionTime.Milliseconds())
		request.Dynamics = &openhue.LightDynamics{
			Duration: &duration,
		}
	}

	return light.ctx.h.UpdateLight(light.Id, *request)
}

// GroupedLight represents a group of lights (typically a room)
type GroupedLight struct {
	Resource
	LightService
	HueData *openhue.GroupedLightGet
}

func (groupedLight *GroupedLight) IsOn() bool {
	// A room may not have any light, in this case the GroupedLight service is nil.
	if groupedLight == nil || groupedLight.HueData == nil || groupedLight.HueData.On == nil || groupedLight.HueData.On.On == nil {
		return false
	}
	return *groupedLight.HueData.On.On
}

func (groupedLight *GroupedLight) Set(o *SetLightOptions) error {
	request := &openhue.UpdateGroupedLightJSONRequestBody{
		On:               buildOnConfig(o),
		Dimming:          buildDimmingConfig(o),
		ColorTemperature: buildColorTemperatureConfig(o),
		Color:            buildColorConfig(o),
	}

	if o.TransitionTime > 0 {
		duration := int(o.TransitionTime.Milliseconds())
		request.Dynamics = &openhue.Dynamics{
			Duration: &duration,
		}
	}

	return groupedLight.ctx.h.UpdateGroupedLight(groupedLight.Id, *request)
}

// Helper functions to build common light configuration parts

func buildOnConfig(o *SetLightOptions) *openhue.On {
	if o.Status == LightStatusUndefined {
		return nil
	}
	return &openhue.On{On: ToBool(o.Status)}
}

func buildDimmingConfig(o *SetLightOptions) *openhue.Dimming {
	if o.Brightness < 0 || o.Brightness > 100.0 {
		return nil
	}
	return &openhue.Dimming{Brightness: &o.Brightness}
}

func buildColorTemperatureConfig(o *SetLightOptions) *openhue.ColorTemperature {
	if o.Temperature < 153 || o.Temperature > 500 {
		return nil
	}
	return &openhue.ColorTemperature{Mirek: &o.Temperature}
}

func buildColorConfig(o *SetLightOptions) *openhue.Color {
	if o.Color == color.UndefinedColor {
		return nil
	}
	return &openhue.Color{
		Xy: &openhue.GamutPosition{
			X: &o.Color.X,
			Y: &o.Color.Y,
		},
	}
}

// SceneService defines the interface for activating scenes
type SceneService interface {
	Activate(action openhue.SceneRecallAction) error
}

// Scene represents a Hue scene - a predefined lighting configuration for a room
type Scene struct {
	Resource
	SceneService
	HueData *openhue.SceneGet
}

// Activate triggers the scene with the specified action (active, static, or dynamic)
func (scene *Scene) Activate(action openhue.SceneRecallAction) error {
	body := openhue.UpdateSceneJSONRequestBody{
		Recall: &openhue.SceneRecall{
			Action: &action,
		},
	}

	return scene.ctx.h.UpdateScene(scene.Id, body)
}
