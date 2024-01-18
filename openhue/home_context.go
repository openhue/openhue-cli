package openhue

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"openhue-cli/openhue/gen"
)

type hueHomeCtx struct {
	// api
	api *gen.ClientWithResponses
	// context
	home          *gen.BridgeHomeGet
	rooms         map[string]gen.RoomGet
	devices       map[string]gen.DeviceGet
	lights        map[string]gen.LightGet
	groupedLights map[string]gen.GroupedLightGet
	scenes        map[string]gen.SceneGet
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

	scenes, err := fetchScenes(api)
	if err != nil {
		return nil, err
	}

	log.Info("Home loaded")

	return &hueHomeCtx{
		api:           api,
		home:          hueHome,
		rooms:         rooms,
		devices:       devices,
		lights:        lights,
		groupedLights: groupedLights,
		scenes:        scenes,
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

func fetchScenes(api *gen.ClientWithResponses) (map[string]gen.SceneGet, error) {
	log.Info("Fetching scenes...")

	resp, err := api.GetScenesWithResponse(context.Background())
	if err != nil {
		return nil, err
	}

	scenes := make(map[string]gen.SceneGet)
	hueScenes := (*resp.JSON200).Data

	for _, scene := range *hueScenes {
		scenes[*scene.Id] = scene
	}

	return scenes, err
}
