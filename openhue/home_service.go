package openhue

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"openhue-cli/openhue/gen"
	"slices"
)

func LoadHome(api *gen.ClientWithResponses) (*Home, error) {
	log.Info("Loading home...")

	ctx, err := loadHueHomeCtx(api)
	if err != nil {
		return nil, err
	}

	home := Home{
		Resource: Resource{
			ctx:    ctx,
			Id:     *ctx.home.Id,
			Name:   "Home",
			Type:   HomeResourceType(gen.ResourceIdentifierRtypeBridgeHome),
			Parent: nil, // explicitly set Parent to nil as Home is the root object
		},
		HueData: ctx.home,
	}

	home.Rooms = getRooms(ctx, &home.Resource, home.HueData.Children)
	home.Devices = getDevices(ctx, &home.Resource, home.HueData.Children)

	return &home, nil
}

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

func FindAllRooms(home *Home) []Room {
	var rooms []Room

	for _, room := range home.Rooms {
		rooms = append(rooms, room)
	}

	return rooms
}

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

//
// Private
//

func (t HomeResourceType) is(r gen.ResourceIdentifierRtype) bool {

	if t == HomeResourceType(r) {
		return true
	}

	return false
}

func getRooms(ctx *hueHomeCtx, parent *Resource, children *[]gen.ResourceIdentifier) []Room {

	var rooms []Room

	for _, child := range *children {
		rType := HomeResourceType(*child.Rtype)
		if rType.is(gen.ResourceIdentifierRtypeRoom) {
			hueRoom := ctx.rooms[*child.Rid]
			room := Room{
				Resource: Resource{
					ctx:    ctx,
					Id:     *hueRoom.Id,
					Name:   *hueRoom.Metadata.Name,
					Type:   rType,
					Parent: parent,
				},
				HueData: &hueRoom,
			}

			room.Devices = getDevices(ctx, &room.Resource, hueRoom.Children)

			groupedLight, err := getGroupedLight(ctx, &room.Resource, hueRoom.Services)
			if err != nil {
				log.Warning(err)
			} else {
				room.GroupedLight = groupedLight
			}

			rooms = append(rooms, room)
		}
	}

	return rooms
}

func getDevices(ctx *hueHomeCtx, parent *Resource, children *[]gen.ResourceIdentifier) []Device {

	var devices []Device

	for _, child := range *children {
		rType := HomeResourceType(*child.Rtype)
		if rType.is(gen.ResourceIdentifierRtypeDevice) {
			hueDevice := ctx.devices[*child.Rid]
			device := Device{
				Resource: Resource{
					ctx:    ctx,
					Id:     *hueDevice.Id,
					Name:   *hueDevice.Metadata.Name,
					Type:   rType,
					Parent: parent,
				},
				HueData: &hueDevice,
			}

			light, err := getLight(ctx, &device.Resource, hueDevice.Services)
			if err != nil {
				log.Warning(err)
			} else {
				device.Light = light
			}

			devices = append(devices, device)
		}
	}

	return devices
}

func getGroupedLight(ctx *hueHomeCtx, parent *Resource, services *[]gen.ResourceIdentifier) (*GroupedLight, error) {

	for _, service := range *services {
		rType := HomeResourceType(*service.Rtype)
		if rType.is(gen.ResourceIdentifierRtypeGroupedLight) {
			hueGroupedLight := ctx.groupedLights[*service.Rid]
			return &GroupedLight{
				Resource: Resource{
					ctx:    ctx,
					Id:     *hueGroupedLight.Id,
					Name:   "Grouped Light (" + parent.Name + ")",
					Type:   rType,
					Parent: parent.Parent,
				},
				HueData: &hueGroupedLight,
			}, nil
		}
	}

	return nil, fmt.Errorf("no 'grouped_light' service found for resource %s (%s)", parent.Id, parent.Name)
}

func getLight(ctx *hueHomeCtx, parent *Resource, services *[]gen.ResourceIdentifier) (*Light, error) {

	for _, service := range *services {
		rType := HomeResourceType(*service.Rtype)
		if rType.is(gen.ResourceIdentifierRtypeLight) {
			light := ctx.lights[*service.Rid]
			return &Light{
				Resource: Resource{
					ctx:    ctx,
					Id:     *light.Id,
					Name:   *light.Metadata.Name,
					Type:   rType,
					Parent: parent,
				},
				HueData: &light,
			}, nil
		}
	}

	return nil, fmt.Errorf("no 'light' service found for resource %s (%s)", parent.Id, parent.Name)
}
