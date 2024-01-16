package get

import (
	"crypto/tls"
	"encoding/json"
	"github.com/r3labs/sse/v2"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"
	"openhue-cli/openhue"
	"openhue-cli/util"
)

type HueEvent struct {
	Id     string                   `json:"id"`
	Events []map[string]interface{} `json:"events"`
}

func NewCmdGetEvents(ctx *openhue.Context, c *CmdGetOptions) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "events",
		Short: "(alpha) Watch events from the bridge",
		Long: `
This command allows to watch events emitted by the Hue Bridge. Events are formatted as JSON objects.
`,
		Run: func(cmd *cobra.Command, args []string) {
			client := createSseClient(ctx)
			err := client.Subscribe("messages", onEvent(ctx))
			cobra.CheckErr(err)
		},
	}

	return cmd
}

func onEvent(ctx *openhue.Context) func(msg *sse.Event) {
	return func(msg *sse.Event) {
		//id := string(msg.ID)
		hueEvent := HueEvent{
			Id: string(msg.ID),
		}
		err := json.Unmarshal(msg.Data, &hueEvent.Events)
		if err != nil {
			log.Warnf("Unable to parse JSON %s", string(msg.Data))
		}

		util.PrintJson(ctx.Io, hueEvent)
	}
}

// createSseClient creates a fully initialized sse.Client to retrieve bridge events
func createSseClient(ctx *openhue.Context) *sse.Client {
	client := sse.NewClient("https://" + ctx.Config.Bridge + "/eventstream/clip/v2")
	client.Connection.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client.Headers["hue-application-key"] = ctx.Config.Key
	return client
}
