package openhue

import (
	"context"

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
	buttons       map[string]openhue.ButtonGet
	rotaries      map[string]openhue.RelativeRotaryGet
	behaviors     map[string]openhue.BehaviorInstanceGet
}

func loadHueHomeCtx(home *openhue.Home) (*hueHomeCtx, error) {
	ctx := context.Background()

	hueHome, err := home.GetBridgeHome(ctx)
	if err != nil {
		return nil, err
	}

	rooms, err := home.GetRooms(ctx)
	if err != nil {
		return nil, err
	}

	devices, err := home.GetDevices(ctx)
	if err != nil {
		return nil, err
	}

	lights, err := home.GetLights(ctx)
	if err != nil {
		return nil, err
	}

	groupedLights, err := home.GetGroupedLights(ctx)
	if err != nil {
		return nil, err
	}

	scenes, err := home.GetScenes(ctx)
	if err != nil {
		return nil, err
	}

	// Controls are not essential to light and scene commands: tolerate failures.
	buttons, err := home.GetButtons(ctx)
	if err != nil {
		log.Warnf("unable to load buttons: %v", err)
		buttons = nil
	}

	rotaries, err := home.GetRelativeRotaries(ctx)
	if err != nil {
		log.Warnf("unable to load relative rotaries: %v", err)
		rotaries = nil
	}

	// Behavior instances only describe what the buttons do: tolerate failures too.
	behaviors, err := home.GetBehaviorInstances(ctx)
	if err != nil {
		log.Warnf("unable to load behavior instances: %v", err)
		behaviors = nil
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
		buttons:       buttons,
		rotaries:      rotaries,
		behaviors:     behaviors,
	}, nil
}
