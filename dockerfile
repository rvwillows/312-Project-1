FROM golang:1.17.6
# Set the home directory to /root
ENV HOME /root
# cd into the home directory
WORKDIR /root
# Copy all app files into the image
COPY . .
# Allow port 8000 to be accessed # from outside the container
EXPOSE 8000

ADD https://github.com/ufoscout/docker-compose-wait/releases/download/2.2.1/wait /wait
RUN chmod +x /wait

# Run the app
CMD /wait && go build . && ./server