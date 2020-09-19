package soundcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/camelva/erzo/parsers"
	"io/ioutil"
)

func (c *Client) getMetadata(ctx context.Context, uri string) (*metadata2, error) {
	cleanURL, err := c.tidyURL(uri)
	if err != nil {
		return nil, err
	}

	resolveURL := fmt.Sprintf("https://api-v2.soundcloud.com/resolve?url=%s", cleanURL)

	resp, err := c.httpGet(ctx, resolveURL, true)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respData, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("can't read %s: %s", resolveURL, err.Error())
	}

	// if response = "{}"
	if len(respData) < 3 {
		return nil, fmt.Errorf("no metadata, URL: %s", cleanURL)
	}

	meta := new(metadata2)
	if err := json.Unmarshal(respData, meta); err != nil {
		return nil, parsers.ErrCantContinue(fmt.Sprintf("can't unmarshal metadata: %s", err.Error()))
	}

	// update DownloadURL field
	meta.DownloadURL = fmt.Sprintf("https://api-v2.soundcloud.com/tracks/%d/download", meta.ID)
	return meta, nil
}
