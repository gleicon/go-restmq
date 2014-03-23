%define debug_package %{nil}

Summary: RestMQ is a message queue server
Name: restmq
Version: 2.0.1
Release: 1
License: MIT
Group: Utilities
URL: http://restmq.com
Packager: Alexandre Fiori <fiorix@gmail.com>
Source: %{name}-%{version}.tar.gz
#Requires: redis >= 2.8
BuildRoot: %{_tmpdir}/%{name}-%{version}-%{release}
BuildRequires: gcc
BuildRequires: make

%description
RestMQ is a message queue which uses HTTP as transport, JSON to
format a minimalist protocol and is organized as REST resources.

%prep
%setup -q

%build
make

%install
make DESTDIR=%{buildroot} install
install -d %{buildroot}/etc/init
install -m 0644 redhat/upstart %{buildroot}/etc/init/restmq.conf
install -d %{buildroot}/etc/logrotate.d
install -m 0644 redhat/logrotate %{buildroot}/etc/logrotate.d/restmq

%files
%config /opt/restmq/restmq.conf
/etc/init
/etc/logrotate.d
/opt/restmq

%post
/usr/sbin/useradd -M -r -d /opt/restmq restmq || :
/bin/mkdir /var/log/restmq || :
/bin/chown -R restmq:restmq /opt/restmq /var/log/restmq || :

%preun
stop restmq > /dev/null 2>&1 || :

%postun
/usr/sbin/userdel restmq || :

%changelog
* Sun Mar 23 2014 Alexandre Fiori <fiorix@gmail.com> 2.0.1-1
- Initial release
