release:
	semver up release
	git add .semver.yaml
	git commit --allow-empty -m "$$(semver get release)"
	git tag -a -m "$$(semver get release)" "$$(semver get release)"