package openhue

import (
	"fmt"
	"github.com/openhue/openhue-go"
	log "github.com/sirupsen/logrus"
	"slices"
	"sort"
	"strings"
)

// LoadHome loads all resources from the Hue bridge and builds the HomeModel hierarchy.
// This includes rooms, devices, lights, grouped lights, and scenes.
func LoadHome(h *openhue.Home) (*HomeModel, error) {
	log.Info("Loading home...")

	ctx, err := loadHueHomeCtx(h)
	if err != nil {
		return nil, err
	}

	home := HomeModel{
		Resource: Resource{
			ctx:    ctx,
			Id:     ctx.home.Id,
			Name:   "HomeModel",
			Type:   HomeResourceType(openhue.ResourceIdentifierRtypeBridgeHome),
			Parent: nil, // explicitly set Parent to nil as HomeModel is the root object
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
func SearchLights(home *HomeModel, roomNameOrId string, nameOrIds []string) []Light {
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

// SearchRooms returns a slice of Room optionally filtered by their IDs or names.
// If nameOrIds is empty, all rooms are returned.
func SearchRooms(home *HomeModel, nameOrIds []string) []Room {
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
func SearchScenes(home *HomeModel, roomNameOrId string, nameOrIds []string) []Scene {
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
// Control
//

// allDevices returns every device of the home, in rooms or not.
func allDevices(home *HomeModel) []Device {
	var devices []Device
	for _, room := range home.Rooms {
		devices = append(devices, room.Devices...)
	}
	return append(devices, home.Devices...)
}

// SearchControls returns the controls of the home, optionally filtered by room (name or ID), device (name or ID),
// kind (empty for all) and control IDs. Filters are combined.
func SearchControls(home *HomeModel, room, device string, kind ControlKind, ids []string) []Control {
	controls := []Control{}

	for _, d := range allDevices(home) {
		if !d.matchesNameOrId(device) {
			continue
		}
		for _, c := range d.Controls {
			if kind != "" && c.Kind != kind {
				continue
			}
			if room != "" {
				r := c.Room()
				if r == nil || !r.matchesNameOrId(room) {
					continue
				}
			}
			if len(ids) > 0 && !slices.Contains(ids, c.Id) {
				continue
			}
			controls = append(controls, c)
		}
	}

	return controls
}

// FindControl resolves exactly one control from either a control ID, or a device name or ID together with a
// control number (controlNumber > 0). Zero or several matches are errors listing the candidates.
func FindControl(home *HomeModel, deviceOrId string, controlNumber int) (*Control, error) {
	var matches []Control

	if controlNumber <= 0 {
		matches = SearchControls(home, "", "", "", []string{deviceOrId})
	}
	if len(matches) == 0 {
		for _, c := range SearchControls(home, "", deviceOrId, "", nil) {
			if controlNumber <= 0 || c.ControlId == controlNumber {
				matches = append(matches, c)
			}
		}
	}

	switch len(matches) {
	case 1:
		return &matches[0], nil
	case 0:
		return nil, fmt.Errorf("no control found for '%s'%s", deviceOrId, controlNumberMsg(controlNumber))
	default:
		return nil, fmt.Errorf("%d controls match '%s'%s, use a control ID or --control to select one:\n%s",
			len(matches), deviceOrId, controlNumberMsg(controlNumber), describeControls(matches))
	}
}

func controlNumberMsg(n int) string {
	if n > 0 {
		return fmt.Sprintf(" with control number %d", n)
	}
	return ""
}

func describeControls(controls []Control) string {
	var lines []string
	for _, c := range controls {
		lines = append(lines, fmt.Sprintf("  %s  %s (control %d)", c.Id, c.Name, c.ControlId))
	}
	return strings.Join(lines, "\n")
}

//
// Private
//

func (t HomeResourceType) is(r openhue.ResourceIdentifierRtype) bool {

	if t == HomeResourceType(r) {
		return true
	}

	return false
}

func getRooms(ctx *hueHomeCtx, parent *Resource, children []openhue.ResourceIdentifier) []Room {

	var rooms []Room

	for _, child := range children {
		rType := HomeResourceType(child.Rtype)
		if rType.is(openhue.ResourceIdentifierRtypeRoom) {
			hueRoom := ctx.rooms[child.Rid]
			room := Room{
				Resource: Resource{
					ctx:    ctx,
					Id:     hueRoom.Id,
					Name:   hueRoom.Metadata.Name,
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
		if scene.Group.Rid == room.Id {
			scenes = append(scenes, Scene{
				Resource: Resource{
					Id:     scene.Id,
					Name:   *scene.Metadata.Name,
					Type:   HomeResourceType(openhue.ResourceIdentifierRtypeScene),
					Parent: room,
					ctx:    ctx,
				},
				HueData: &hueScene,
			})
		}
	}

	return scenes
}

func getDevices(ctx *hueHomeCtx, parent *Resource, children []openhue.ResourceIdentifier) []Device {

	var devices []Device

	for _, child := range children {
		rType := HomeResourceType(child.Rtype)
		if rType.is(openhue.ResourceIdentifierRtypeDevice) {
			hueDevice := ctx.devices[child.Rid]
			device := Device{
				Resource: Resource{
					ctx:    ctx,
					Id:     hueDevice.Id,
					Name:   hueDevice.Metadata.Name,
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

			device.Controls = getControls(ctx, &device.Resource)

			devices = append(devices, device)
		}
	}

	return devices
}

// getControls returns the buttons and relative rotaries owned by the given device, ordered by kind then control number.
func getControls(ctx *hueHomeCtx, device *Resource) []Control {
	var controls []Control

	for _, b := range ctx.buttons {
		if b.Owner.Rid != device.Id {
			continue
		}
		hueButton := b
		controls = append(controls, Control{
			Resource:  Resource{ctx: ctx, Id: b.Id, Name: controlName(device, ControlKindButton, controlNumber(b.Metadata)), Type: HomeResourceType(openhue.ResourceIdentifierRtypeButton), Parent: device},
			Kind:      ControlKindButton,
			ControlId: controlNumber(b.Metadata),
			Button:    &hueButton,
		})
	}

	for _, r := range ctx.rotaries {
		if r.Owner.Rid != device.Id {
			continue
		}
		hueRotary := r
		controls = append(controls, Control{
			Resource:  Resource{ctx: ctx, Id: r.Id, Name: controlName(device, ControlKindRotary, controlNumber(r.Metadata)), Type: HomeResourceType(openhue.ResourceIdentifierRtypeRelativeRotary), Parent: device},
			Kind:      ControlKindRotary,
			ControlId: controlNumber(r.Metadata),
			Rotary:    &hueRotary,
		})
	}

	sort.Slice(controls, func(i, j int) bool {
		if controls[i].Kind != controls[j].Kind {
			return controls[i].Kind == ControlKindButton
		}
		if controls[i].ControlId != controls[j].ControlId {
			return controls[i].ControlId < controls[j].ControlId
		}
		return controls[i].Id < controls[j].Id
	})

	return controls
}

func controlNumber(m *struct {
	ControlId *int `json:"control_id,omitempty"`
}) int {
	if m == nil || m.ControlId == nil {
		return 1
	}
	return *m.ControlId
}

func controlName(device *Resource, kind ControlKind, number int) string {
	return fmt.Sprintf("%s %s %d", device.Name, kind, number)
}

func getGroupedLight(ctx *hueHomeCtx, parent *Resource, services []openhue.ResourceIdentifier) (*GroupedLight, error) {

	for _, service := range services {
		rType := HomeResourceType(service.Rtype)
		if rType.is(openhue.ResourceIdentifierRtypeGroupedLight) {
			hueGroupedLight := ctx.groupedLights[service.Rid]
			return &GroupedLight{
				Resource: Resource{
					ctx:    ctx,
					Id:     hueGroupedLight.Id,
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

func getLight(ctx *hueHomeCtx, parent *Resource, services []openhue.ResourceIdentifier) (*Light, error) {

	for _, service := range services {
		rType := HomeResourceType(service.Rtype)
		if rType.is(openhue.ResourceIdentifierRtypeLight) {
			light := ctx.lights[service.Rid]
			return &Light{
				Resource: Resource{
					ctx:    ctx,
					Id:     light.Id,
					Name:   light.Metadata.Name,
					Type:   rType,
					Parent: parent,
				},
				HueData: &light,
			}, nil
		}
	}

	return nil, fmt.Errorf("no 'light' service found for resource %s (%s)", parent.Id, parent.Name)
}
