# Homepage

A lightweight and customizable web application that serves as a dynamic homepage for organizing and accessing your services. Built with Go, it provides a clean and modern interface to display your services with icons, descriptions, and URLs.

Inspired by [notclickable-jordan/starbase-80](https://github.com/notclickable-jordan/starbase-80)

## Features

- Dynamic configuration loading and hot-reload support
- Clean and responsive web interface
- Custom icons support
- JSON logging
- Docker labels support

## Configuration

### Environment Variables

- `HOMEPAGE_PORT`: HTTP server port
- `HOMEPAGE_LOGLEVEL`: Logging level (default: `info`)
- `HOMEPAGE_TITLE`: Page title
- `HOMEPAGE_LOGO`: Logo icon
- `HOMEPAGE_ICONSDIR`: Custom icons directory
- `HOMEPAGE_PROVIDERS_FILE_FILENAME`: Load dynamic configuration from a file (default: `""`)
- `HOMEPAGE_PROVIDERS_FILE_WATCH`: Watch provider (default: `true`)
- `HOMEPAGE_PROVIDERS_DOCKER`: Enable Docker backend with default settings (default: `false`)
- `HOMEPAGE_PROVIDERS_DOCKER_ENDPOINT`: Docker server endpoint (default: `unix:///var/run/docker.sock`)
- `HOMEPAGE_PROVIDERS_DOCKER_WATCH`: Watch Docker events (default: `true`)

### Providers

#### File

Dynamic configuration:

```yaml
Entertainment:
  Instagram:
    url: https://www.instagram.com/
    icon: sh-instagram
  Spotify:
    url: https://open.spotify.com/
    icon: https://cdn.jsdelivr.net/gh/walkxcode/dashboard-icons/png/spotify.png

Monitoring:
  Grafana:
    description: Data visualization service
    url: https://grafana.example.com/
    icon: grafana.png
  Prometheus:
    description: Monitoring system
    url: https://prometheus.example.com/
    icon: prometheus.webp
```

#### Docker

Dynamic configuration with Docker Labels:

```yml
services:
  web:
    image: nginx:latest
    labels:
    - homepage.group=Web Services
    - homepage.service=Nginx
    - homepage.url=https://nginx.org/
    - homepage.icon=sh-nginx
    - homepage.description=HTTP web server
```

### Icon Resolution

Icons are resolved in the following order:

1. If path is empty, returns default icon (`/static/icons/no-icon.svg`)
2. If path starts with `http` or `/`, uses the path as is
3. For other paths:
   - If path starts with `sh-`, fetches icon from [selfhst/icons](https://github.com/selfhst/icons) CDN
   - Otherwise, fetches icon from [homarr-labs/dashboard-icons](https://github.com/homarr-labs/dashboard-icons) CDN

Note: If file extension is not provided in the icon path, `.png` will be used as default.
