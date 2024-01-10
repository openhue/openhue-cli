package openhue

import (
	"context"
	"errors"
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

type hueHomeCtx struct {
	// api
	api *gen.ClientWithResponses
	// context
	home          *gen.BridgeHomeGet
	rooms         map[string]gen.RoomGet
	devices       map[string]gen.DeviceGet
	lights        map[string]gen.LightGet
	groupedLights map[string]gen.GroupedLightGet
}

func loadHueHomeCtx(api *gen.ClientWithResponses) (*hueHomeCtx, error) {

	hueHome, err := fetchBridgeHome(api)
	if err != nil {
		return nil, err
	}

	rooms, err := fetchRooms(api)
	if err != nil {
		return nil, err
	}

	devices, err := fetchDevices(api)
	if err != nil {
		return nil, err
	}

	lights, err := fetchLights(api)
	if err != nil {
		return nil, err
	}

	groupedLights, err := fetchGroupedLights(api)
	if err != nil {
		return nil, err
	}

	return &hueHomeCtx{
		api:           api,
		home:          hueHome,
		rooms:         rooms,
		devices:       devices,
		lights:        lights,
		groupedLights: groupedLights,
	}, nil
}

func fetchBridgeHome(api *gen.ClientWithResponses) (*gen.BridgeHomeGet, error) {
	log.Info("Fetching home...")

	resp, err := api.GetBridgeHomesWithResponse(context.Background())
	if err != nil {
		return nil, err
	}

	if len(*(*resp.JSON200).Data) != 1 {
		return nil, errors.New("more than 1 home attached to the bridge is not supported yet")
	}

	return &(*(*resp.JSON200).Data)[0], nil
}

func fetchDevices(api *gen.ClientWithResponses) (map[string]gen.DeviceGet, error) {
	log.Info("Fetching devices...")

	resp, err := api.GetDevicesWithResponse(context.Background())
	if err != nil {
		return nil, err
	}

	devices := make(map[string]gen.DeviceGet)
	hueDevices := (*resp.JSON200).Data

	for _, device := range *hueDevices {
		devices[*device.Id] = device
	}

	return devices, err
}

func fetchRooms(api *gen.ClientWithResponses) (map[string]gen.RoomGet, error) {
	log.Info("Fetching rooms...")

	resp, err := api.GetRoomsWithResponse(context.Background())
	if err != nil {
		return nil, err
	}

	rooms := make(map[string]gen.RoomGet)
	hueRooms := (*resp.JSON200).Data

	for _, room := range *hueRooms {
		rooms[*room.Id] = room
	}

	return rooms, err
}

func fetchLights(api *gen.ClientWithResponses) (map[string]gen.LightGet, error) {
	log.Info("Fetching lights...")

	resp, err := api.GetLightsWithResponse(context.Background())
	if err != nil {
		return nil, err
	}

	lights := make(map[string]gen.LightGet)
	hueLights := (*resp.JSON200).Data

	for _, light := range *hueLights {
		lights[*light.Id] = light
	}

	return lights, nil
}

func fetchGroupedLights(api *gen.ClientWithResponses) (map[string]gen.GroupedLightGet, error) {
	log.Info("Fetching grouped lights...")

	resp, err := api.GetGroupedLightsWithResponse(context.Background())
	if err != nil {
		return nil, err
	}

	groupedLights := make(map[string]gen.GroupedLightGet)
	hueGroupedLights := (*resp.JSON200).Data

	for _, groupedLight := range *hueGroupedLights {
		groupedLights[*groupedLight.Id] = groupedLight
	}

	return groupedLights, err
}
