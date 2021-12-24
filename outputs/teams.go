package outputs

import (
	"encoding/json"

	"github.com/aquasecurity/postee/data"
	"github.com/aquasecurity/postee/formatting"
	"github.com/aquasecurity/postee/layout"
	"github.com/aquasecurity/postee/log"
	"github.com/aquasecurity/postee/utils"

	msteams "github.com/aquasecurity/postee/teams"
)

const (
	teamsSizeLimit = 18000 // 28 KB is an approximate limit for MS Teams
)

type TeamsOutput struct {
	Name        string
	AquaServer  string
	teamsLayout layout.LayoutProvider
	Webhook     string
}

func (teams *TeamsOutput) GetName() string {
	return teams.Name
}

func (teams *TeamsOutput) CloneSettings() *data.OutputSettings {
	return &data.OutputSettings{
		Name:   teams.Name,
		Url:    teams.Webhook,
		Enable: true,
		Type:   "teams",
	}
}

func (teams *TeamsOutput) Init() error {
	log.Logger.Infof("Starting MS Teams output %q....", teams.Name)
	teams.teamsLayout = new(formatting.HtmlProvider)
	return nil
}

func (teams *TeamsOutput) Send(input map[string]string) error {
	log.Logger.Infof("Sending to MS Teams via %q...", teams.Name)
	utils.Debug("Title for %q: %q\n", teams.Name, input["title"])
	utils.Debug("Url(s) for %q: %q\n", teams.Name, input["url"])
	utils.Debug("Webhook for %q: %q\n", teams.Name, teams.Webhook)
	utils.Debug("Length of Description for %q: %d/%d\n",
		teams.Name, len(input["description"]), teamsSizeLimit)

	var body string
	if len(input["description"]) > teamsSizeLimit {
		utils.Debug("MS Team output will send SHORT message\n")
		body = buildShortMessage(teams.AquaServer, input["url"], teams.teamsLayout)
	} else {
		utils.Debug("MS Team output will send LONG message\n")
		body = input["description"]
	}
	utils.Debug("Message is: %q\n", body)

	escaped, err := escapeJSON(body)
	if err != nil {
		log.Logger.Errorf("Error while escaping payload: %v", err)
		return err
	}

	err = msteams.CreateMessageByWebhook(teams.Webhook, teams.teamsLayout.TitleH2(input["title"])+escaped)

	if err != nil {
		log.Logger.Errorf("TeamsOutput Send Error: %v", err)
		return err
	}

	log.Logger.Infof("Sending to MS Teams via %q was successful!", teams.Name)
	return nil
}

func (teams *TeamsOutput) Terminate() error {
	log.Logger.Infof("MS Teams output %q terminated", teams.Name)
	return nil
}

func (teams *TeamsOutput) GetLayoutProvider() layout.LayoutProvider {
	return teams.teamsLayout
}

func escapeJSON(s string) (string, error) {
	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	// Trim the beginning and trailing " character
	return string(b[1 : len(b)-1]), nil
}
