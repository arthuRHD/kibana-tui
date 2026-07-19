# kibana-tui

An easy way to see logs when your kibana instance is too slow.

## Install

```sh
go install github.com/arthuRHD/kibana-tui@latest
```

## Config

Configuration is loaded from `config.yaml` in one of these locations (first found wins):
- `$XDG_CONFIG_HOME/kibana-tui/` (falls back to `~/.config/kibana-tui/`)
- `/etc/kibana-tui/`
- current directory

All values can also be set via environment variables prefixed with `KIBANA_TUI_` (e.g. `KIBANA_TUI_ES`).

### Data source (choose one)

| Key | Env | Type | Default | Description |
|-----|-----|------|---------|-------------|
| `file` | `KIBANA_TUI_FILE` | string | – | Path to a JSON log file (one JSON object per line) |
| `es` | `KIBANA_TUI_ES` | string | – | Elasticsearch address (e.g. `https://localhost:9200`) |
| *(none)* | – | – | – | If neither `file` nor `es` is set, sample data is generated |

### Elasticsearch

| Key | Env | Type | Default | Description |
|-----|-----|------|---------|-------------|
| `es` | `KIBANA_TUI_ES` | string | – | Elasticsearch base URL |
| `es_index` | `KIBANA_TUI_ES_INDEX` | string | `logs-*` | Index pattern |
| `es_user` | `KIBANA_TUI_ES_USER` | string | – | Username for basic auth |
| `es_pass` | `KIBANA_TUI_ES_PASS` | string | – | Password for basic auth |
| `es_apikey` | `KIBANA_TUI_ES_APIKEY` | string | – | Base64-encoded API key (overrides user/pass) |
| `es_cacert` | `KIBANA_TUI_ES_CACERT` | string | – | Path to CA certificate for TLS |
| `es_insecure` | `KIBANA_TUI_ES_INSECURE` | bool | `false` | Skip TLS certificate verification |

### Queries

| Key | Env | Type | Default | Description |
|-----|-----|------|---------|-------------|
| `es_default_esql_query` | `KIBANA_TUI_ES_DEFAULT_ESQL_QUERY` | string | – | ES|QL query to pre-fill on startup |
| `sample_size` | `KIBANA_TUI_SAMPLE_SIZE` | int | `500` | Number of documents to fetch from ES or generate as sample |

### Example

```yaml
es: https://my-cluster:9200
es_user: elastic
es_pass: changeme
es_cacert: /etc/ssl/my-ca.pem
sample_size: 200
es_default_esql_query: "FROM logs-* | WHERE @timestamp > NOW() - 1 HOUR | LIMIT 100"
```
