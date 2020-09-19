package soundcloud

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
)

var tokenFile = path.Join(os.TempDir(), "soundcloud-token.txt")

func (c *Client) updateToken() error {
	var token string
	resp, err := c.Get("https://soundcloud.com", false)
	if err != nil {
		return fmt.Errorf("can't fetch soundcloud.com: %s", err)
	}
	defer resp.Body.Close()
	respContent, err := ioutil.ReadAll(resp.Body)
	var scriptRE = regexp.MustCompile(`<script[^>]+src="([^"]+)"`)
	var clientRE = regexp.MustCompile(`client_id\s*:\s*"([0-9a-zA-Z]{32})"`)
	scripts := scriptRE.FindAllStringSubmatch(string(respContent), -1)
	for _, script := range scripts {
		scriptURL := script[1]
		scriptBodyResp, err := c.Get(scriptURL, false)
		if err != nil {
			continue
		}
		scriptBodyContent, err := ioutil.ReadAll(scriptBodyResp.Body)
		scriptBodyResp.Body.Close()
		if err != nil {
			continue
		}
		matches := clientRE.FindSubmatch(scriptBodyContent)
		if matches == nil {
			continue
		}
		token = string(matches[1])
		break
	}

	if token == "" {
		return ErrNoToken
	}

	_ = ioutil.WriteFile(tokenFile, []byte(token), 0644)
	c.ClientID = token
	return nil
}

func readTokenFromFile() (string, error) {
	data, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
