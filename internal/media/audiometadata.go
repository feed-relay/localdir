package media

//go:generate moq --out ./mocks/audiometadata_mock.go --pkg mocks --skip-ensure --with-resets -fmt goimports . AudioMetadata

type Audio struct {
	Metadata AudioMetadata
	Duration int64
}

type AudioMetadata interface {
	Title() string
	Artist() string
	AlbumArtist() string
	Comment() string
	//Picture() *tag.Picture
}
