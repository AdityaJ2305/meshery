package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/layer5io/meshery/mesheryctl/internal/cli/root/config"
	"github.com/layer5io/meshery/mesheryctl/pkg/utils"
	"github.com/layer5io/meshery/server/models"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	availableSubcommands []*cobra.Command
	file                 string
	validSourceTypes     []string
)

// AppCmd represents the root command for app commands
var AppCmd = &cobra.Command{
	Use:   "app",
	Short: "Service Mesh Apps Management",
	Long:  `Manage all apps operations; import, list, view, onboard and offboard`,
	Example: `
// Base command
mesheryctl app [subcommand]
	`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if ok := utils.IsValidSubcommand(availableSubcommands, args[0]); !ok {
			return errors.New(utils.SystemError(fmt.Sprintf("invalid command: \"%s\"", args[0])))
		}
		return nil
	},
}

func init() {
	AppCmd.PersistentFlags().StringVarP(&utils.TokenFlag, "token", "t", "", "Path to token file default from current context")

	availableSubcommands = []*cobra.Command{onboardCmd, viewCmd, offboardCmd, listCmd, importCmd}
	AppCmd.AddCommand(availableSubcommands...)
}

func getSourceTypes() error {
	mctlCfg, err := config.GetMesheryCtl(viper.GetViper())
	if err != nil {
		return err
	}
	validTypesURL := mctlCfg.GetBaseMesheryURL() + "/api/application/types"
	client := &http.Client{}
	req, err := utils.NewRequest("GET", validTypesURL, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.Errorf("Response Status Code %d, possible Server Error", resp.StatusCode)
	}
	defer resp.Body.Close()

	var response []*models.ApplicationSourceTypesAPIResponse

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, utils.PerfError("failed to read response body"))
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal response body")
	}

	for _, apiResponse := range response {
		validSourceTypes = append(validSourceTypes, apiResponse.ApplicationType)
	}

	return nil
}

func isValidSource(sType string) bool {
	for _, validType := range validSourceTypes {
		if validType == sType {
			return true
		}
	}

	return false
}
