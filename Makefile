.PHONY: build dist

CMD = $@
ENV = $*

NAME = pubsrc
VERSION = 0.0.1
BUILD 	= $(shell git rev-parse --short HEAD)
CONFIGDATA 	= $(shell cat config.${ENV}.yaml | base64)

GO = GO111MODULE=on go

%-dev: GO = GO111MODULE=on GOOS=linux GOARCH=amd64 go
%-prod: GO = GO111MODULE=on GOOS=linux GOARCH=amd64 go

fox-%: out = ${NAME}.${ENV}
fox-%:
	@${GO} build -ldflags "-w -s 		\
	-X 'main.NAME=${NAME}' 				\
	-X 'main.VERSION=${VERSION}' 		\
	-X 'main.BUILD=${BUILD}' 			\
	-X 'main.CONFIGDATA=${CONFIGDATA}' 		\
	" -o ${out}
	@echo ""
	@echo `file ${out}`, `du -h ${out} | cut -f1`

fox: fox-local

%-dev: DISTHOST = ubuntu@foxone-dev
dist-%: out = ${NAME}.${ENV}
dist-%:
	@echo copy ${out} to aws ...
	scp ${out} ${DISTHOST}:/usr/local/foxone/luckycoin/lucky.new
	ssh ${DISTHOST} 'bash -s' < ./scripts/deploy.sh

dist: build-dev dist-dev
