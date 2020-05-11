# auth_plugin
Skeleton of a Docker authentication plugin

### BUILD INSTRUCTIONS

1. Go to the releases and get the URL to the .tar.gz file
2. cd ~/rpmbuild/SOURCES
3. curl -L -o auth_plugin.tar.gz <curl>
4. cd ../SPECS
4. tar -O -xf ../SOURCES/auth_plugin.tar.gz auth_plugin-0.1/auth_plugin.spec > auth_plugin.spec
5. rpmbuild -bb auth_plugin.spec

### PACKAGE DETAILS

When this package is installed the following tasks are performed:
the following tasks:

1. A systemd override file is created in /etc/systemd/system/docker.service.d/override.conf
2. This package is enabled via `systemctl enable auth_plugin`
3. Auth_plugin is started via `systemctl start auth_plugin`
4. `systemctl daemon-reload` is executed to get systemd to read the override.
5. Docker is restarted via `systemctl restart docker` so that Docker will
enable the authentication plugin and start communicating with it.

When the RPM is uninstalled all of that is reversed:

1. The auth plugin is stopped
2. The auth plugin is disabled
3. The override file gets deleted
4. `systemctl daemon-reload` is called
5. `systemctl restart docker` is called.
