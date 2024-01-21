release:
	semver up release
	git add .semver.yaml
	VERSION=$(shell semver get release)
	git commit --allow-empty -m "$(VERSION)"
	git tag -a -m "$(VERSION)" "$(VERSION)"