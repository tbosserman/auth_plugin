Name:		auth_plugin
Version:	0.1
Release:	1%{?dist}
Summary:	An docker plugin for authenticating calls to dockerd

License:        BSD
Source0:	auth_plugin.tar.gz

%description
Not really an auth plugin per-se. Right now all it does is log all the
attempts to talk to dockerd. It returns "allow" for all requests.

%prep
%setup -q

%build
make %{?_smp_mflags}

%install
make install DESTDIR=%{buildroot}

%files
/usr/local/bin/auth_plugin
/usr/lib/systemd/system/auth_plugin.service
/etc/systemd/system/docker.service.d/override.conf

%doc

%changelog

%post
systemctl enable auth_plugin
systemctl start auth_plugin
systemctl daemon-reload
systemctl restart docker

%preun
systemctl stop auth_plugin
systemctl disable auth_plugin

%postun
systemctl daemon-reload
systemctl restart docker
