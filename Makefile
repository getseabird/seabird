patch:
	@if [ "$$(git rev-parse --abbrev-ref HEAD)" != "main" ]; then exit 1; fi
	git pull -r
	semver up release

minor:
	@if [ "$$(git rev-parse --abbrev-ref HEAD)" != "main" ]; then exit 1; fi
	git pull -r
	semver up minor

major:
	@if [ "$$(git rev-parse --abbrev-ref HEAD)" != "main" ]; then exit 1; fi
	git pull -r
	semver up major

release:
	sed -i "/<releases>/a \    <release version=\"$$(semver get release)\" date=\"$$(date +%F)\"/>" dev.skynomads.Seabird.appdata.xml
	git add .semver.yaml dev.skynomads.Seabird.appdata.xml
	git commit -m "$$(semver get release)"
	git tag -a -m "$$(semver get release)" "$$(semver get release)"