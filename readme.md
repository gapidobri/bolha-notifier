# bolha-notifier

An app for sending new bolha.com posts to discord

## Configuration

```yaml
# config.yaml

webhook_url: "<discord webhook url>"
check_interval: 300 # check interval in seconds (default 5 minutes)
# List of urls to check
urls:
  - "https://www.bolha.com/search/?keywords=camera"
# Words to exclude when searching
excluded_words:
  - "sony"
```

The app will check for `config.yaml` file in current directory and `/etc/bolha-notifier`.
The location can be overridden by specifying `--config <path to config>` flag.
