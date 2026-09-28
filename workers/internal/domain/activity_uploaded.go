package domain

type ActivityUploaded struct {
	ActivityID       int    `json:"activityId"`
	UserID           int    `json:"userId"`
	OriginalFilename string `json:"originalFilename"`
	StorageBucket    string `json:"storageBucket"`
	ObjectKey        string `json:"objectKey"`
	ChecksumSHA256   string `json:"checksumSha256"`
	FileSizeBytes    int    `json:"fileSizeBytes"`
	UploadedAt       string `json:"uploadedAt"`
}
