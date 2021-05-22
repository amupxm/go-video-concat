FROM  golang:1.16.4

WORKDIR /go/src/app/amupxm

RUN apt-get update && apt-get install -y \
    ffmpeg \
    libmediainfo-dev \
    zlib* \
    gcc  && rm -rf /var/lib/apt/lists/*
COPY . .

RUN go get .


RUN go build -v .

CMD [ "./go-video-concat" ]