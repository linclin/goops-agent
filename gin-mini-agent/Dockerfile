# syntax=registry.cn-shenzhen.aliyuncs.com/dev-ops/dockerfile-upstream:1.21.0
FROM registry.cn-shenzhen.aliyuncs.com/dev-ops/golang:1.26.1-alpine3.23 as golang
ENV APP gin-mini-agent  
RUN sed -i 's/https/http/' /etc/apk/repositories && \
    sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk update && \
    apk add --no-cache ca-certificates git && \
    rm -rf /var/cache/apk/*   /tmp/*  
ADD ./ /app/gin-mini-agent
ADD .git/ /app/gin-mini-agent/.git
WORKDIR /app/gin-mini-agent
RUN export GitBranch=$(git name-rev --name-only HEAD) && \
    export GitRevision=$(git rev-parse --short HEAD) && \
    export GitCommitLog=`git log --pretty=oneline -n 1` && \
    export BuildTime=`date +'%Y.%m.%d.%H%M%S'` && \
    export BuildGoVersion=`go version` && \
    export LDFlags="-s -w -X 'main.GitBranch=${GitBranch}' -X 'main.GitRevision=${GitRevision}' -X 'main.GitCommitLog=${GitCommitLog}' -X 'main.BuildTime=${BuildTime}' -X 'main.BuildGoVersion=${BuildGoVersion}'"  && \
    go build -tags=jsoniter -ldflags="$LDFlags" -o  ./gin-mini-agent

FROM registry.cn-shenzhen.aliyuncs.com/dev-ops/alpine:3.23.3
LABEL MAINTAINER="13579443@qq.com"
ENV TZ='Asia/Shanghai' 
ENV LANG UTF-8
ENV LC_ALL zh_CN.UTF-8
ENV LC_CTYPE zh_CN.UTF-8
RUN TERM=linux && export TERM
RUN sed -i 's/https/http/' /etc/apk/repositories && \
    sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk update && \
    apk add --no-cache ca-certificates tzdata bash sudo busybox-extras curl && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtim && \
    echo "Asia/Shanghai" > /etc/timezone && \ 
    rm -rf /var/cache/apk/*   /tmp/* && \
    addgroup -g 1000 app && \
    adduser -u 1000 -G app -D app && \
    adduser -u 1001 -G app -D dev && \
    chmod 4755 /bin/busybox  && \
    mkdir -p /app  && \
    chown -R app:app /app 
WORKDIR /app/gin-mini-agent/
COPY --from=golang --chown=app:app --chmod=755 /app/gin-mini-agent/gin-mini-agent /app/gin-mini-agent/gin-mini-agent 
COPY --from=golang --chown=app:app --chmod=755 /app/gin-mini-agent/conf /app/gin-mini-agent/conf  
CMD ["./gin-mini-agent"]