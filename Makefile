VERSION=$(shell semver get release)

release-patch:
	semver up release
	make release

release:
	git commit --allow-empty -m "$(VERSION)"
	git tag -a -m "$(VERSION)" "$(VERSION)"