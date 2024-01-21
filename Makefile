VERSION=$(shell semver get release)

release:
	semver up release
	git add .semver.yaml
	git commit --allow-empty -m "$(VERSION)"
	git tag -a -m "$(VERSION)" "$(VERSION)"