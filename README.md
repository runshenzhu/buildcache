# BUILD CACHE
Make Docker Build Great Again!

## Usage
### Prerequisites
  * Linux
  * Docker : ```curl -fsSL https://get.docker.com/ | sh && start docker```

### .profile
`alias dockergreat="docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock -v /var/lib/docker/:/var/lib/docker/ -v $HOME/.docker/config.json:/credentials.json runshenzhujm/buildcache:latest"`
### Commands
`dockergreat --push <registry/reponame> --registry-addr <registry>`

`dockergreat --pull <registry/reponame> --registry-addr <registry>`
