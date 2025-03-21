// Copyright 2024 Meshery Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package relationships

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/layer5io/meshery/mesheryctl/internal/cli/root/config"
	"github.com/layer5io/meshery/mesheryctl/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	searchModelName string
	searchType      string
	searchSubType   string
	searchKind      string
)

// represents the mesheryctl exp relationship search [query-text] subcommand.
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search registered relationship(s)",
	Long:  "Search registred relationship(s) used by different models",
	Example: `
// Search for relationship using a query
mesheryctl exp relationship search [--kind <kind>] [--type <type>] [--subtype <subtype>] [--model <model>] [query-text]`,
	Args: func(cmd *cobra.Command, args []string) error {
		const usage = "mesheryctl exp relationship search [--kind <kind>] [--type <type>] [--subtype <subtype>] [--model <model>]"
		errMsg := fmt.Errorf("[--kind, --subtype or --type or --model] and [query-text] are required\n\nUsage: %s\nRun 'mesheryctl exp relationship search --help'", usage)

		if searchKind == "" && searchSubType == "" && searchType == "" && searchModelName == "" {
			err := utils.ErrInvalidArgument(errMsg)
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		mctlCfg, err := config.GetMesheryCtl(viper.GetViper())
		if err != nil {
			log.Fatalln(err, "error processing config")
		}

		baseUrl := mctlCfg.GetBaseMesheryURL()
		url := buildSearchUrl(baseUrl)
		relationshipResponse, err := fetchRelationships(url)

		if err != nil {
			return err
		}

		rows := [][]string{}

		for _, relationship := range relationshipResponse.Relationships {
			if len(relationship.Type()) > 0 {
				evaluationQuery := ""
				if relationship.EvaluationQuery != nil {
					evaluationQuery = *relationship.EvaluationQuery
				}
				rows = append(rows, []string{string(relationship.Kind), relationship.SchemaVersion, relationship.Model.DisplayName, relationship.SubType, evaluationQuery})
			}
		}

		data := relationshipsData{
			Headers:          []string{"kind", "apiVersion", "model-name", "subType", "regoQuery"},
			Rows:             rows,
			Count:            relationshipResponse.Count,
			DisplayCountOnly: false,
		}
		return listRelationships(cmd, data)
	},
}

func init() {
	searchCmd.Flags().SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		return pflag.NormalizedName(strings.ToLower(name))
	})
	searchCmd.Flags().StringVarP(&searchKind, "kind", "k", "", "search particular kind of relationships")
	searchCmd.Flags().StringVarP(&searchSubType, "subtype", "s", "", "search particular subtype of relationships")
	searchCmd.Flags().StringVarP(&searchModelName, "model", "m", "", "search relationships of particular model name")
	searchCmd.Flags().StringVarP(&searchType, "type", "t", "", "search particular type of relationships")
}

func buildSearchUrl(baseUrl string) string {
	searchUrl := ""
	escapedType := url.QueryEscape(searchType)
	escapeKind := url.QueryEscape(searchKind)
	escapeSubType := url.QueryEscape(searchSubType)
	if searchModelName == "" {
		searchUrl = fmt.Sprintf("%s/api/meshmodels/relationships?type=%s&kind=%s&subType=%s&pagesize=all", baseUrl, escapedType, escapeKind, escapeSubType)

	} else {
		escapeModelName := url.QueryEscape(searchModelName)
		searchUrl = fmt.Sprintf("%s/api/meshmodels/models/%s/relationships?type=%s&kind=%s&subType=%s&pagesize=all", baseUrl, escapeModelName, escapedType, escapeKind, escapeSubType)
	}
	return searchUrl
}
