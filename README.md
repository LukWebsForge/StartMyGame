# StartMyGame (SMG)

Creates a cloud server based on a snapshot and shuts it down after a inactivity.
This program is indented to work with Garrys Mod and DigitalOcean and Hetzner 
as cloud providers.

To use this software, you have to setup a running gmod server on Hetzner or DigitalOcean
and add let it start automatically. Then you should shutdown the server and save it as a 
snapshot.

This software is written in Go and uses vgo (Versioned Go Prototype).

A guide how to configure (everything is in a single config.json file) will follow soon.

## Configuration


## Libraries

* [net/http](https://golang.org/pkg/net/http/)
* [rs/cors](https://github.com/rs/cors)
* [james4k/rcon](https://github.com/james4k/rcon)
* [hetznercloud/hcloud-go](https://github.com/hetznercloud/hcloud-go)
* [digitalocean/godo](https://github.com/digitalocean/godo)