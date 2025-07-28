# Default target
build:
	go build -o op-dotenv .

install: build
	cp op-dotenv /usr/local/bin/

# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
#   Run release commands
# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
release/patch:
	@if [ $$(git tag | wc -l) -eq 0 ]; then \
		NEW_TAG="v0.0.1"; \
	else \
		LATEST_TAG=$$(git describe --tags `git rev-list --tags --max-count=1`); \
		MAJOR=$$(echo $$LATEST_TAG | cut -d. -f1 | tr -d 'v'); \
		MINOR=$$(echo $$LATEST_TAG | cut -d. -f2); \
		PATCH=$$(echo $$LATEST_TAG | cut -d. -f3); \
		NEW_PATCH=$$((PATCH + 1)); \
		NEW_TAG="v$$MAJOR.$$MINOR.$$NEW_PATCH"; \
	fi; \
	git tag -a $$NEW_TAG -m "Release $$NEW_TAG" && \
	echo "Created new tag: $$NEW_TAG" && \
	git push origin $$NEW_TAG

release/minor:
	@if [ $$(git tag | wc -l) -eq 0 ]; then \
		NEW_TAG="v0.1.0"; \
	else \
		LATEST_TAG=$$(git describe --tags `git rev-list --tags --max-count=1`); \
		MAJOR=$$(echo $$LATEST_TAG | cut -d. -f1 | tr -d 'v'); \
		MINOR=$$(echo $$LATEST_TAG | cut -d. -f2); \
		NEW_MINOR=$$((MINOR + 1)); \
		NEW_TAG="v$$MAJOR.$$NEW_MINOR.0"; \
	fi; \
	git tag -a $$NEW_TAG -m "Release $$NEW_TAG" && \
	echo "Created new tag: $$NEW_TAG" && \
	git push origin $$NEW_TAG

release/major:
	@if [ $$(git tag | wc -l) -eq 0 ]; then \
		NEW_TAG="v1.0.0"; \
	else \
		LATEST_TAG=$$(git describe --tags `git rev-list --tags --max-count=1`); \
		MAJOR=$$(echo $$LATEST_TAG | cut -d. -f1 | tr -d 'v'); \
		NEW_MAJOR=$$((MAJOR + 1)); \
		NEW_TAG="v$$NEW_MAJOR.0.0"; \
	fi; \
	git tag -a $$NEW_TAG -m "Release $$NEW_TAG" && \
	echo "Created new tag: $$NEW_TAG" && \
	git push origin $$NEW_TAG

clean:
	rm -rf op-dotenv dist/

.PHONY: build install release/patch release/minor release/major clean
