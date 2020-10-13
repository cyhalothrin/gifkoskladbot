package cmd

import (
	"github.com/cyhalothrin/gifkoskladbot/favchannel/publish"
	"github.com/spf13/cobra"
)

var isCommandCollect bool
var isCommandPublish bool

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Posts gif with tags to channel",
	RunE: func(cmd *cobra.Command, args []string) error {
		if isCommandCollect {
			return publish.PublishGifWithTags(publish.CommandCollect)
		}
		if isCommandPublish {
			return publish.PublishGifWithTags(publish.CommandPublish)
		}

		//return errors.New("unknown command")
		publish.TestPublish()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// publishCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// publishCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	publishCmd.Flags().BoolVar(&isCommandCollect, "collect", false, "collects messages")
	publishCmd.Flags().BoolVar(&isCommandPublish, "publish", false, "posts gifs to channel")
}
