package openhue

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/openhue/openhue-go"
)

// FindRoom returns the single room matching the name or the ID.
func FindRoom(home *HomeModel, nameOrId string) (*Room, error) {
	var found []*Room
	for i := range home.Rooms {
		if home.Rooms[i].Id == nameOrId || home.Rooms[i].Name == nameOrId {
			found = append(found, &home.Rooms[i])
		}
	}
	switch len(found) {
	case 0:
		return nil, fmt.Errorf("room '%s' not found", nameOrId)
	case 1:
		return found[0], nil
	default:
		return nil, fmt.Errorf("several rooms are named '%s', use the ID instead", nameOrId)
	}
}

// FindScene returns the single scene (and its room) matching the name or the ID, optionally within a room.
func FindScene(home *HomeModel, roomNameOrId, nameOrId string) (*Scene, *Room, error) {
	var scenes []*Scene
	var rooms []*Room
	for i := range home.Rooms {
		room := &home.Rooms[i]
		if roomNameOrId != "" && !room.matchesNameOrId(roomNameOrId) {
			continue
		}
		for j := range room.Scenes {
			if room.Scenes[j].matchesNameOrId(nameOrId) {
				scenes = append(scenes, &room.Scenes[j])
				rooms = append(rooms, room)
			}
		}
	}
	switch len(scenes) {
	case 0:
		if roomNameOrId != "" {
			return nil, nil, fmt.Errorf("scene '%s' not found in room '%s'", nameOrId, roomNameOrId)
		}
		return nil, nil, fmt.Errorf("scene '%s' not found", nameOrId)
	case 1:
		return scenes[0], rooms[0], nil
	default:
		return nil, nil, fmt.Errorf("several scenes match '%s', use --room or the scene ID", nameOrId)
	}
}

// CreateScene sends the creation request and returns the ID of the new scene.
//
// The bridge answers '201 Created' while openhue-go only accepts '200 OK', so a 201 is reported as an error
// although the scene exists: in that case the scene is looked up by name in the room.
func (r *Room) CreateScene(h *openhue.Home, body openhue.ScenePost) (string, error) {
	ctx := context.Background()
	id, err := h.CreateScene(ctx, body)
	if err == nil {
		return id.Rid, nil
	}

	var apiErr *openhue.ApiError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusCreated {
		return "", err
	}

	scenes, lookupErr := h.GetScenes(ctx)
	if lookupErr != nil {
		return "", fmt.Errorf("scene created but its ID could not be retrieved: %w", lookupErr)
	}
	for _, s := range scenes {
		if s.Group.Rid == r.Id && s.Metadata.Name != nil && body.Metadata.Name != nil && *s.Metadata.Name == *body.Metadata.Name {
			return s.Id, nil
		}
	}
	return "", fmt.Errorf("scene created but not found afterwards")
}

// Update sends the update request.
func (s *Scene) Update(h *openhue.Home, body openhue.ScenePut) error {
	return h.UpdateScene(context.Background(), s.Id, body)
}

// Delete deletes the scene.
func (s *Scene) Delete(h *openhue.Home) error {
	return h.DeleteScene(context.Background(), s.Id)
}

// ReloadScene fetches the current data of the scene identified by its ID.
func ReloadScene(h *openhue.Home, id string) (*openhue.SceneGet, error) {
	return h.GetSceneById(context.Background(), id)
}
