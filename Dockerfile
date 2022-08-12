FROM golang:1.18-buster

RUN apt-get update && apt-get install -y apt-transport-https

RUN curl -fsSL https://deb.nodesource.com/setup_18.x | bash -

RUN apt install -y fontconfig nodejs build-essential libcairo2-dev libpango1.0-dev libjpeg-dev libgif-dev librsvg2-dev

ENV NODE_OPTIONS --max-old-space-size=4096

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN cd ./renderer && npm install

# Build golang app
RUN go build -o server .

CMD [ "/app/server" ]
