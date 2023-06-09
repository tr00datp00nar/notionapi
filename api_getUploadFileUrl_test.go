package notionapi

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/kjk/common/assert"
)

const (
	getUploadFileURLJSON1 = `
	{
		"url": "https://s3-us-west-2.amazonaws.com/secure.notion-static.com/246e2166-e2d6-4396-82b5-559c723f57f9/test_file.svg",
		"signedGetUrl": "https://s3.us-west-1.amazonaws.com/SignedGetUrl",
		"signedPutUrl": "https://s3.us-west-2.amazonaws.com/SignedPutUrl"
	}
`
)

func isUploadTestEnabled() bool {
	v := os.Getenv("ENABLE_UPLOAD_TEST")
	return v != ""
}

func TestGetUploadFileURLResponse(t *testing.T) {
	// TODO: re-enable test
	if !isUploadTestEnabled() {
		return
	}
	var res GetUploadFileUrlResponse
	err := json.Unmarshal([]byte(getUploadFileURLJSON1), &res)
	assert.NoError(t, err)

	assert.Equal(t, res.URL, "https://s3-us-west-2.amazonaws.com/secure.notion-static.com/246e2166-e2d6-4396-82b5-559c723f57f9/test_file.svg")
	assert.Equal(t, res.SignedGetURL, "https://s3.us-west-1.amazonaws.com/SignedGetUrl")
	assert.Equal(t, res.SignedPutURL, "https://s3.us-west-2.amazonaws.com/SignedPutUrl")

	res.Parse()
	assert.Equal(t, res.FileID, "246e2166-e2d6-4396-82b5-559c723f57f9")
}

func TestUploadFile(t *testing.T) {
	// TODO: re-enable test
	if !isUploadTestEnabled() {
		return
	}

	const injectionPointText = "Graph (Autogenerated - DO NOT EDIT)"

	client := &Client{
		AuthToken: "<AUTH_TOKEN>",
		Logger:    os.Stdout,
	}

	page, err := client.DownloadPage("6b181fb69a7945ed8c5f424bcb34721c")
	assert.NoError(t, err)

	root := page.Root()
	var parent, embeddedBlock *Block
	for _, b := range root.Content {
		if b.Type != BlockToggle {
			continue
		}

		prop := b.GetProperty("title")
		if len(prop) != 1 || prop[0].Text != injectionPointText {
			continue
		}
		parent = b

		if len(b.Content) == 0 {
			break
		}

		assert.Len(t, b.Content, 1)
		assert.Equal(t, b.Content[0].Type, BlockEmbed)
		embeddedBlock = b.Content[0]

		break
	}
	assert.NotEmpty(t, parent)

	// "a485fd92-b373-47e8-a417-f298689e344b"
	userID := page.UserRecords[0].Block.ID

	file, err := os.Open("test_file.svg")
	assert.NoError(t, err)

	fileID, fileURL, err := client.UploadFile(file)
	assert.NoError(t, err)

	var ops []*Operation
	if embeddedBlock == nil {
		embeddedBlock, ops = parent.EmbedUploadedFileOps(client, userID, fileID, fileURL)
		assert.NoError(t, err)

		ops = append(ops, parent.ListAfterContentOp(embeddedBlock.ID, ""))
	} else {
		ops = embeddedBlock.UpdateEmbeddedFileOps(userID, fileID, fileURL)
		assert.NoError(t, err)
	}

	err = client.SubmitTransaction(ops)
	assert.NoError(t, err)

	t.Logf("got newBlock: %#v", embeddedBlock)
}
