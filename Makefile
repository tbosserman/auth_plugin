all: auth_plugin

auth_plugin: auth_plugin.go
	go build auth_plugin.go

install: auth_plugin
	install -d $(DESTDIR)/usr/local/bin
	install -d $(DESTDIR)/usr/lib/systemd/system
	install -d $(DESTDIR)/etc/systemd/system/docker.service.d
	install -m 0755 auth_plugin $(DESTDIR)/usr/local/bin
	install -m 0644 auth_plugin.service $(DESTDIR)/usr/lib/systemd/system
	install -m 0644 override.conf $(DESTDIR)/etc/systemd/system/docker.service.d
