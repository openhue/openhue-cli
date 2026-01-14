package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	op "github.com/openhue/openhue-go"
	"openhue-cli/openhue"
	"openhue-cli/util/color"
)

// registerTools adds all OpenHue tools to the MCP server
func registerTools(s *server.MCPServer, h *op.Home) {
	// List tools
	s.AddTool(listLightsTool(), listLightsHandler(h))
	s.AddTool(listRoomsTool(), listRoomsHandler(h))
	s.AddTool(listScenesTool(), listScenesHandler(h))

	// Control tools
	s.AddTool(setLightTool(), setLightHandler(h))
	s.AddTool(setRoomTool(), setRoomHandler(h))
	s.AddTool(activateSceneTool(), activateSceneHandler(h))
}

//
// List Lights
//

func listLightsTool() mcp.Tool {
	return mcp.NewTool("list_lights",
		mcp.WithDescription("List all available lights in your Hue system. Returns light ID, name, status (on/off), brightness, and room."),
		mcp.WithString("room",
			mcp.Description("Optional: filter lights by room name or ID"),
		),
	)
}

func listLightsHandler(h *op.Home) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		home, err := openhue.LoadHome(h)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to load home: %v", err)), nil
		}

		room := req.GetString("room", "")
		lights := openhue.SearchLights(home, room, nil)

		result := make([]map[string]any, 0, len(lights))
		for _, light := range lights {
			info := map[string]any{
				"id":   light.Id,
				"name": light.Name,
				"on":   light.IsOn(),
			}

			if light.HueData != nil && light.HueData.Dimming != nil && light.HueData.Dimming.Brightness != nil {
				info["brightness"] = *light.HueData.Dimming.Brightness
			}

			if light.Parent != nil && light.Parent.Parent != nil {
				info["room"] = light.Parent.Parent.Name
			}

			result = append(result, info)
		}

		jsonData, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

//
// List Rooms
//

func listRoomsTool() mcp.Tool {
	return mcp.NewTool("list_rooms",
		mcp.WithDescription("List all rooms in your Hue system. Returns room ID, name, status (on/off), and brightness."),
	)
}

func listRoomsHandler(h *op.Home) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		home, err := openhue.LoadHome(h)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to load home: %v", err)), nil
		}

		rooms := openhue.SearchRooms(home, nil)

		result := make([]map[string]any, 0, len(rooms))
		for _, room := range rooms {
			info := map[string]any{
				"id":   room.Id,
				"name": room.Name,
				"on":   room.GroupedLight.IsOn(),
			}

			if room.GroupedLight != nil && room.GroupedLight.HueData != nil &&
				room.GroupedLight.HueData.Dimming != nil && room.GroupedLight.HueData.Dimming.Brightness != nil {
				info["brightness"] = *room.GroupedLight.HueData.Dimming.Brightness
			}

			result = append(result, info)
		}

		jsonData, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

//
// List Scenes
//

func listScenesTool() mcp.Tool {
	return mcp.NewTool("list_scenes",
		mcp.WithDescription("List all scenes in your Hue system. Returns scene ID, name, and the room it belongs to."),
		mcp.WithString("room",
			mcp.Description("Optional: filter scenes by room name or ID"),
		),
	)
}

func listScenesHandler(h *op.Home) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		home, err := openhue.LoadHome(h)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to load home: %v", err)), nil
		}

		room := req.GetString("room", "")
		scenes := openhue.SearchScenes(home, room, nil)

		result := make([]map[string]any, 0, len(scenes))
		for _, scene := range scenes {
			info := map[string]any{
				"id":   scene.Id,
				"name": scene.Name,
			}

			if scene.Parent != nil {
				info["room"] = scene.Parent.Name
			}

			result = append(result, info)
		}

		jsonData, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

//
// Set Light
//

func setLightTool() mcp.Tool {
	return mcp.NewTool("set_light",
		mcp.WithDescription("Control a light: turn on/off, set brightness, color, or temperature."),
		mcp.WithString("light",
			mcp.Required(),
			mcp.Description("Light name or ID to control"),
		),
		mcp.WithBoolean("on",
			mcp.Description("Turn the light on (true) or off (false)"),
		),
		mcp.WithNumber("brightness",
			mcp.Description("Brightness level from 0 to 100"),
		),
		mcp.WithString("color",
			mcp.Description("Color as hex RGB value (e.g., '#FF5500' or 'FF5500')"),
		),
		mcp.WithNumber("temperature",
			mcp.Description("Color temperature in Mirek (153=cold/blue to 500=warm/yellow)"),
		),
		mcp.WithString("room",
			mcp.Description("Optional: specify room to disambiguate lights with the same name"),
		),
	)
}

func setLightHandler(h *op.Home) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		home, err := openhue.LoadHome(h)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to load home: %v", err)), nil
		}

		lightNameOrId, err := req.RequireString("light")
		if err != nil {
			return mcp.NewToolResultError("light parameter is required"), nil
		}

		room := req.GetString("room", "")
		lights := openhue.SearchLights(home, room, []string{lightNameOrId})

		if len(lights) == 0 {
			return mcp.NewToolResultError(fmt.Sprintf("light '%s' not found", lightNameOrId)), nil
		}

		opts := openhue.NewSetLightOptions()

		// Handle on/off - check if parameter was provided via arguments map
		if args := req.GetArguments(); args != nil {
			if _, exists := args["on"]; exists {
				if req.GetBool("on", false) {
					opts.Status = openhue.LightStatusOn
				} else {
					opts.Status = openhue.LightStatusOff
				}
			}

			// Handle brightness
			if _, exists := args["brightness"]; exists {
				brightness := req.GetFloat("brightness", -1)
				if brightness < 0 || brightness > 100 {
					return mcp.NewToolResultError("brightness must be between 0 and 100"), nil
				}
				opts.Brightness = float32(brightness)
			}

			// Handle temperature
			if _, exists := args["temperature"]; exists {
				tempInt := int(req.GetFloat("temperature", -1))
				if tempInt < 153 || tempInt > 500 {
					return mcp.NewToolResultError("temperature must be between 153 and 500"), nil
				}
				opts.Temperature = tempInt
			}
		}

		// Handle color
		if colorHex := req.GetString("color", ""); colorHex != "" {
			rgb, err := color.NewRGBFomHex(colorHex)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("invalid color: %v", err)), nil
			}
			opts.Color = *rgb.ToXY()
		}

		var results []string
		for _, light := range lights {
			if err := light.Set(opts); err != nil {
				results = append(results, fmt.Sprintf("failed to set '%s': %v", light.Name, err))
			} else {
				results = append(results, fmt.Sprintf("'%s' updated successfully", light.Name))
			}
		}

		jsonData, _ := json.Marshal(map[string]any{"results": results})
		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

//
// Set Room
//

func setRoomTool() mcp.Tool {
	return mcp.NewTool("set_room",
		mcp.WithDescription("Control all lights in a room: turn on/off, set brightness, color, or temperature."),
		mcp.WithString("room",
			mcp.Required(),
			mcp.Description("Room name or ID to control"),
		),
		mcp.WithBoolean("on",
			mcp.Description("Turn the room lights on (true) or off (false)"),
		),
		mcp.WithNumber("brightness",
			mcp.Description("Brightness level from 0 to 100"),
		),
		mcp.WithString("color",
			mcp.Description("Color as hex RGB value (e.g., '#FF5500' or 'FF5500')"),
		),
		mcp.WithNumber("temperature",
			mcp.Description("Color temperature in Mirek (153=cold/blue to 500=warm/yellow)"),
		),
	)
}

func setRoomHandler(h *op.Home) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		home, err := openhue.LoadHome(h)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to load home: %v", err)), nil
		}

		roomNameOrId, err := req.RequireString("room")
		if err != nil {
			return mcp.NewToolResultError("room parameter is required"), nil
		}

		rooms := openhue.SearchRooms(home, []string{roomNameOrId})

		if len(rooms) == 0 {
			return mcp.NewToolResultError(fmt.Sprintf("room '%s' not found", roomNameOrId)), nil
		}

		opts := openhue.NewSetLightOptions()

		// Handle on/off - check if parameter was provided via arguments map
		if args := req.GetArguments(); args != nil {
			if _, exists := args["on"]; exists {
				if req.GetBool("on", false) {
					opts.Status = openhue.LightStatusOn
				} else {
					opts.Status = openhue.LightStatusOff
				}
			}

			// Handle brightness
			if _, exists := args["brightness"]; exists {
				brightness := req.GetFloat("brightness", -1)
				if brightness < 0 || brightness > 100 {
					return mcp.NewToolResultError("brightness must be between 0 and 100"), nil
				}
				opts.Brightness = float32(brightness)
			}

			// Handle temperature
			if _, exists := args["temperature"]; exists {
				tempInt := int(req.GetFloat("temperature", -1))
				if tempInt < 153 || tempInt > 500 {
					return mcp.NewToolResultError("temperature must be between 153 and 500"), nil
				}
				opts.Temperature = tempInt
			}
		}

		// Handle color
		if colorHex := req.GetString("color", ""); colorHex != "" {
			rgb, err := color.NewRGBFomHex(colorHex)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("invalid color: %v", err)), nil
			}
			opts.Color = *rgb.ToXY()
		}

		var results []string
		for _, room := range rooms {
			if room.GroupedLight == nil {
				results = append(results, fmt.Sprintf("room '%s' has no grouped light service", room.Name))
				continue
			}
			if err := room.GroupedLight.Set(opts); err != nil {
				results = append(results, fmt.Sprintf("failed to set room '%s': %v", room.Name, err))
			} else {
				results = append(results, fmt.Sprintf("room '%s' updated successfully", room.Name))
			}
		}

		jsonData, _ := json.Marshal(map[string]any{"results": results})
		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

//
// Activate Scene
//

func activateSceneTool() mcp.Tool {
	return mcp.NewTool("activate_scene",
		mcp.WithDescription("Activate a scene to set predefined lighting for a room."),
		mcp.WithString("scene",
			mcp.Required(),
			mcp.Description("Scene name or ID to activate"),
		),
		mcp.WithString("room",
			mcp.Description("Optional: specify room to disambiguate scenes with the same name"),
		),
		mcp.WithString("action",
			mcp.Description("Action type: 'active' (default), 'static', or 'dynamic'"),
		),
	)
}

func activateSceneHandler(h *op.Home) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		home, err := openhue.LoadHome(h)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to load home: %v", err)), nil
		}

		sceneNameOrId, err := req.RequireString("scene")
		if err != nil {
			return mcp.NewToolResultError("scene parameter is required"), nil
		}

		room := req.GetString("room", "")
		scenes := openhue.SearchScenes(home, room, []string{sceneNameOrId})

		if len(scenes) == 0 {
			return mcp.NewToolResultError(fmt.Sprintf("scene '%s' not found", sceneNameOrId)), nil
		}

		// Parse action
		actionStr := req.GetString("action", "active")
		var action op.SceneRecallAction
		switch actionStr {
		case "active":
			action = op.SceneRecallActionActive
		case "static":
			action = op.SceneRecallActionStatic
		case "dynamic":
			action = op.SceneRecallActionDynamicPalette
		default:
			return mcp.NewToolResultError(fmt.Sprintf("invalid action '%s': must be 'active', 'static', or 'dynamic'", actionStr)), nil
		}

		var results []string
		for _, scene := range scenes {
			if err := scene.Activate(action); err != nil {
				results = append(results, fmt.Sprintf("failed to activate '%s': %v", scene.Name, err))
			} else {
				roomName := "unknown"
				if scene.Parent != nil {
					roomName = scene.Parent.Name
				}
				results = append(results, fmt.Sprintf("scene '%s' activated in room '%s'", scene.Name, roomName))
			}
		}

		jsonData, _ := json.Marshal(map[string]any{"results": results})
		return mcp.NewToolResultText(string(jsonData)), nil
	}
}
