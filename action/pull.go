package buildcache
import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/Sirupsen/logrus"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"golang.org/x/net/context"
	"github.com/docker/engine-api/types/reference"
	"strings"
	"net/http"
	"errors"
)

var (
	ErrParse = errors.New("parse parent fail")
)

func Pull(imageName, registryAddr string) error {
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.10"}
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.24", nil, defaultHeaders)
	if err != nil {
		panic(err)
	}

	return pull(imageName, registryAddr, cli)
}

func pull(imageName, registryAddr string, cli *client.Client) error {
	ctx := context.Background()

	// first pull itself
	// fixme: set RegistryAuth to correct value
	if rc, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{RegistryAuth: "123"}); err != nil {
		logrus.Errorln("push image gets error: ", err)
		return err
	} else {
		dec := json.NewDecoder(rc)
		m := map[string]interface{}{}
		for {
			if err := dec.Decode(&m); err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
		}
		// if the final stream object contained an error, return it
		if errMsg, ok := m["error"]; ok {
			logrus.Warnln("pull error:", errMsg)
			return fmt.Errorf("%v", errMsg)
		}

		logrus.Debugln("pull:", imageName, m)
	}

	if parent, err := getParent(imageName, registryAddr); err == nil {
		logrus.Debugln("parent:", parent)
		pull(parent, registryAddr, cli)
	} else {
		logrus.Warnln("get parent fail:", err)
	}
	return nil
}

func getParent(imageName, registryAddr string) (string, error) {
	// parse image base name, tag
	ref, tag, _ := reference.Parse(imageName)
	if strings.HasPrefix(ref, registryAddr) {
		ref = imageName[len(registryAddr) + 1 : len(ref)]
	}

	// construct get url
	url := "http://" + registryAddr + "/v2/" + ref + "/manifests/" + tag
	logrus.Debugln("get url:", url)

	resp, err := http.Get(url)
	if err != nil {
		// handle error
		return "", err
	}
	rc := resp.Body
	defer rc.Close()

	// parse response, store it into a map
	m := map[string]interface{}{}
	dec := json.NewDecoder(rc)
	for {
		if err := dec.Decode(&m); err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
	}

	if errMsg, ok := m["error"]; ok {
		return "", fmt.Errorf("%v", errMsg)
	}

	// inspect history section to get parent id
	// fixme: this convert workflow is tooooo ugly
	if history, ok := m["history"].([]interface{}); !ok {
		return "", ErrParse
	} else if len(history) == 0 {
		return "", fmt.Errorf("empty history")
	} else {
		v1, _ := history[0].(map[string]interface{})
		var content map[string]interface{}
		jString, _ := v1["v1Compatibility"].(string)
		err = json.Unmarshal([]byte(jString), &content)
		if err != nil {
			return "", err
		}

		config, ok := content["config"]
		if !ok {
			return "", ErrParse
		}

		configMap, _ := config.(map[string]interface{})
		parentBase, ok := configMap["Image"]
		if !ok {
			return "", ErrParse
		}

		parentBaseString, _ := parentBase.(string)
		if parentBaseString == "" {
			return "", ErrParse
		}

		parent := fmt.Sprintf("%v/%v", registryAddr, parentBase)
		return parent, nil
	}
}
