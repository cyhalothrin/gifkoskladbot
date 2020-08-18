package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

type FileMetaStorage struct {
	f          *os.File
	meta       *metaData
	hasChanges bool
}

func NewFileMetaStorage(path string) (*FileMetaStorage, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	var meta *metaData
	if err := json.NewDecoder(f).Decode(meta); err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, err
		}
	}

	return &FileMetaStorage{
		f:    f,
		meta: meta,
	}, nil
}

func (f *FileMetaStorage) GetTags() []string {
	return nil
}

func (f *FileMetaStorage) SetTags(strings []string) {
	f.hasChanges = true
	fmt.Println(strings)
}

func (f *FileMetaStorage) GetTagsAliases() map[string]string {
	return map[string]string{}
}

func (f *FileMetaStorage) SetTagsAliases(m map[string]string) {
	f.hasChanges = true
	fmt.Println(m)
}

func (f *FileMetaStorage) GetTagsListMessageID() int {
	return f.meta.TagsListMessageID
}

func (f *FileMetaStorage) Close() {
	defer f.f.Close()

	if !f.hasChanges {
		return
	}

	data, err := json.Marshal(f.meta)
	if err != nil {
		panic(err)
	}

	if _, err := f.f.Write(data); err != nil {
		panic(err)
	}
}

func (f *FileMetaStorage) GetSentAnimations() map[string]*SentAnimation {
	return f.meta.Messages
}

func (f *FileMetaStorage) AddSentAnimations(messages map[string]*SentAnimation) {
	if f.meta.Messages == nil {
		f.meta.Messages = messages

		return
	}

	for key, msg := range messages {
		f.meta.Messages[key] = msg
	}
}

type metaData struct {
	Tags              []string
	TagsAliases       map[string]string
	TagsListMessageID int
	// Messages все отправленные ранее сообщения для редактирования
	Messages map[string]*SentAnimation
}

type SentAnimation struct {
	MessageID int
	FileID    string
	Tags      []string
}
