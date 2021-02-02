NAME=jstatbeat
VERSION=1.0.1

RPM_PATH=${NAME}-${VERSION}
RPM_FILE=${RPM_PATH}.tgz

all:

build: prepare
	rpmbuild -tb ${RPM_FILE}

prepare:
	if [ -f ../${NAME} ]; then 	\
		echo "Please compile ${NAME} firsat"; \
	fi
	mkdir -p ${RPM_PATH}
	cp ../${NAME} ${RPM_PATH}/
	cp *.yml *.json *.spec *.service ${RPM_PATH}/
	tar -czf ${RPM_FILE} ${RPM_PATH}

