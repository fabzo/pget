package premiumize

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const listTorrentsURL = "https://www.premiumize.me/api/transfer/list"
const browseTorrentURL = "https://www.premiumize.me/api/torrent/browse"
const startTorrentURL = "https://www.premiumize.me/api/transfer/create"
const deleteTorrentURL = "https://www.premiumize.me/api/transfer/delete"
const premiumizeErrorStatus = "error"

type Client struct {
	customerID string
	pin        string
	http       *http.Client
	debug      bool
}

func New(customerID string, pin string) *Client {
	httpClient := &http.Client{}

	return &Client{
		customerID: customerID,
		pin:        pin,
		http:       httpClient,
		debug:      false,
	}
}

func (c *Client) SetDebug(enabled bool) {
	c.debug = enabled
}

func (c *Client) ListTorrents() (TorrentList, error) {
	form := url.Values{}
	form.Set("customer_id", c.customerID)
	form.Set("pin", c.pin)

	resp, err := c.http.PostForm(listTorrentsURL, form)
	if err != nil {
		return TorrentList{}, err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return TorrentList{}, err
	}

	if c.debug {
		fmt.Printf("DEBUG ENABLED. DUMPING LIST RESPONSE:\n%s\n", content)
	}

	list := TorrentList{}
	err = json.Unmarshal(content, &list)
	if err != nil {
		return TorrentList{}, err
	}

	return list, nil
}

func (c *Client) FindTorrentByName(name string) (TorrentItem, error) {
	list, err := c.ListTorrents()
	if err != nil {
		panic(err)
	}

	var torrent TorrentItem
	for _, transfer := range list.Transfers {
		if strings.HasPrefix(strings.ToLower(transfer.Name), strings.ToLower(name)) {
			if torrent.Name != "" {
				return TorrentItem{}, fmt.Errorf("Torrent name with prefix '%s' is ambigious", name)
			}
			torrent = transfer
		}
	}
	if torrent.Name == "" {
		return TorrentItem{}, fmt.Errorf("Unable to find torrent with prefix '%s'", name)
	}

	return torrent, nil
}

func (c *Client) BrowseTorrent(hash string) (Torrent, error) {
	form := url.Values{}
	form.Set("customer_id", c.customerID)
	form.Set("pin", c.pin)
	form.Set("hash", hash)

	resp, err := c.http.PostForm(browseTorrentURL, form)
	if err != nil {
		return Torrent{}, err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Torrent{}, err
	}

	if c.debug {
		fmt.Printf("DEBUG ENABLED. DUMPING BROWSE RESPONSE:\n%s\n", content)
	}

	torrent := Torrent{}
	err = json.Unmarshal(content, &torrent)
	if err != nil {
		return Torrent{}, err
	}

	return torrent, nil
}

func (c *Client) DeleteTorrent(id string) (DeleteResponse, error) {
	form := url.Values{}
	form.Set("customer_id", c.customerID)
	form.Set("pin", c.pin)
	form.Set("type", "torrent")
	form.Set("id", id)

	resp, err := c.http.PostForm(deleteTorrentURL, form)
	if err != nil {
		return DeleteResponse{}, err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return DeleteResponse{}, err
	}

	if c.debug {
		fmt.Printf("DEBUG ENABLED. DUMPING DELETE RESPONSE:\n%s\n", content)
	}

	response := DeleteResponse{}
	err = json.Unmarshal(content, &response)
	if err != nil {
		return DeleteResponse{}, err
	}

	return response, nil
}

func (c *Client) UploadTorrentFile(path string) (UploadResponse, error) {
	file, err := os.Open(path)
	if err != nil {
		return UploadResponse{}, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("src", filepath.Base(path))
	if err != nil {
		return UploadResponse{}, err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return UploadResponse{}, err
	}

	writer.WriteField("customer_id", c.customerID)
	writer.WriteField("pin", c.pin)
	writer.WriteField("type", "torrent")

	req, err := http.NewRequest("POST", startTorrentURL, body)
	if err != nil {
		return UploadResponse{}, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.http.Do(req)
	if err != nil {
		return UploadResponse{}, err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return UploadResponse{}, err
	}

	if c.debug {
		fmt.Printf("DEBUG ENABLED. DUMPING UPLOAD RESPONSE:\n%s\n", content)
	}

	response := UploadResponse{}
	err = json.Unmarshal(content, &response)
	if err != nil {
		return UploadResponse{}, err
	}

	if response.Status == premiumizeErrorStatus {
		return UploadResponse{}, fmt.Errorf("%s", response.Message)
	}

	return response, nil
}

func (c *Client) UploadMagnetLink(link string) (UploadResponse, error) {
	form := url.Values{}
	form.Set("customer_id", c.customerID)
	form.Set("pin", c.pin)
	form.Set("type", "torrent")
	form.Set("src", link)

	resp, err := c.http.PostForm(startTorrentURL, form)
	if err != nil {
		return UploadResponse{}, err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return UploadResponse{}, err
	}

	if c.debug {
		fmt.Printf("DEBUG ENABLED. DUMPING UPLOAD RESPONSE:\n%s\n", content)
	}

	response := UploadResponse{}
	err = json.Unmarshal(content, &response)
	if err != nil {
		return UploadResponse{}, err
	}

	if response.Status == premiumizeErrorStatus {
		return UploadResponse{}, fmt.Errorf("%s", response.Message)
	}

	return response, nil
}
