package buildcache

import (
	"encoding/json"
	"github.com/docker/engine-api/types"
	"github.com/Sirupsen/logrus"
	"encoding/base64"
	"net/http"
	"io"
	"fmt"
)

func EncodeAuthToBase64() (string, error) {
	authConfig := types.AuthConfig{
		Username: "runshenzhujm",
		Password: "runshenzhujm",
	}
	buf, err := json.Marshal(authConfig)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(buf), nil
}

func getToken(repo string) (string, error) {
	auth, _ := EncodeAuthToBase64()
	// url := "https://index.docker.io/v2/runshenzhu/test/manifests/latest"
	url := "https://auth.docker.io/token?service=registry.docker.io&scope=repository:"+repo+":pull,push"
	logrus.Debugln("get url:", url, auth)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	// req.Header.Set("Authorization", "Basic " + auth)
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	dec := json.NewDecoder(res.Body)
	m := map[string]interface{}{}
	for {
		if err := dec.Decode(&m); err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
	}
	// if the final stream object contained an error, return nil, it
	if errMsg, ok := m["error"]; ok {
		logrus.Warnln("pull error:", errMsg)
		return "", fmt.Errorf("%v", errMsg)
	}

	token, _ := m["token"].(string)
	return token, nil
}
