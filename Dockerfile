FROM docker.io/cimg/go:1.25-browsers

WORKDIR /app

COPY ./src .

ENV HOST=0.0.0.0:3000
# THIS MUST BE SET!!
ENV SECRET_KEY=

EXPOSE 3000

CMD ["go", "run", "main.go"]
