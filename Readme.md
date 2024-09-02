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

1. Create a folder for API collection, default is `restler`
2. Create a folder `requests/<request-name>` like `requests/posts` inside the API collection folder and create a request file inside it.
3. Create a request file with `.request.yaml` extension. For example `posts.request.yaml`
4. (Optional) Create `env` folder inside the API collection folder and create a `.yaml` file inside it. For example `env/default.yaml` for using environment variables in request.
5. (Optional) Create `config.yaml` file inside the API collection folder for configurations like environment.
6. (Optional) Change folder for API collection in runtime by setting `RESTLER_PATH` environment variable. For example `export RESTLER_PATH=app-prefix-collection`
7. Run `restler <request-name>` to run the request. For example `restler posts`
8. Check the output files in `requests/<request-name>` folder. For example `requests/posts/.post.response.txt`
9. for other supports please check the [TODO](./todos/Todo.md) file.

## Inspiration

## Build

go build -o restler bin/main.go
