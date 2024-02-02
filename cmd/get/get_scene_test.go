package get

import (
	"github.com/stretchr/testify/assert"
	"openhue-cli/openhue"
	"openhue-cli/openhue/gen"
	"testing"
)

func TestGetAllScenes(t *testing.T) {
	mockCtx, _ := openhue.NewTestContext(NewFakeHome())
	cmd := NewCmdGetScene(mockCtx, &CmdGetOptions{Json: true})
	err := cmd.Execute()
	assert.Nil(t, err)
}

func NewFakeHome() *openhue.Home {
	return &openhue.Home{
		Resource: openhue.Resource{
			Id:     "home-01",
			Name:   "Home",
			Type:   openhue.HomeResourceType(gen.BridgeHomeGetTypeBridgeHome),
			Parent: nil,
		},
		Rooms: []openhue.Room{
			{
				Resource: openhue.Resource{
					Id:   "room-01",
					Name: "Room 1",
					Type: openhue.HomeResourceType(gen.ResourceGetTypeRoom),
				},
				Devices: nil,
				Scenes: []openhue.Scene{
					{
						Resource: openhue.Resource{
							Id:   "scene-01",
							Name: "Soho",
						},
						HueData: &gen.SceneGet{},
					},
				},
				GroupedLight: nil,
				HueData:      &gen.RoomGet{},
			},
		},
		Devices: []openhue.Device{},
		HueData: &gen.BridgeHomeGet{},
	}
}
