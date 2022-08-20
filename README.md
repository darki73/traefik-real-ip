# Traefik Real IP Plugin

Traefik middleware used to obtain the real IP of the client.

![Github Actions](https://img.shields.io/github/workflow/status/darki73/traefik-real-ip/Build?style=flat-square)
![Go Report](https://goreportcard.com/badge/github.com/darki73/traefik-real-ip?style=flat-square)
![Go Version](https://img.shields.io/github/go-mod/go-version/darki73/traefik-real-ip?style=flat-square)
![Latest Release](https://img.shields.io/github/release/darki73/traefik-real-ip/all.svg?style=flat-square)

- [Traefik Real IP Plugin](#traefik-real-ip-plugin)
    - [Features](#features)
    - [Usage](#usage)
        - [Plugin Installation](#plugin-installation)
        - [Plugin Configuration](#plugin-configuration)
## Features
- Supports multiple providers
  - **Generic** - uses `X-Real-Ip` and `X-Forwarded-For` headers to determine the real IP
  - **Cloudflare** - uses `True-Client-IP` and `CF-Connecting-IP` headers to determine the real IP
  - **Qrator** - uses `X-Qrator-IP-Source` header to determine the real IP
- Allows to specify `excluded networks` and `excluded addresses`
- You can specify which providers to use (default is all, always uses generic provider as fallback)
- You can set preferred provider, which is used to determine the real IP even if other providers also provide the real IP (default is generic)

## Usage
### Plugin Installation
To install plugin, you need to have the following:
1. Experimental feature enabled
2. Add plugin to the `plugins` section of the experimental configuration

```yaml
experimental:
  plugins:
    traefik-real-ip:
      moduleName: "github.com/darki73/traefik-real-ip"
      version: "__CURRENT_PLUGIN_VERSION__"
```

In order to find the latest version (`__CURRENT_PLUGIN_VERSION__`) of the plugin, you can either:
1. Go to [Releases](https://github.com/darki73/traefik-real-ip/releases) page on the Github
2. Go to [Traefik Plugins](https://plugins.traefik.io/) storefront and find the latest version there

### Plugin Configuration
First of all, you need to create a middleware, which will be used in conjunction with the plugin:
```yaml
http:
  middlewares:
    real-ip:
      plugin:
        traefik-real-ip:
            excludedNetworks: []
            excludedAddresses: []
            providers: []
            preferredProvider: ""
```

**excludedNetworks** - list of networks to exclude from the real IP determination  
**excludedAddresses** - list of addresses to exclude from the real IP determination  
**providers** - list of providers to use for the real IP determination  
**preferredProvider** - preferred provider to use for the real IP determination  

All of those options can be left unspecified, in which case the plugin will use the default values.

After middleware is created, you can add it to your router configuration:
```yaml
http:
  services:
    whoami:
      loadBalancer:
        servers:
          - url: "http://localhost:5000"   
  routers:
    whoami:
      rule: Path(`/whoami`)
      service: whoami
      entryPoints:
        - http
      middlewares:
        - real-ip
```