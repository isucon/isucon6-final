FROM openjdk:8-jdk-alpine

# https://hub.docker.com/r/hseeberger/scala-sbt/~/dockerfile/

RUN apk --update add bash curl && rm -rf /var/cache/apk/*

ENV SBT_VERSION 0.13.12
ENV SBT_HOME /usr/local/sbt
ENV PATH ${PATH}:${SBT_HOME}/bin

RUN curl -sL "http://dl.bintray.com/sbt/native-packages/sbt/$SBT_VERSION/sbt-$SBT_VERSION.tgz" | gunzip | tar -x -C /usr/local && \
    echo -ne "- with sbt $SBT_VERSION\n" >> /root/.built

WORKDIR /app

# キャッシュ効率を上げるためにsbtのインストールだけ先にする
RUN mkdir -p /app/project
COPY ./build.sbt /app
COPY ./project/plugins.sbt /app/project
COPY ./project/build.properties /app/project
RUN sbt compile

COPY . /app

RUN sbt assembly

CMD ["java", "-jar", "target/scala-2.11/isuketch-assembly-1.0.jar"]
