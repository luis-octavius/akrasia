package commands

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/luis-octavius/akrasia/pkg/i18n"
)

// rpad adds padding to the right of a string.
func rpad(s string, padding int) string {
	formattedString := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(formattedString, s)
}

func trimRightSpace(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

// useLine puts out the full usage for a given command (including parents).
func useLine(c *cobra.Command) string {
	var useline string
	use := strings.Replace(c.Use, c.Name(), c.DisplayName(), 1)
	if c.HasParent() {
		useline = c.Parent().CommandPath() + " " + use
	} else {
		useline = use
	}
	if c.DisableFlagsInUseLine {
		return useline
	}
	if c.HasAvailableFlags() && !strings.Contains(useline, i18n.T("cobraFlagsString")) {
		useline += " " + i18n.T("cobraFlagsString")
	}
	return useline
}


func UsageFunc(cmd *cobra.Command) error {
	cmd.Print(i18n.T("cobraUsage"))
	if cmd.Runnable() {
		cmd.Printf("\n  %s", useLine(cmd))
	}
	if cmd.HasAvailableSubCommands() {
		cmd.Printf(i18n.T("cobraUsageSubcommand"), cmd.CommandPath())
	}
	if len(cmd.Aliases) > 0 {
		cmd.Printf(i18n.T("cobraAliases"))
		cmd.Printf("  %s", cmd.NameAndAliases())
	}
	if cmd.HasExample() {
		cmd.Printf(i18n.T("cobraExamples"))
		cmd.Printf("%s", cmd.Example)
	}
	if cmd.HasAvailableSubCommands() {
		cmds := cmd.Commands()
		if len(cmd.Groups()) == 0 {
			cmd.Printf(i18n.T("cobraAvailableCommands"))
			for _, subcmd := range cmds {
				if subcmd.IsAvailableCommand() || subcmd.Name() == "help" {
					cmd.Printf("\n  %s %s", rpad(subcmd.Name(), subcmd.NamePadding()), subcmd.Short)
				}
			}
		} else {
			for _, group := range cmd.Groups() {
				cmd.Printf("\n\n%s", group.Title)
				for _, subcmd := range cmds {
					if subcmd.GroupID == group.ID && (subcmd.IsAvailableCommand() || subcmd.Name() == "help") {
						cmd.Printf("\n  %s %s", rpad(subcmd.Name(), subcmd.NamePadding()), subcmd.Short)
					}
				}
			}
			if !cmd.AllChildCommandsHaveGroup() {
				cmd.Printf(i18n.T("cobraAdditionalCommands"))
				for _, subcmd := range cmds {
					if subcmd.GroupID == "" && (subcmd.IsAvailableCommand() || subcmd.Name() == "help") {
						cmd.Printf("\n  %s %s", rpad(subcmd.Name(), subcmd.NamePadding()), subcmd.Short)
					}
				}
			}
		}
	}
	if cmd.HasAvailableLocalFlags() {
		cmd.Printf(i18n.T("cobraFlags"))
		cmd.Print(trimRightSpace(cmd.LocalFlags().FlagUsages()))
	}
	if cmd.HasAvailableInheritedFlags() {
		cmd.Printf(i18n.T("cobraGlobalFlags"))
		cmd.Print(trimRightSpace(cmd.InheritedFlags().FlagUsages()))
	}
	if cmd.HasHelpSubCommands() {
		cmd.Printf(i18n.T("cobraAdditionalHelp"))
		for _, subcmd := range cmd.Commands() {
			if subcmd.IsAdditionalHelpTopicCommand() {
				cmd.Printf("\n  %s %s", rpad(subcmd.CommandPath(), subcmd.CommandPathPadding()), subcmd.Short)
			}
		}
	}
	if cmd.HasAvailableSubCommands() {
		cmd.Printf(i18n.T("cobraHelpInfo"), cmd.CommandPath())
	}
	cmd.Println()
	return nil
	
}
