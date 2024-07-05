package openhue

import (
	"github.com/openhue/openhue-go"
	log "github.com/sirupsen/logrus"
)

type hueHomeCtx struct {
	h *openhue.Home
	// context
	home          *openhue.BridgeHomeGet
	rooms         map[string]openhue.RoomGet
	devices       map[string]openhue.DeviceGet
	lights        map[string]openhue.LightGet
	groupedLights map[string]openhue.GroupedLightGet
	scenes        map[string]openhue.SceneGet
}

func loadHueHomeCtx(home *openhue.Home) (*hueHomeCtx, error) {

	hueHome, err := home.GetBridgeHome()
	if err != nil {
		return nil, err
	}

	rooms, err := home.GetRooms()
	if err != nil {
		return nil, err
	}

	devices, err := home.GetDevices()
	if err != nil {
		return nil, err
	}

	lights, err := home.GetLights()
	if err != nil {
		return nil, err
	}

	groupedLights, err := home.GetGroupedLights()
	if err != nil {
		return nil, err
	}

	scenes, err := home.GetScenes()
	if err != nil {
		return nil, err
	}

	log.Info("HomeModel loaded")

	return &hueHomeCtx{
		h:             home,
		home:          hueHome,
		rooms:         rooms,
		devices:       devices,
		lights:        lights,
		groupedLights: groupedLights,
		scenes:        scenes,
	}, nil
}
