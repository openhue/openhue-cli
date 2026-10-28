package fake

import (
	"net/http"

	oh "github.com/openhue/openhue-go"
	"github.com/stretchr/testify/mock"
	"openhue-cli/openhue"
)

// FakeLight describes a light of the fake home.
type FakeLight struct {
	Name     string
	On       bool
	Bri      float32
	Mirek    *int
	XY       *[2]float32
	Color    bool // supports color
	Ct       bool // supports color temperature
	Gradient bool
	Effects  bool

	NoDimming bool
}

// NewFakeLight returns a color + temperature capable light that is on at 50% in warm white.
func NewFakeLight(name string) FakeLight {
	m := 366
	return FakeLight{Name: name, On: true, Bri: 50, Mirek: &m, Color: true, Ct: true}
}

// NewFakeHome returns a home with one room "Office" (id room-1) holding the given lights (ids light-<name>)
// and the given scenes.
func NewFakeHome(lights []FakeLight, scenes ...openhue.Scene) *openhue.HomeModel {
	room := openhue.Room{
		Resource: openhue.Resource{Id: "room-1", Name: "Office", Type: openhue.HomeResourceType(oh.ResourceGetTypeRoom)},
		HueData:  &oh.RoomGet{},
	}
	for _, l := range lights {
		room.Devices = append(room.Devices, openhue.Device{
			Resource: openhue.Resource{Id: "dev-" + l.Name, Name: l.Name},
			Light: &openhue.Light{
				Resource: openhue.Resource{Id: "light-" + l.Name, Name: l.Name},
				HueData:  fakeLightData(l),
			},
		})
	}
	for _, s := range scenes {
		s.Parent = &room.Resource
		room.Scenes = append(room.Scenes, s)
	}
	return &openhue.HomeModel{Rooms: []openhue.Room{room}, HueData: &oh.BridgeHomeGet{}}
}

func fakeLightData(l FakeLight) *oh.LightGet {
	d := &oh.LightGet{Id: "light-" + l.Name, On: oh.On{On: l.On}}
	b := l.Bri
	if !l.NoDimming {
		d.Dimming = &struct {
			Brightness  *oh.Brightness `json:"brightness,omitempty"`
			MinDimLevel *float32       `json:"min_dim_level,omitempty"`
		}{Brightness: &b}
	}
	if l.Ct {
		d.ColorTemperature = &struct {
			Mirek       *int `json:"mirek,omitempty"`
			MirekSchema *struct {
				MirekMaximum *int `json:"mirek_maximum,omitempty"`
				MirekMinimum *int `json:"mirek_minimum,omitempty"`
			} `json:"mirek_schema,omitempty"`
			MirekValid *bool `json:"mirek_valid,omitempty"`
		}{Mirek: l.Mirek}
	}
	if l.Color {
		xy := oh.GamutPosition{X: 0.3, Y: 0.3}
		if l.XY != nil {
			xy = oh.GamutPosition{X: l.XY[0], Y: l.XY[1]}
		}
		d.Color = &struct {
			Gamut *struct {
				Blue  *oh.GamutPosition `json:"blue,omitempty"`
				Green *oh.GamutPosition `json:"green,omitempty"`
				Red   *oh.GamutPosition `json:"red,omitempty"`
			} `json:"gamut,omitempty"`
			GamutType *oh.LightGetColorGamutType `json:"gamut_type,omitempty"`
			Xy        *oh.GamutPosition          `json:"xy,omitempty"`
		}{Xy: &xy}
	}
	if l.Gradient {
		d.Gradient = &struct {
			Mode          *oh.SupportedGradientMode   `json:"mode,omitempty"`
			ModeValues    *[]oh.SupportedGradientMode `json:"mode_values,omitempty"`
			PixelCount    *int                        `json:"pixel_count,omitempty"`
			Points        *[]oh.Color                 `json:"points,omitempty"`
			PointsCapable *int                        `json:"points_capable,omitempty"`
		}{}
	}
	if l.Effects {
		d.Effects = &struct {
			Effect       *oh.SupportedEffects   `json:"effect,omitempty"`
			EffectValues *[]oh.SupportedEffects `json:"effect_values,omitempty"`
			Status       *oh.SupportedEffects   `json:"status,omitempty"`
			StatusValues *[]oh.SupportedEffects `json:"status_values,omitempty"`
		}{}
	}
	return d
}

// NewFakeScene returns a scene of the fake room with the given stored actions.
func NewFakeScene(id, name string, actions ...oh.ActionGet) openhue.Scene {
	n := name
	return openhue.Scene{
		Resource: openhue.Resource{Id: id, Name: name},
		HueData: &oh.SceneGet{
			Id:       id,
			Actions:  actions,
			Metadata: oh.SceneMetadata{Name: &n},
			Speed:    0.5,
		},
	}
}

// StoredAction returns a stored action of a light at the given brightness in the given mirek.
func StoredAction(lightName string, on bool, brightness float32, mirek int) oh.ActionGet {
	a := oh.ActionGet{Target: oh.ResourceIdentifier{Rid: "light-" + lightName, Rtype: oh.ResourceIdentifierRtypeLight}}
	a.Action.On = &oh.On{On: on}
	a.Action.Dimming = &oh.Dimming{Brightness: brightness}
	a.Action.ColorTemperature = &oh.ColorTemperature{Mirek: &mirek}
	return a
}

// NewTestHomeWithHandlers returns an openhue-go Home backed by the mock where scene creation, update and
// deletion succeed (creating the scene "new-scene").
func NewTestHomeWithHandlers() (*oh.Home, *oh.ClientWithResponsesMock) {
	h, m := oh.NewTestHome()
	ok := &http.Response{StatusCode: http.StatusOK}
	m.On("CreateSceneWithResponse", mock.Anything, mock.Anything, mock.Anything).Return(&oh.CreateSceneResponse{
		HTTPResponse: ok,
		JSON200: &struct {
			Data   []oh.ResourceIdentifier `json:"data"`
			Errors []oh.Error              `json:"errors"`
		}{Data: []oh.ResourceIdentifier{{Rid: "new-scene", Rtype: oh.ResourceIdentifierRtypeScene}}},
	}, nil).Maybe()
	m.On("UpdateSceneWithResponse", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&oh.UpdateSceneResponse{HTTPResponse: ok}, nil).Maybe()
	m.On("DeleteSceneWithResponse", mock.Anything, mock.Anything, mock.Anything).Return(&oh.DeleteSceneResponse{HTTPResponse: ok}, nil).Maybe()
	return h, m
}
