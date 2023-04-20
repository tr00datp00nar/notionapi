package notionapi

// /api/v3/loadCachedPageChunk request
type loadCachedPageChunkRequest struct {
	Page            loadCachedPageChunkRequestPage `json:"page"`
	ChunkNumber     int                            `json:"chunkNumber"`
	Limit           int                            `json:"limit"`
	Cursor          cursor                         `json:"cursor"`
	VerticalColumns bool                           `json:"verticalColumns"`
}

type loadCachedPageChunkRequestPage struct {
	ID string `json:"id"`
}

type cursor struct {
	Stack [][]stack `json:"stack"`
}

type stack struct {
	ID    string `json:"id"`
	Index int    `json:"index"`
	Table string `json:"table"`
}

// LoadPageChunkResponse is a response to /api/v3/loadPageChunk api
type LoadCachedPageChunkResponse struct {
	RecordMap *RecordMap `json:"recordMap"`
	Cursor    cursor     `json:"cursor"`

	RawJSON map[string]interface{} `json:"-"`
}

// RecordMap contains a collections of blocks, a space, users, and collections.
type RecordMap struct {
	Version         int                `json:"__version__"`
	Activities      map[string]*Record `json:"activity"`
	Blocks          map[string]*Record `json:"block"`
	Spaces          map[string]*Record `json:"space"`
	NotionUsers     map[string]*Record `json:"notion_user"`
	UsersRoot       map[string]*Record `json:"user_root"`
	UserSettings    map[string]*Record `json:"user_setting"`
	Collections     map[string]*Record `json:"collection"`
	CollectionViews map[string]*Record `json:"collection_view"`
	Comments        map[string]*Record `json:"comment"`
	Discussions     map[string]*Record `json:"discussion"`
}

// LoadPageChunk executes a raw API call /api/v3/loadCachedPageChunk
func (c *Client) LoadCachedPageChunk(pageID string, chunkNo int, cur *cursor) (*LoadCachedPageChunkResponse, error) {
	// emulating notion's website api usage: 30 items on first request,
	// 50 on subsequent requests
	limit := 30
	if cur == nil {
		cur = &cursor{
			// to mimic browser api which sends empty array for this argment
			Stack: make([][]stack, 0),
		}
		limit = 50
	}
	req := &loadCachedPageChunkRequest{
		ChunkNumber:     chunkNo,
		Limit:           limit,
		Cursor:          *cur,
		VerticalColumns: false,
	}
	req.Page.ID = pageID
	var rsp LoadCachedPageChunkResponse
	var err error
	apiURL := "/api/v3/loadCachedPageChunk"
	if err = c.doNotionAPI(apiURL, req, &rsp, &rsp.RawJSON); err != nil {
		return nil, err
	}
	if err = ParseRecordMap(rsp.RecordMap); err != nil {
		return nil, err
	}
	return &rsp, nil
}

func ParseRecordMap(recordMap *RecordMap) error {
	for _, r := range recordMap.Activities {
		if err := parseRecord(TableActivity, r); err != nil {
			return err
		}
	}

	for _, r := range recordMap.Blocks {
		if err := parseRecord(TableBlock, r); err != nil {
			return err
		}
	}

	for _, r := range recordMap.Spaces {
		if err := parseRecord(TableSpace, r); err != nil {
			return err
		}
	}

	for _, r := range recordMap.NotionUsers {
		if err := parseRecord(TableNotionUser, r); err != nil {
			return err
		}
	}

	for _, r := range recordMap.UsersRoot {
		if err := parseRecord(TableUserRoot, r); err != nil {
			return err
		}
	}

	for _, r := range recordMap.UserSettings {
		if err := parseRecord(TableUserSettings, r); err != nil {
			return err
		}
	}

	for _, r := range recordMap.CollectionViews {
		if err := parseRecord(TableCollectionView, r); err != nil {
			return err
		}
	}

	for _, r := range recordMap.Collections {
		if err := parseRecord(TableCollection, r); err != nil {
			return err
		}
	}

	for _, r := range recordMap.Discussions {
		if err := parseRecord(TableDiscussion, r); err != nil {
			return err
		}
	}

	for _, r := range recordMap.Comments {
		if err := parseRecord(TableComment, r); err != nil {
			return err
		}
	}

	return nil
}
