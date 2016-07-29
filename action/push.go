package buildcache

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"golang.org/x/net/context"
)

func Push(imageName, registryAddr string) error {
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.10"}
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.23", nil, defaultHeaders)
	if err != nil {
		panic(err)
	}

	_, _, err = cli.ImageInspectWithRaw(context.Background(), imageName, false)

	if err != nil {
		logrus.Debugln("inspect on root image gets error:", err)
		return err
	}

	return push(imageName, registryAddr, cli)
}

func push(imageName, registryAddr string, cli *client.Client) error {
	ctx := context.Background()
	var err = error(nil)
	for{
		if err != nil {
			logrus.Errorln("push image gets error: ", err)
		}

		imageInfo, _, err := cli.ImageInspectWithRaw(ctx, imageName, false)
		if err != nil {
			logrus.Debugln("inspect on parent image gets error:", err)
			return err
		}

		// if imageName not begin from registryAddr, tag it
		if !strings.HasPrefix(imageName, registryAddr) {
			imageName = registryAddr + "/" + imageName
			if err = cli.ImageTag(ctx, imageInfo.ID, imageName); err != nil {
				logrus.Debugln("tag image err:", err)
			}
		}

		// first push itself
		// fixme: set RegistryAuth to correct value
		if rc, err := cli.ImagePush(ctx, imageName, types.ImagePushOptions{RegistryAuth: "123"}); err == nil {
			dec := json.NewDecoder(rc)
			m := map[string]interface{}{}
			for {
				if err = dec.Decode(&m); err != nil {
					if err == io.EOF {
						break
					}
					return err
				}
			}
			// if the final stream object contained an error, return it
			if errMsg, ok := m["error"]; ok {
				logrus.Warnln("push error:", errMsg)
				return fmt.Errorf("%v", errMsg)
			}

			logrus.Debugln("push:", imageName)
		} else {
			logrus.Errorln("push image gets error: ", err)
			return err
		}

		// then, push its parent
		// if we get any errors, just give up.
		// Since its parent images may not exist or get deleted

		parent := imageInfo.Parent

		if parent == "" {
			return nil
		} 
		imageName = parent
	}
}
