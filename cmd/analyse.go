package cmd

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"github.com/xeals/signal-back/types"
)

// Analyse fulfils the `analyse` subcommand.
var Analyse = cli.Command{
	Name:               "analyse",
	Usage:              "Information about the backup file",
	UsageText:          "Display statistical information about the backup file.",
	Aliases:            []string{"analyze"},
	CustomHelpTemplate: SubcommandHelp,
	Action: func(c *cli.Context) error {
		bf, err := setup(c)
		if err != nil {
			return err
		}

		a, err := AnalyseTables(bf)
		fmt.Println(a)

		return nil
	},
}

// AnalyseTables calculates the frequency of all records in the backup file.
func AnalyseTables(bf *types.BackupFile) (map[string]int, error) {
	counts := make(map[string]int)

	frames, err := bf.Slurp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to slurp frames")
	}
	for _, f := range frames {
		if f.GetHeader() != nil {
			counts["header"]++
			continue
		}
		if f.GetVersion() != nil {
			counts["version"]++
			continue
		}
		if f.GetAttachment() != nil {
			counts["attachment"]++
			continue
		}
		if f.GetAvatar() != nil {
			counts["avatar"]++
			continue
		}
		if f.GetPreference() != nil {
			counts["pref"]++
			continue
		}
		if stmt := f.GetStatement(); stmt != nil {
			if strings.HasPrefix(*stmt.Statement, "DROP TABLE") {
				if counts["drop_table"] == 0 {
					fmt.Println(stmt)
				}
				counts["drop_table"]++
				continue
			}
			if strings.HasPrefix(*stmt.Statement, "CREATE TABLE") {
				if counts["create_table"] == 0 {
					fmt.Println(stmt)
				}
				counts["create_table"]++
				continue
			}
			if strings.HasPrefix(*stmt.Statement, "DROP INDEX") {
				if counts["drop_index"] == 0 {
					fmt.Println(stmt)
				}
				counts["drop_index"]++
				continue
			}
			if strings.HasPrefix(*stmt.Statement, "CREATE INDEX") ||
				strings.HasPrefix(*stmt.Statement, "CREATE UNIQUE INDEX") {
				if counts["create_index"] == 0 {
					fmt.Println(stmt)
				}
				counts["create_index"]++
				continue
			}
			if strings.HasPrefix(*stmt.Statement, "INSERT INTO") {
				table := strings.Split(*stmt.Statement, " ")[2]
				if counts["insert_into_"+table] == 0 {
					fmt.Println(stmt)
				}
				counts["insert_into_"+table]++
				continue
			}

			counts["other_stmt"]++
		}
	}

	return counts, nil
}
