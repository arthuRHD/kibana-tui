# kibana-tui

An easy way to see logs when your kibana instance is too slow.

## Install

```sh
go install github.com/arthur-richard/kibana-tui@latest
```

## Config

You need to create a file in you XDG config (`~/.config/kibana-tui/config.yaml`)

```yaml
es: https://my-cluster:9200
es_user: elastic
es_pass: changeme
es_cacert: /etc/ssl/my-ca.pem
sample_size: 200
```
