package openstack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func GetImageID(name string) (imgUUID string, err error) {
	conn, _ := GetOpenstackConnection("glance", PublicDomain)
	resp, err := OpenstackCall(conn, "GET", "/v1/images", "")
	if err != nil {
		return "", fmt.Errorf("glance GET error: %v", err)
	}
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", fmt.Errorf("glance GET : read body: %v", err)
	}
	var resData map[string][]map[string]interface{}
	if err := json.Unmarshal([]byte(body), &resData); err != nil {
		panic(err)
	}

	for idx := range resData["images"] {
		img := resData["images"][idx]
		imgName := img["name"]
		if imgName == name {
			imgID := img["id"]
			return imgID.(string), nil
		}
	}
	return "", fmt.Errorf("Image not found")
}
