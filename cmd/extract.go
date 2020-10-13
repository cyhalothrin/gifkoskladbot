package cmd

import (
	"github.com/spf13/cobra"

	"github.com/cyhalothrin/gifkoskladbot/favchannel/extractor"
)

// extractCmd represents the extract command
var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extracts gifs from user channel",
	RunE: func(cmd *cobra.Command, args []string) error {
		return extractor.ForwardWithEmptyCaption()
	},
}

func init() {
	rootCmd.AddCommand(extractCmd)
}
