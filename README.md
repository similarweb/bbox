
# BBOX

`bbox` is a handy CLI tool for working with TeamCity. Whether you need to kick off builds, manage the build queue, or clean up unused resources, bbox makes these tasks simpler and more efficient. It's designed to be versatile and user-friendly, helping you streamline your continuous integration workflows.





* [Installation](#installation)
* [Usage](#usage)
* [Flags](#global-flags)
* [Commands](#commands)
* [Contributing](#contributing)
* [License](#license)
* [Acknowledgements](#acknowledgements)
* [Authors](#authors)

<div style="text-align: center;">
    <img src="./assets/logo.svg" height="500">
</div>

## Installation

### Binary
You can download a binary for Linux or OS X on the [GitHub releases page](https://github.com/similarweb/bbox/releases). You
can use `curl` or `wget` to download it. Don't forget to `chmod +x` the file!
### Docker

Pull `bbox` from the Docker repository:

```bash
docker pull similarweb/bbox
```

Or run `bbox` from source:

```bash
git clone https://github.com/similarweb/bbox.git
```
## Usage

If you are within the cloned repository, execute bbox using:

```bash
go run bbox [command] [Flags]
```

For Docker users, run:

```bash
docker run -it bbox [command] [Flags]
```
* Note: The -it flag is crucial when running bbox with Docker. It ensures that the container runs interactively, allowing you to provide input and receive output in real-time.


## Global Flags

Global flags can be used with any bbox command to control its behavior:

| Flags                  | Description                                                           | 
|-------------------------------|-----------------------------------------------------------------------|
| `-h, --help`                  | Display help for Bbox or any specific command. Use `bbox -h` for general help and `bbox [command] -h` for command-specific help.                                                         |
| `-l, --log-level string`      | Log level (debug, info, warn, error, fatal, panic) (default "info")   |
| `--teamcity-url string`       | Teamcity URL (default "https://teamcity.similarweb.io/")              |
| `--teamcity-username string`  | Teamcity username                                                     |
| `--teamcity-password string`  | Teamcity password                                                     |



## Commands


### Trigger Command

The trigger command is used to trigger a single TeamCity build. It allows you to specify various parameters such as the build type, branch name, and properties.

#### Usage

`go run bbox trigger [flags]`



#### Trigger Flags

| Flags                         | Description                                       |
|-------------------------------|---------------------------------------------------|
| `--artifacts-path string`     | Path to download artifacts to (default "./")      |
| `-b, --branch-name string`    | The branch name (default "master")                |
| `-i, --build-type-id string`  | The build type                                    |
| `-d, --download-artifacts`    | Download artifacts                                |
| `-p, --properties stringToString` | The properties in key=value format (default []) |
| `--require-artifacts`         | If downloadArtifacts is true, and no artifacts found, return an error |
| `-w, --wait-for-build`        | Wait for build to finish and get status           |
| `-t, --wait-timeout duration` | Timeout for waiting for build to finish (default 15m0s) |


#### Example

```bash
go run main.go trigger \
    --teamcity-username "<Username>" \
    --teamcity-password '<Password>' \
    --build-type-id "<BuildIDType>" \
    --properties "key1=value1,key2=value2"
```
### Multi-Trigger Command

The multi-trigger command is used to trigger multiple TeamCity builds simultaneously. It accepts a combination of build parameters, allowing for more complex and automated build processes.

#### Usage

`go run bbox multi-trigger [flags]`

#### Multi-Trigger Flags

| Flags| Description|
|------|------------|
| `--artifacts-path string`| Path to download artifacts to (default "./")|
| `-c, --build-params-combination strings` | Combinations as 'buildTypeID;branchName;downloadArtifactsBool;key1=value1&key2=value2' format. Repeatable. Example: 'byBuildId;master;true;key=value&key2=value2' |
| `--require-artifacts`| If downloadArtifactsBool is true, and no artifacts found, return an error|
| `-w, --wait-for-builds`| Wait for builds to finish and get status (default true)|
| `-t, --wait-timeout duration`| Timeout for waiting for builds to finish, default is 15 minutes (default 15m0s)|

#### Example

```bash
go run main.go multi-trigger \
    --teamcity-username "<Username>" \
    --teamcity-password '<Password>' \
    --build-params-combination "<BuildIDType>;<Branch>;<Properties key=value,key=value>"
```


### Clean Command

The `clean` command is used to remove unused or unwanted resources in a TeamCity server environment. This command helps in maintaining a clean and efficient CI environment.

#### Usage

`go run bbox clean [command] [flags]`

#### Available Sub-Commands

- `queue` Clear the TeamCity Build Queue
- `vcs` Delete all unused VCS Roots

### Clean Queue

This sub-command clears the build queue in a TeamCity server environment. It identifies all queued builds and removes them from the queue, ensuring that no pending builds remain.

##### Usage

`go run bbox clean queue [flags]`

### Clean Vcs Roots

This sub-command identifies, lists, and deletes all unused VCS Roots in a TeamCity server environment. **An unused VCS Root is defined as a VCS Root that is neither linked to any build configurations nor included in any build templates**. This helps in keeping the TeamCity environment clean and free of unnecessary resources.

##### Usage

`go run bbox clean vcs [flags]`

#### Clean Vcs Flags

| Flags| Description|
|------|------------|
| `-c, --confirm`| Automatically confirm to delete all unused VCS Roots without prompting the user|

#### Example

```bash
go run main.go clean vcs  \
    --teamcity-username "<Username>" \
    --teamcity-password '<Password>' \
    --confirm
```

### Completion Command

The `completion` command generates the autocompletion script for `bbox` for the specified shell. Autocompletion scripts help to improve the user experience by providing command and flag suggestions as you type. See each sub-command's help for details on how to use the generated script.

#### Usage

`go run bbox completion [command] [flags]`

| Flags| Description|
|------|------------|
|`--no-descriptions` | disable completion descriptions|

#### Available Sub-Commands

- `bash` Generate the autocompletion script for Bash
- `fish` Generate the autocompletion script for Fish
- `powershell` Generate the autocompletion script for PowerShell
- `zsh` Generate the autocompletion script for Zsh


### Bash

Generate the autocompletion script for the Bash shell. This script depends on the `bash-completion` package. If it is not installed already, you can install it via your OS's package manager.

#### Usage

`go run bbox completion bash [flags]`

#### Loading Completions

To load completions in your current shell session:

```bash
source <(bbox completion bash)
```

To load completions for every new session, execute once:

##### Linux:

	bbox completion bash > /etc/bash_completion.d/bbox

##### macOS:

	bbox completion bash > $(brew --prefix)/etc/bash_completion.d/bbox

* Note: You will need to start a new shell for this setup to take effect.

### Fish

Generate the autocompletion script for the Fish shell.

#### Usage

`go run bbox completion fish [flags]`

#### Loading Completions

To load completions in your current shell session:

```bash
bbox completion fish | source
```

To load completions for every new session, execute once:

    bbox completion fish > ~/.config/fish/completions/bbox.fish

* Note: You will need to start a new shell for this setup to take effect.

### Powershell

Generate the autocompletion script for powershell.

#### Usage

`go run bbox completion fish [flags]`

#### Loading Completions

To load completions in your current shell session:

```bash
bbox completion powershell | Out-String | Invoke-Expression
```
To load completions for every new session, add the output of the above command
to your powershell profile.

### Zsh

Generate the autocompletion script for the Zsh shell.

#### Usage

`go run bbox completion zsh [flags]`

If shell completion is not already enabled in your environment, you will need to enable it. You can execute the following once:

```bash
echo "autoload -U compinit; compinit" >> ~/.zshrc
```

#### Loading Completions

To load completions in your current shell session:

```bash
source <(bbox completion zsh)
```

To load completions for every new session, execute once:

##### Linux:

	bbox completion zsh > "${fpath[1]}/_bbox"

##### macOS:

	bbox completion zsh > $(brew --prefix)/share/zsh/site-functions/_bbox

* Note: You will need to start a new shell for this setup to take effect.
### Version Command

Print the version number of bbox

#### Usage

`bbox version [flags]`
## Contributing

Contributions are always welcome!

See `contributing.md` for ways to get started.

Please adhere to this project's `code of conduct`.


## License

[MIT](https://choosealicense.com/licenses/mit/)


## Acknowledgements

 - [Awesome Readme Templates](https://awesomeopensource.com/project/elangosundar/awesome-README-templates)
 - [Awesome README](https://github.com/matiassingers/awesome-readme)
 - [How to write a Good readme](https://bulldogjob.com/news/449-how-to-write-a-good-readme-for-your-github-project)


## Authors

- [@cregev](https://www.github.com/cregev)
- [@MorErel](https://www.github.com/MorErel)
- [@OzBena](https://www.github.com/OzBena)


