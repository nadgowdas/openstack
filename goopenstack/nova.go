package openstack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func GetFlavorID(name string) (flavorID string, err error) {
	conn, _ := GetOpenstackConnection("nova", PublicDomain)
	resp, err := OpenstackCall(conn, "GET", "/flavors", "")
	if err != nil {
		return "", fmt.Errorf("nova flavor GET error: %v", err)
	}
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", fmt.Errorf("nova flavor GET : read body: %v", err)
	}
	var resData map[string][]map[string]interface{}
	if err := json.Unmarshal([]byte(body), &resData); err != nil {
		panic(err)
	}

	for idx := range resData["flavors"] {
		flavorProp := resData["flavors"][idx]
		flavorName := flavorProp["name"]
		if flavorName == name {
			flavorID := flavorProp["id"]
			return flavorID.(string), nil
		}
	}
	return "", fmt.Errorf("Flavor not found")
}
