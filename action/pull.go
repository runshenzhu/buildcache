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
	//"strings"
	"net/http"
	"errors"
	"os"
)

var (
	ErrParse = errors.New("parse parent fail")
)

func Pull(imageName, registryAddr string) error {
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.10"}
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.23", nil, defaultHeaders)
	if err != nil {
		panic(err)
	}

	imageIDs, err := pull(imageName, registryAddr, cli)
	if err == nil {
		return restore(imageIDs, cli)
	} else {
		return err
	}
}

func pull(imageName, registryAddr string, cli *client.Client) ([]string ,error) {
	ctx := context.Background()
	imageIDs := []string{}
	// first pull itself
	auth, _ := EncodeAuthToBase64()
	for{
		if rc, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{RegistryAuth: auth}); err == nil {
			dec := json.NewDecoder(rc)
			m := map[string]interface{}{}
			for {
				if err := dec.Decode(&m); err != nil {
					if err == io.EOF {
						break
					}
					return nil, err
				}
			}
			// if the final stream object contained an error, return nil, it
			if errMsg, ok := m["error"]; ok {
				logrus.Warnln("pull error:", errMsg)
				return imageIDs, nil
			}

			logrus.Debugln("pull:", imageName, m)
		} else {
			logrus.Errorln("push image gets error: ", err)
			return nil, err
		}

		childInfo, _, err := cli.ImageInspectWithRaw(context.Background(), imageName, false)
		if err != nil {
			logrus.Debugln("inspect on child image gets error:", err)
			panic(err)
		}
		childID := childInfo.ID[len("sha256:"):]
		imageIDs = append(imageIDs, childID)
		if parentID, err := getParent(imageName, registryAddr); err == nil {
			logrus.Debugln("parent:", parentID)
			imageName = registryAddr + "/" + parentID
			if err == nil {
				setParent(childID, parentID)			
			} 
		} else {
			return nil, err
		}
	}
}

// query registry to get parent info of a image
func getParent(imageName, registryAddr string) (string, error) {
	// parse image base name, tag
	ref, tag, _ := reference.Parse(imageName)
	token, _ := getToken(ref)

	// construct get url
	url := "https://registry-1.docker.io/v2/" + ref + "/manifests/" + tag
	logrus.Debugln("get url:", url)

	client := http.Client{}

	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Set("Authorization", "Bearer " + token)
	resp, err := client.Do(request)
	if err != nil {
		// handle error
		return "", err
	}
	logrus.Debugln("get parent, resp", resp)
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

		parentID, _ := parentBase.(string)
		if parentID == "" {
			return "", ErrParse
		}
		return parentID, nil
	}
}

const METAPATH = "/var/lib/docker/image/aufs/imagedb/metadata/sha256/"

// set parent relationship among child and parent image
func setParent(child, parent string) error {
	logrus.Debugln("child:", child, "parent:", parent)
	path := METAPATH + child
	logrus.Debugln("path:", path)

	// if parent meta already exit, do nothing
	//if exit, err := isParentMetaExist(path); exit {
	//	// fixme: fix op not permitted error
	//	if err != nil {
	//		logrus.Warnln("check exit get error:", err)
	//	}
	//	return nil
	//}

	// set relationship
	// fixme: perm mode may not be correct
	os.Mkdir(path, 0777)
	fd, err := os.Create(path + "/parent")
	if err != nil {
		return err
	}
	defer fd.Close()

	fd.WriteString(parent)
	return nil
}

// check if image(childID) has meta data of parent
func isParentMetaExist(path string) (bool, error) {

	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return true, err
}

func restore(imageSet []string, cli *client.Client) error {
	logrus.Println(imageSet)
	rc, err := cli.ImageSave(context.Background(), imageSet)
	if err != nil {
		return err
	}
	defer rc.Close()

	_ , err = cli.ImageLoad(context.Background(), rc, true)
	if err != nil {
		return err
	}
	return nil
}
