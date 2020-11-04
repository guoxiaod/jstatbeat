Name:		jstatbeat
Version:	1.0.0
Release:	2%{?dist}
Summary:	jstatbeat to elasticsearch

Group:		default
License:	MIT
URL:		https://github.com/guoxiaod/jstatbeat
Source0:	%{name}-%{version}.tgz

%description
jstatbeat

%prep
%setup -q


%build


%install

mkdir -p %{buildroot}%{_sysconfdir}/jstatbeat/
mkdir -p %{buildroot}%{_sbindir}
mkdir -p %{buildroot}%{_unitdir}
cp -r *.json *.yml %{buildroot}%{_sysconfdir}/jstatbeat/
cp -r jstatbeat %{buildroot}%{_sbindir}
cp -r jstatbeat.service %{buildroot}%{_unitdir}
mkdir -p %{buildroot}%{_localstatedir}/log/jstatbeat/

%post

systemctl daemon-reload

%postun

systemctl daemon-reload

%files
%config(noreplace) %{_sysconfdir}/jstatbeat/
%{_sbindir}/jstatbeat
%{_unitdir}/jstatbeat.service
%{_localstatedir}/log/jstatbeat/


%changelog

