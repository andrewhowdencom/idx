# Configuration

`idx` supports configuration via file, environment variables, and command-line flags. 

The configuration precedence is (from highest to lowest):
1. Command-line flags
2. Environment variables
3. Configuration file
4. Defaults

## Configuration File

By default, `idx` follows the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html) and will look for its configuration file at:
`$XDG_CONFIG_HOME/idx/config.yaml` 

If `$XDG_CONFIG_HOME` is not set, it typically defaults to `~/.config/idx/config.yaml`.
It will also search in any directories specified by `$XDG_CONFIG_DIRS`.

You can explicitly override the configuration file path using the `--config` flag.

## Environment Variables

All configuration options can be defined via environment variables using the `IDX_` prefix. 
Dots (`.`) and hyphens (`-`) in flag names are converted to underscores (`_`) in the environment variable.

### Examples

| CLI Flag | Environment Variable | 
| :--- | :--- | 
| `--dir` | `IDX_DIR` |
| `--ollama-host` | `IDX_OLLAMA_HOST` |
| `--ollama-model` | `IDX_OLLAMA_MODEL` |
| `--http.address` | `IDX_HTTP_ADDRESS` |
| `--log-level` | `IDX_LOG_LEVEL` |
