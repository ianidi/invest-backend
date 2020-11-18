############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder

ENV VIPSVERSION 8.8.3
# Install git.
# Git is required for fetching the dependencies.
#RUN apk update && apk add --no-cache git

RUN apk update && apk add --no-cache git
RUN apk add build-base pkgconfig glib-dev gobject-introspection-dev expat-dev tiff-dev libjpeg-turbo-dev libexif-dev giflib-dev librsvg-dev lcms2-dev libpng orc-dev libwebp-dev libheif-dev
RUN apk add libimagequant-dev --repository=http://dl-cdn.alpinelinux.org/alpine/edge/community
#RUN apk add libimagequant-dev --repository=https://pkgs.alpinelinux.org/package/edge/community

RUN \
  # Build libvips
  cd /tmp && \
  wget https://github.com/libvips/libvips/releases/download/v$VIPSVERSION/vips-$VIPSVERSION.tar.gz && \
  tar zxvf vips-$VIPSVERSION.tar.gz && \
  cd /tmp/vips-$VIPSVERSION && \
  ./configure --enable-debug=no --without-python $1 && \
  make && \
  make install

RUN ldconfig /

# Create appuser.
#ENV USER=appuser
#ENV UID=10001 
# See https://stackoverflow.com/a/55757473/12429735RUN 
#RUN adduser \    
#  --disabled-password \    
#  --gecos "" \    
#  --home "/nonexistent" \    
#  --shell "/sbin/nologin" \    
#  --no-create-home \    
#  --uid "${UID}" \    
#  "${USER}"
WORKDIR $GOPATH/src/mypackage/myapp/
COPY . .
# Fetch dependencies.
# Using go get.
#RUN go get -d -v
# Using go mod.
# RUN go mod download
# RUN go mod verify
ENV GO111MODULE=on
ENV GIN_MODE=release
# Build the binary.
RUN GOOS=linux GOARCH=amd64 go build -tags=jsoniter -ldflags="-w -s" -o /go/bin/main
############################
# STEP 2 build a small image
############################
#FROM scratch
# Import the user and group files from the builder.
#COPY --from=builder /etc/passwd /etc/passwd
#COPY --from=builder /etc/group /etc/group
# Copy our static executable.
#COPY --from=builder /go/bin/main /go/bin/main
# Use an unprivileged user.
#USER appuser:appuser

RUN chmod +x /go/bin/main
EXPOSE 4000
# Run the hello binary.
ENTRYPOINT ["/go/bin/main"]

#CMD ["/go/bin/main"]