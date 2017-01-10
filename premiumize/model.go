package premiumize

type CloudList struct {
	Status  string      `json:"status"`
	Content []CloudItem `json:"content"`
	Message string      `json:"message"`
}

type CloudItem struct {
	ID        string `json:"id"`
	Hash      string `json:"hash"`
	Size      string `json:"size"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	Type      string `json:"type"`
}

type TorrentList struct {
	Status    string        `json:"status"`
	Transfers []TorrentItem `json:"transfers"`
	Message   string        `json:"message"`
}

type TorrentItem struct {
	ID        string  `json:"id"`
	Hash      string  `json:"hash"`
	Status    string  `json:"status"`
	Size      int64   `json:"size"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	SpeedUp   int     `json:"seed_up"`
	SpeedDown int     `json:"speed_down"`
	Seeder    int     `json:"seeder"`
	Leecher   int     `json:"leecher"`
	Ratio     float64 `json:"ratio"`
	ETA       int     `json:"eta"`
	Progress  float64 `json:"progress"`
}

type Torrent struct {
	Size    int64                     `json:"size"`
	Items   int                       `json:"items"`
	Zip     string                    `json:"zip"`
	Status  string                    `json:"status"`
	Content map[string]TorrentContent `json:"content"`
}

type TorrentContent struct {
	Type     string                    `json:"type"`
	Name     string                    `json:"name"`
	Size     int64                     `json:"size"`
	Items    int                       `json:"items"`
	Zip      string                    `json:"zip"`
	Ext      string                    `json:"ext"`
	MimeType string                    `json:"mimetype"`
	URL      string                    `json:"url"`
	Path     string                    `json:"path"`
	Children map[string]TorrentContent `json:"children"`
}

type UploadResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Type    string `json:"type"`
	ID      string `json:"id"`
	Name    string `json:"name"`
}
