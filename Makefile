run_race:
	go run -race main.go

run: build run_root

build:
	go build -o ./bin/gifextractor .

run_root:
	./bin/gifextractor --config=./config.json

gen_mock:
	minimock -i ./favchannel/extractor.extractorClient -o ./favchannel/extractor/
	minimock -i ./favchannel/extractor.storage -o ./favchannel/extractor/
	minimock -i ./bot.GifkoskladMetaStorage -o ./favchannel/publish
	minimock -i ./favchannel/publish.publisherClient -o ./favchannel/publish

test_cover:
	go test -v -coverprofile=test-coverage.out ./... && \
	go tool cover -html=test-coverage.out -o=test-coverage.html && \
	rm test-coverage.out

db_backup:
	cp db.json db.backup.json
	cp gifs_with_tags.json gifs_with_tags.backup.json