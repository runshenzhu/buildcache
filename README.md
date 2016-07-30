# buildcache
make docker build great again

go: https://github.com/golang/go/wiki/Ubuntu

`$ alias dockergreat="docker run -it --rm -v /var/run/docker.sock:/var/run/docker.sock -v /var/lib/docker/:/var/lib/docker/ -v $HOME/.docker/config.json:/credentials.json runshenzhujm/buildcache:latest"`

`dockergreat --push <reponame> --registry-addr <registry>`

`dockergreat --pull <reponame> --registry-addr <registry>`
