package get

import (
	oh "github.com/openhue/openhue-go"
	"github.com/stretchr/testify/assert"
	"openhue-cli/openhue"
	"testing"
)

func TestGetAllScenes(t *testing.T) {
	mockCtx, _ := openhue.NewTestContext(NewFakeHome())
	cmd := NewCmdGetScene(mockCtx, &CmdGetOptions{Json: true})
	err := cmd.Execute()
	assert.Nil(t, err)
}

func NewFakeHome() *openhue.HomeModel {
	return &openhue.HomeModel{
		Resource: openhue.Resource{
			Id:     "home-01",
			Name:   "HomeModel",
			Type:   openhue.HomeResourceType(oh.BridgeHomeGetTypeBridgeHome),
			Parent: nil,
		},
		Rooms: []openhue.Room{
			{
				Resource: openhue.Resource{
					Id:   "room-01",
					Name: "Room 1",
					Type: openhue.HomeResourceType(oh.ResourceGetTypeRoom),
				},
				Devices: nil,
				Scenes: []openhue.Scene{
					{
						Resource: openhue.Resource{
							Id:   "scene-01",
							Name: "Soho",
						},
						HueData: &oh.SceneGet{},
					},
				},
				GroupedLight: nil,
				HueData:      &oh.RoomGet{},
			},
		},
		Devices: []openhue.Device{},
		HueData: &oh.BridgeHomeGet{},
	}
}
