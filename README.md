# Homepage

This is a simple web application that serves as a customizable homepage, allowing users to display various services with corresponding icons, descriptions, and URLs. The application reads configuration from a YAML file and serves a web page based on the provided data.

Inspired by [notclickable-jordan/starbase-80](https://github.com/notclickable-jordan/starbase-80)

## Configuration

The application relies on a configuration file in YAML format. By default, it looks for a file named `config.yml` in the root directory, but you can specify a different path using the `CONFIG_PATH` environment variable.

```yaml
title: Homepage
logo: logo.png
icons: /icons

categories:
- name: Management
  services:
  - name: Kafka UI
    url: https://grafana.example.com/
    icon: kafka-ui
- name: Monitoring
  services:
  - name: Grafana
    description: Data visualization service
    url: https://grafana.example.com/
    icon: grafana
  - name: Prometheus
    description: Monitoring system
    url: https://prometheus.example.com/
    icon: prometheus
```

### Custom Icons

You can provide custom icons by specifying a directory path in the `icons` field of your `config.yml`. The application will attempt to load icons from this directory first, falling back to the embedded static icons if necessary.

### Icon Resolution Order

The application resolves icons using the following order:

1. The application first checks the directory specified in the `icons` field of your `config.yml`.
2. If the icon is not found in the `icons`, the application searches the embedded static icons.
3. If the icon is still not found, the application uses a default icon (`no-icon.svg`).

Additionally, you can specify a URL as the icon field value, and the application will use the image directly from that URL.
