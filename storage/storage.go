package storage

import (
	"encoding/json"
	"errors"
	"fmt"
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
		return nil, fmt.Errorf("open storage file: %w", err)
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

func (f *FileMetaStorage) SetFavChannelLastForwardedMessageIDWithoutCaption(id int64) {
	if f.meta.LastForwardedMessageIDWithoutCaption != id {
		f.meta.LastForwardedMessageIDWithoutCaption = id
		f.hasChanges = true
	}
}

func (f *FileMetaStorage) GetFavChannelLastForwardedMessageIDWithoutCaption() int64 {
	return f.meta.LastForwardedMessageIDWithoutCaption
}

func (f *FileMetaStorage) Close() {
	if !f.hasChanges {
		return
	}

	if err := f.Write(); err != nil {
		panic(err)
	}
}

func (f *FileMetaStorage) Write() error {
	data, err := json.Marshal(f.meta)
	if err != nil {
		return fmt.Errorf("marshal meta data: %w", err)
	}

	if err := ioutil.WriteFile(f.filename, data, 0666); err != nil {
		return fmt.Errorf("write meta data file: %w", err)
	}

	return nil
}

type metaData struct {
	Tags        []string
	TagsAliases map[string]string
	// Messages все отправленные ранее сообщения для редактирования
	Messages                             map[string]*SentAnimation
	LastForwardedMessageIDWithoutCaption int64
}

type SentAnimation struct {
	MessageID int
	FileID    string
	Tags      []string
}
