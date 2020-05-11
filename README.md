# auth_plugin
Skeleton of a Docker authentication plugin

### BUILD INSTRUCTIONS

1. Go to the releases and get the URL to the .tar.gz file
2. cd ~/rpmbuild/SOURCES
3. curl -o auth_plugin.tar.gz <curl>
4. cd ../SPECS
4. tar -O -xf ../SOURCES/auth_plugin.tar.gz auth_plugin-0.1/auth_plugin.spec > auth_plugin.spec
5. rpmbuild -bb auth_plugin.spec
