package openhue

import (
	"github.com/stretchr/testify/assert"
	"openhue-cli/openhue/gen"
	"testing"
)

var home = mockHome()

func TestFindAllLights(t *testing.T) {
	lights := SearchLights(home, "", []string{})
	assert.Len(t, lights, 4, "Home should contain 4 lights")
}

func TestFindLightsByName(t *testing.T) {
	lights := SearchLights(home, "", []string{"room1_light2"})
	assert.Len(t, lights, 1, "Home should contain 1 single light for this name")
	assert.Empty(t, SearchLights(home, "", []string{"foo"}), "Home should not contain any light with name 'foo'")
}

func TestFindLightsByIds(t *testing.T) {
	lights := SearchLights(home, "", []string{"room1_light2", "room2_light1"})
	assert.Len(t, lights, 2, "Home should contain 2 lights for this name")
}

func TestFindAllRooms(t *testing.T) {
	rooms := SearchRooms(home, []string{})
	assert.Len(t, rooms, 2, "Home should contain 2 rooms")
}

func TestFindRoomsByName(t *testing.T) {
	rooms := SearchRooms(home, []string{"room1"})
	assert.Len(t, rooms, 1, "Home should contain 1 single room for this name")
	assert.Empty(t, SearchRooms(home, []string{"foo"}), "Home should not contain any room with name 'foo'")
}

func TestFindRoomById(t *testing.T) {
	rooms := SearchRooms(home, []string{"room2"})
	assert.Len(t, rooms, 1, "Home should contain 1 single room for this name")
	assert.Empty(t, SearchRooms(home, []string{"foo"}), "Home should not contain any room with ID 'foo'")
}

//
// internal
//

func mockHome() *Home {
	home := Home{}

	home.Rooms = []Room{
		mockRoom("room1"),
		mockRoom("room2"),
	}

	return &home
}

func mockRoom(name string) Room {
	room := Room{
		Resource: Resource{
			Id:     name,
			Name:   name,
			Type:   HomeResourceType(gen.ResourceGetTypeRoom),
			Parent: nil,
			ctx:    nil,
		},
		HueData:      nil,
		GroupedLight: nil,
	}

	room.Devices = []Device{
		mockDeviceWithLight(room, name+"_light1"),
		mockDeviceWithLight(room, name+"_light2"),
	}

	return room
}

func mockDeviceWithLight(parent Room, name string) Device {
	device := Device{
		Resource: Resource{
			Id:     name,
			Name:   name,
			Type:   HomeResourceType(gen.ResourceGetTypeLight),
			Parent: &parent.Resource,
			ctx:    nil,
		},
		HueData: nil,
	}

	device.Light = &Light{
		Resource: Resource{
			Id:     name,
			Name:   name,
			Parent: &device.Resource,
		},
		LightService: nil,
		HueData:      nil,
	}

	return device
}
