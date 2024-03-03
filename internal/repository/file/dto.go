package file

type fileDTO struct {
	ID          string `bson:"_id,omitempty"`
	Filename    string `bson:"filename"`
	Size        int64  `bson:"size"`
	ContentType string `bson:"content_type"`
	Extension   string `bson:"extension"`
	Status      string `bson:"status"`
	URL         string `bson:"url"`
}