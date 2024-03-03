SEMVER = go run github.com/maykonlf/semver-cli/cmd/semver@latest

.PHONY: patch
patch:
	@if [ "$$(git rev-parse --abbrev-ref HEAD)" != "main" ]; then exit 1; fi
	git pull -r
	$(SEMVER) up release

.PHONY: minor
minor:
	@if [ "$$(git rev-parse --abbrev-ref HEAD)" != "main" ]; then exit 1; fi
	#git pull -r
	$(SEMVER) up minor

.PHONY: release
major:
	@if [ "$$(git rev-parse --abbrev-ref HEAD)" != "main" ]; then exit 1; fi
	git pull -r
	$(SEMVER) up major

.PHONY: release
release:
	sed -i "/<releases>/a \    <release version=\"$$($(SEMVER) get release)\" date=\"$$(date +%F)\"/>" dev.skynomads.Seabird.appdata.xml
	git add .semver.yaml dev.skynomads.Seabird.appdata.xml
	git commit -m "$$($(SEMVER) get release)"
	git tag -a -m "$$($(SEMVER) get release)" "$$($(SEMVER) get release)"