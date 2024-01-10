package openhue

import (
	log "github.com/sirupsen/logrus"
	"slices"
)

// FindAllLights returns all the lights that belong to a Home
func FindAllLights(home *Home) []Light {
	var lights []Light

	for _, room := range home.Rooms {
		for _, device := range room.Devices {
			if device.Light != nil {
				lights = append(lights, *device.Light)
			}
		}
	}

	return lights
}

// FindLightByName returns a slice containing a single Light. The slice will be empty if no light was found.
func FindLightByName(home *Home, name string) []Light {
	for _, room := range home.Rooms {
		for _, device := range room.Devices {
			if device.Light != nil {
				if device.Light.Name == name {
					return []Light{*device.Light}
				}
			}
		}
	}

	log.Warn("no light found with name ", name)
	return []Light{}
}

func FindLightsByIds(home *Home, ids []string) []Light {

	var lights []Light

	for _, room := range home.Rooms {
		for _, device := range room.Devices {
			if device.Light != nil {
				if slices.Contains(ids, device.Light.Id) {
					lights = append(lights, *device.Light)
				}
			}
		}
	}

	return lights
}

//
// Room
//

func FindRoomById(home *Home, id string) []Room {
	for _, room := range home.Rooms {
		if room.Id == id {
			return []Room{room}
		}
	}

	log.Warn("no light found with ID ", id)
	return []Room{}
}

func FindRoomByName(home *Home, name string) []Room {
	for _, room := range home.Rooms {
		if room.Name == name {
			return []Room{room}
		}
	}

	log.Warn("no room found with name ", name)
	return []Room{}
}

func FindAllRooms(home *Home) []Room {
	var rooms []Room

	for _, room := range home.Rooms {
		rooms = append(rooms, room)
	}

	return rooms
}
