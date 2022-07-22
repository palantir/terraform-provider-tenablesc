package tenablesc

import (
	"fmt"
)

const feedEndpoint = "/feed"

// Feeds represents the output structure for https://docs.tenable.com/tenablesc/api/Feed.htm queries.
// The feed endpoint returns a map of Feed objects with the name as primary key.
type Feeds map[string]*Feed

type Feed struct {
	UpdateTime    ProbablyString
	Stale         FakeBool
	UpdateRunning FakeBool
}

func (c *Client) GetAllFeeds() (Feeds, error) {
	var feeds Feeds

	_, err := c.getResource(feedEndpoint, &feeds)
	if err != nil {
		return nil, err
	}

	return feeds, nil
}

func (c *Client) GetFeed(name string) (*Feed, error) {
	var feeds map[string]*Feed

	_, err := c.getResource(fmt.Sprintf("%s/%s", feedEndpoint, name), &feeds)
	if err != nil {
		return nil, err
	}
	if feed, ok := feeds[name]; ok {
		return feed, nil
	}

	return nil, fmt.Errorf("API response did not include requested feed name")
}

func (c *Client) UpdateFeed(name string) error {
	// in case you're staring at this api and wondering, 'all' is acceptable for this name, and it'll ask for an update
	// to all feeds.
	_, err := c.postResource(fmt.Sprintf("%s/%s/update", feedEndpoint, name), nil, nil)
	if err != nil {
		return err
	}

	return nil
}
