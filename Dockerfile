# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:1.13

# Copy the local package files to the container's workspace.
ADD . /go/src/github.com/yamamushi/covidtracker

# Create our shared volume
RUN mkdir /covidtracker-bot

# Run our dependency installation for Opus Encoding/Decoding
#RUN apt-get update && \
#        DEBIAN_FRONTEND=noninteractive apt-get install -y libav-tools opus-tools -f && \
#        apt-get clean && \
#        rm -rf /var/lib/apt/lists/

# Get the du-discordbot dependencies inside the container.
RUN go get github.com/anaskhan96/soup
RUN cd /go/src/github.com/anaskhan96/soup && git checkout ad448eafe
RUN go get github.com/bwmarrin/discordgo
RUN go get github.com/BurntSushi/toml


# Install and run du-discordbot
RUN go install github.com/yamamushi/covidtracker

# Run the outyet command by default when the container starts.
WORKDIR /covidtracker-bot
ENTRYPOINT /go/bin/covidtracker