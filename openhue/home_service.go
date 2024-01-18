package openhue

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"openhue-cli/openhue/gen"
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

//
// Light
//

// SearchLights returns a slice of Light optionally filtered by their room and their IDs or names
func SearchLights(home *Home, roomNameOrId string, nameOrIds []string) []Light {
	var lights []Light

	for _, room := range home.Rooms {
		if room.matchesNameOrId(roomNameOrId) {
			for _, device := range room.Devices {
				if device.Light != nil {
					if len(nameOrIds) == 0 {
						lights = append(lights, *device.Light)
					} else {
						for _, nameOrId := range nameOrIds {
							if device.Light.matchesNameOrId(nameOrId) {
								lights = append(lights, *device.Light)
							}
						}
					}
				}
			}
		}
	}

	return lights
}

//
// Room
//

func SearchRooms(home *Home, nameOrIds []string) []Room {
	var rooms []Room

	for _, room := range home.Rooms {
		if len(nameOrIds) == 0 {
			rooms = append(rooms, room)
		} else {
			for _, nameOrId := range nameOrIds {
				if room.matchesNameOrId(nameOrId) {
					rooms = append(rooms, room)
				}
			}
		}
	}

	return rooms
}

//
// Scene
//

// SearchScenes returns a slice of Scene filtered by the room
func SearchScenes(home *Home, roomNameOrId string, nameOrIds []string) []Scene {
	var scenes []Scene

	for _, room := range home.Rooms {
		if len(roomNameOrId) == 0 || (len(roomNameOrId) > 0 && room.matchesNameOrId(roomNameOrId)) {
			for _, scene := range room.Scenes {
				if len(nameOrIds) == 0 {
					scenes = append(scenes, scene)
				} else {
					for _, nameOrId := range nameOrIds {
						if scene.Resource.matchesNameOrId(nameOrId) {
							scenes = append(scenes, scene)
						}
					}
				}
			}
		}
	}

	return scenes
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
			room.Scenes = getScenes(ctx, &room.Resource)

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

// getScenes returns a slice of Scene that belong to a given Room
func getScenes(ctx *hueHomeCtx, room *Resource) []Scene {
	var scenes []Scene

	for _, scene := range ctx.scenes {
		hueScene := scene
		if *scene.Group.Rid == room.Id {
			scenes = append(scenes, Scene{
				Resource: Resource{
					Id:     *scene.Id,
					Name:   *scene.Metadata.Name,
					Type:   HomeResourceType(gen.ResourceIdentifierRtypeScene),
					Parent: room,
					ctx:    ctx,
				},
				HueData: &hueScene,
			})
		}
	}

	return scenes
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
