package storage

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
)

type FileMetaStorage struct {
	filename   string
	meta       *metaData
	hasChanges bool
}

func NewFileMetaStorage(path string) (*FileMetaStorage, error) {
	f, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	meta := &metaData{}
	if err := json.NewDecoder(f).Decode(meta); err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, err
		}
	}

	return &FileMetaStorage{
		filename: path,
		meta:     meta,
	}, nil
}

func (f *FileMetaStorage) GetTags() []string {
	return f.meta.Tags
}

func (f *FileMetaStorage) SetTags(tags []string) {
	f.hasChanges = true
	f.meta.Tags = tags
}

func (f *FileMetaStorage) GetTagsAliases() map[string]string {
	return f.meta.TagsAliases
}

func (f *FileMetaStorage) SetTagsAliases(aliases map[string]string) {
	f.hasChanges = true
	f.meta.TagsAliases = aliases
}

func (f *FileMetaStorage) GetTagsListMessageID() int {
	return f.meta.TagsListMessageID
}

func (f *FileMetaStorage) SetTagsListMessageID(id int) {
	f.meta.TagsListMessageID = id
}

func (f *FileMetaStorage) GetSentAnimations() map[string]*SentAnimation {
	return f.meta.Messages
}

func (f *FileMetaStorage) AddSentAnimations(messages map[string]*SentAnimation) {
	f.hasChanges = true

	if f.meta.Messages == nil {
		f.meta.Messages = messages

		return
	}

	for key, msg := range messages {
		f.meta.Messages[key] = msg
	}
}

func (f *FileMetaStorage) Close() {
	if !f.hasChanges {
		return
	}

	data, err := json.Marshal(f.meta)
	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile(f.filename, data, 0666); err != nil {
		panic(err)
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
