package openstack

import (
	"encoding/json"
	"fmt"
	"github.com/goinggo/mapstructure"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

var AdminDomain = "Admin"
var PublicDomain = "Public"

type URLs struct {
	PublicURL string `jpath:"publicURL"`
	AdminURL  string `jpath:"adminURL"`
}

type serviceCatalog struct {
	Name     string `jpath:"name"`
	EndPoint []URLs `jpath:"endpoints"`
}

type accessProps struct {
	AuthToken      string           `jpath:"access.token.id"`
	TenantID       string           `jpath:"access.token.tenant.id"`
	TenantName     string           `jpath:"access.token.tenant.name"`
	ServiceCatalog []serviceCatalog `jpath:"access.serviceCatalog"`
}

type OpenstackConnection struct {
	AuthToken  string
	TenantName string
	TenantID   string
	AccessURL  string
}

var (
	Tenant   string
	User     string
	Password string
	AuthUrl  string
)

func init() {
	LoadAuth()
}

type AuthContainer struct {
	Auth Auth `json:"auth"`
}

type Auth struct {
	PasswordCredentials *PasswordCredentials `json:"passwordCredentials,omitempty"`
	TenantName          string               `json:"tenantName,omitempty"`
}

type PasswordCredentials struct {
	Username string `json:"username"`
	Pwd      string `json:"password"`
}

func getAuthCredentials() Auth {

	return Auth{
		PasswordCredentials: &PasswordCredentials{
			Username: User,
			Pwd:      Password,
		},
		TenantName: Tenant,
	}
}

func OpenstackAuth() (*accessProps, error) {
	client := &http.Client{}
	authCreds := &AuthContainer{Auth: getAuthCredentials()}
	inArgs, _ := json.Marshal(authCreds)
	body := strings.NewReader(string(inArgs))
	path := fmt.Sprintf("%s/tokens", AuthUrl)

	req, err := http.NewRequest("POST", path, body)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	response, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("Failed API call: %d ", response.Status)
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var resData map[string]interface{}

	if err := json.Unmarshal([]byte(data), &resData); err != nil {
		panic(err)
	}
	var props accessProps
	mapstructure.DecodePath(resData, &props)

	return &props, nil
}

func getEndPoint(svcName string, domain string, svcCatalog []serviceCatalog) (string, error) {

	for i := 0; i < len(svcCatalog); i++ {
		if svcCatalog[i].Name == svcName {
			points := svcCatalog[i].EndPoint
			if domain == AdminDomain {

				return points[0].AdminURL, nil
			} else if domain == PublicDomain {

				return points[0].PublicURL, nil
			} else {

				return "", nil
			}
		}
	}
	return "", nil
}

func OpenstackCall(conn *OpenstackConnection, requestType string, uri string, body string) (*http.Response, error) {

	client := &http.Client{}
	path := fmt.Sprintf("%s%s", conn.AccessURL, uri)

	req, err := http.NewRequest(requestType, path, strings.NewReader(body))
	req.Header.Add("X-Auth-Token", conn.AuthToken)
	req.Header.Add("X-Auth-Project-Id", conn.TenantName)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if response.StatusCode >= 400 {
		return nil, fmt.Errorf("Failed API call: %d ", response.Status)
	}

	return response, nil
}

func LoadAuth() error {
	//Load openstack credentials from environment variables
	Tenant = os.Getenv("OS_TENANT_NAME")
	User = os.Getenv("OS_USERNAME")
	Password = os.Getenv("OS_PASSWORD")
	AuthUrl = os.Getenv("OS_AUTH_URL")

	if User != "" && Tenant != "" && Password != "" && AuthUrl != "" {
		fmt.Errorf("Openstack environment varibales are not set: OS_TENANT_NAME,OS_USERNAME,OS_PASSWORD,OS_AUTH_URL")
	}
	return nil
}

func GetOpenstackConnection(svcName string, domain string) (*OpenstackConnection, error) {
	if User == "" || Tenant == "" || Password == "" || AuthUrl == "" {
		return nil, nil
	}
	props, _ := OpenstackAuth()
	accessEndpoint, _ := getEndPoint(svcName, domain, props.ServiceCatalog)

	var connection = OpenstackConnection{
		AuthToken:  props.AuthToken,
		TenantName: props.TenantName,
		TenantID:   props.TenantID,
		AccessURL:  accessEndpoint,
	}
	return &connection, nil
}

func IsAuthenticated() bool {
	return (User != "" && Tenant != "" && Password != "" && AuthUrl != "")
}
