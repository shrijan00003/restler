# Restler

Restler is a file based REST API client. If you are familiar with tools like Postman, Insomnia, or curl, then you will feel right at home with Restler. This is file based, opinionated, command line tool. Please follow Installation and Usage guide to get started.

## Installation

### MacOS x86_64

```bash
curl -s https://raw.githubusercontent.com/shrijan00003/restler/main/install/darwin-amd64.sh | bash
```

### MacOS arm64

```bash
curl -s https://raw.githubusercontent.com/shrijan00003/restler/main/install/darwin-arm64.sh | bash
```

### Linux x86_64

```bash
curl -s https://raw.githubusercontent.com/shrijan00003/restler/main/install/linux-amd64.sh | bash
```

## User Guide

1. Create a folder for the API collection, default is `restler`.
2. Create a folder `requests/<request-name>` like `requests/posts` inside the API collection folder and create a request file inside it.
3. Create a request file with `.<http-method>.yaml` extension. For example `posts.post.yaml`, or `posts.get.yaml`
4. (Optional) Create `env` folder inside the API collection folder and create a `.yaml` file inside it. For example `env/default.yaml` for using environment variables in request.
5. (Optional) Create `config.yaml` file inside the API collection folder for configurations like environment.
6. (Optional) Change folder for API collection in runtime by setting `RESTLER_PATH` environment variable. For example `export RESTLER_PATH=app-prefix-collection`
7. Run `restler <http-method> <request-name>` to run the request. For example `restler psot posts` to run post request and `restler get posts` to run get request.
8. Check the output files in `requests/<request-name>` folder. For example `requests/posts/.post.res.txt` for post response and `requests/posts/.get.res.txt` for get request response.
9. for other supports please check the [TODO](./todos/Todo.md) file.

## Commands Available
- `restler post <request-name>`
- `restler get <request-name>`
- `restler patch <request-name>`
- `restler put <request-name>`
- `restler delete <request-name>`

## Proxy Usage

Restler respects the `HTTPS_PROXY` and `HTTP_PROXY` environment variables. You can specify a proxy URL for individual requests using the `R-Proxy-Url` header. To disable the proxy for specific requests, use the `R-Proxy-Enable: N` header.

for eg:
```yaml
Name: Get Posts
URL: "{{API_URL}}"
Method: GET

Headers:
  Accept: text/html, application/json
  Accept-Encoding: utf-8
  R-Proxy-Url: https://something.com:8080
  R-Proxy-Enable: N # N or Y, Y is default, if HTTPS_PROXY, HTTP_PROXY or R-Proxy-Url is set
  User-Agent: rs-client-0.0.1
  Content-Type: application/json

Body:
```

## Inspiration

## Build

go build -o restler bin/main.go
