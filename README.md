# BUILD CACHE
Make Docker Build Great Again!

## Usage
### Prerequisites
  install docker: https://docs.docker.com/engine/installation/linux/ubuntulinux/
### .bash_profile
`alias dockergreat="docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock -v /var/lib/docker/:/var/lib/docker/ -v $HOME/.docker/config.json:/credentials.json runshenzhujm/buildcache:latest"`
### Commands
`dockergreat --push <reponame> --registry-addr <registry>`

`dockergreat --pull <reponame> --registry-addr <registry>`
