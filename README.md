# GLab
A custom Gitlab Cli tool written in Go (golang)
Work seamlessly with Gitlab from the command line.

## USAGE
  ```bash
  glab <command> <subcommand> [flags]
  ```

## CORE COMMANDS
  ```bash
  issue:      Create and view issues
  repo:       Create, clone, fork, and view repositories
  mr:         Create, view, approve and merge merge requests
  ```

## ADDITIONAL COMMANDS
  
  ```bash
  config:     Manage configuration for gh
  help:       Help about any command
  ```

## FLAGS
  ```bash
  --help      Show help for command
  --version   Show glab version
  ```

## EXAMPLES
  ```bash
  $ glab issue create
  $ glab issue list --closed
  $ glab repo clone profclems/glab
  $ glab pr checkout 321
  ```

## ENVIRONMENT VARIABLES
  ```bash
  GITLAB_TOKEN: an authentication token for API requests. Setting this avoids being
  prompted to authenticate and overrides any previously stored credentials.

  GITLAB_REPO: specify the Gitlab repository in "OWNER/REPO" format for commands that
  otherwise operate on a local repository.

  GITLAB_URI: specify the url of the gitlab server if self hosted (eg: gitlab.example.com)
  ```
  
## LEARN MORE
  Use "glab <command> <subcommand> --help" for more information about a command.
  Read the manual at https://glab.clementsam.tech

## FEEDBACK
  Open an issue using “glab issue create -R profclems/glab”


Built with ❤ by Clement Sam <https://clementsam.tech>
