package create

import (
	"time"

	"github.com/profclems/glab/commands/cmdutils"
	"github.com/spf13/cobra"
	"github.com/xanzy/go-gitlab"
)

type ReleaseAssets struct {
	URL  string
	Name string
	Type string
}

type CreateOpts struct {
	Name       string
	TagName    string
	Notes      string
	NotesFile  string
	Milestone  []string
	AssetURLs  []string
	ReleasedAt string

	Assets []*ReleaseAssets

	factory *cmdutils.Factory
}

func NewCmdCreate(f *cmdutils.Factory, runE func(opts *CreateOpts) error) *cobra.Command {
	opts := &CreateOpts{
		factory: f,
	}

	cmd := &cobra.Command{
		Use:   "create <tag> [<files>...]",
		Short: "Create a new GitLab Release for a repository",
		Long: `Create a new GitLab Release for a repository.

You need push access to the repository to create a Release.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if runE != nil {
				return runE(opts)
			}

			return createRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Name, "name", "n", "", "The release `name` or title")
	cmd.Flags().StringVarP(&opts.Notes, "notes", "N", "", "The release notes/description. You can use Markdown")
	cmd.Flags().StringVarP(&opts.NotesFile, "notes-file", "F", "", "Read release notes `file`")
	cmd.Flags().StringVarP(&opts.ReleasedAt, "released-at", "D", "", "The `date` when the release is/was ready. Defaults to the current datetime. Expected in ISO 8601 format (2019-03-15T08:00:00Z)")
	cmd.Flags().StringSliceVarP(&opts.Milestone, "milestone", "m", []string{}, "The title of each milestone the release is associated with")
	cmd.Flags().StringSliceVarP(&opts.AssetURLs, "asset-url", "u", []string{}, "Release assets link `URLs`")

	return cmd
}

func createRun(opts *CreateOpts) error {
	createAPIOpts := &gitlab.CreateReleaseOptions{}
	if opts.Name != "" {
		createAPIOpts.Name = gitlab.String(opts.Name)
	}
	if opts.ReleasedAt != "" {
		// Parse the releasedAt to the expected format of the API
		// From the API docs "Expected in ISO 8601 format (2019-03-15T08:00:00Z)".
		t, err := time.Parse("2019-03-15T08:00:00Z", opts.ReleasedAt)
		if err != nil {
			return err
		}
		createAPIOpts.ReleasedAt = gitlab.Time(t)
	}
	return nil
}
