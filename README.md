# StartMyGame (SMG)

Creates a cloud server based on a snapshot and shuts it down after a inactivity.
This program is indented to work with Garrys Mod and DigitalOcean and Hetzner 
as cloud providers.

This software is written in Go and uses vgo (Versioned Go Prototype).

### Does it makes sense for you?

Before you read all the other stuff, it's important to know what I wanted to achieve with 
the application and whether it makes sense for you too run it also.

If you have a small web/mail/etc. server which is running anyways, 
but can't support running a big game, and don't expect the game server 
to run full-time, then this is perfect for you.  
The cloud server only bills when the game is running and there are people playing the game, 
so you can **save a lot of money** in comparision to a server which runs 24/7.

If have a game server which should run full-time, there's no need for this application,
just let your server run.

A practical example for the first case, where this application makes sense:  
You sometimes play with your friends and provide a server.  
By using this application you just press one button, 
when your team wants to play and leave the server when it's over.  
BUT: By using the application you save a lot money 
in comparision to a expensive game server which runs 24/7.

This application fits your needs? Nice  
Go ahead and start with the [setup](https://github.com/LukWebsForge/StartMyGame/wiki).

### Libraries

* [net/http](https://golang.org/pkg/net/http/)
* [rs/cors](https://github.com/rs/cors)
* [james4k/rcon](https://github.com/james4k/rcon)
* [hetznercloud/hcloud-go](https://github.com/hetznercloud/hcloud-go)
* [digitalocean/godo](https://github.com/digitalocean/godo)