FROM docker.io/cimg/go:1.25-browsers

WORKDIR /app

COPY ./src .

ENV HOST=0.0.0.0:3000
ENV SECRET_KEY= >>> THIS IS THE JWT's SECRET KEY <<<

EXPOSE 3000

CMD ["go", "run", "main.go"]
