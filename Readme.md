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

## Getting Started

## Project Initialization and Structure

**1. Initialization using `restler init`:**

- **Usage:** Run `restler init` in your terminal.
- **Prompt:** You'll be prompted to enter the desired project name.
- **Example:** If you enter `collection`, a folder named `collection` will be created in your current directory. This folder includes default files and subfolders for your API collection.
- **Configuration:**
  - The `.env` file will be created with `RESTLER_PATH=collection` to specify the project directory.
  - The `.gitignore` file will be updated to ignore environment and response files from version control.

**2. Manual Project Structure Creation:**

- **API Collection Folder:** Create a folder for your API collection. The default name is `restler`, but you can choose a different name.
- **Request Folder Structure:**
  - Create a subfolder named `requests` inside the API collection folder.
  - Within the `requests` folder, create subfolders for each API request (e.g., `requests/posts`).
  - Inside each request subfolder, create a YAML file with the extension corresponding to the HTTP method:
    - `.<http-method>.yaml` (e.g., `posts.post.yaml` for a POST request or `posts.get.yaml` for a GET request).
- **Environment Variables (Optional):**
  - Create an `env` folder inside the API collection folder.
  - Within the `env` folder, create a YAML file for environment variables (e.g., `env/default.yaml`).
  - This allows you to define and use environment variables in your API requests.
- **Configuration (Optional):**
  - A `config.yaml` file can be created in the API collection folder to store configurations like the default environment.
- **Runtime Environment Variable:**
  - To change the project directory at runtime, set the `RESTLER_PATH` environment variable.
  - Example: `export RESTLER_PATH=app-prefix-collection`
- **Running Requests:**
  - Use the `restler` command followed by the HTTP method and request name to execute the desired request.
  - Example: `restler p posts` executes a POST request defined in `posts.post.yaml`, while `restler g posts` executes a GET request from `posts.get.yaml`.

This approach allows for a more customized project structure and configuration for your specific needs.

## User Guide

1. Create a folder for the API collection, default is `restler`.
2. Create a folder `requests/<request-name>` like `requests/posts` inside the API collection folder and create a request file inside it.
3. Create a request file with `.<http-method>.yaml` extension. For example `posts.post.yaml`, or `posts.get.yaml`
4. (Optional) Create `env` folder inside the API collection folder and create a `.yaml` file inside it. For example `env/default.yaml` for using environment variables in request.
5. (Optional) Create `config.yaml` file inside the API collection folder for configurations like environment.
6. (Optional) Change folder for API collection in runtime by setting `RESTLER_PATH` environment variable. For example `export RESTLER_PATH=app-prefix-collection`
7. Run `restler <http-method> <request-name>` to run the request. For example `restler psot posts` to run post request and `restler get posts` to run get request.
8. Check the output files in `requests/<request-name>` folder. For example `requests/posts/.<request-name>.post.res.md` for post response and `requests/posts/.get.res.txt` for get request response.
9. for other supports please check the [TODO](./todos/Todo.md) file.

## Commands Available

- `restler post <request-name>` or `restler p <request-name>`
- `restler get <request-name>` or `restler g <request-name>`
- `restler patch <request-name>` or `restler m <request-name>`
- `restler put <request-name>` or `restler u <request-name>`
- `restler delete <request-name>` or `restler d <request-name>`
- `restler --help` or `restler -h` or or `restler -h`
- `restler --version` or `restler -v`

## Flag support

Now all our REST method commands supports following flags:

### --env or -e

`env` flag is compatible for selecting environment from the command line. If we don't pass this option in command `Env` from config.yaml is default.

### Usage

```bash
restler <http-command> --env <env-value> <request name>
```

Fog eg,

```bash
restler p -e dev posts
```

### --request or -r

`request` flag is useful for seleting individual request from the request collection if you have multiple requests of same http method. Consider following structure:

```bash
reslter
    requests
        posts
            - posts.post.yaml
            - posts-v2.post.yaml


```

For running posts-v2, we can use `-r` like following:

```bash
restler p -r posts-v2 posts
```

## Setup Environment Variables from Response

For setting environment variables from response, we can use `After` section in request file. For example:

```yaml
Name: Create Post
URL: "{{API_URL}}"
Method: POST

Headers: ...

Body: ...

After:
  Env:
    ADDRESS_STREET: Body[address][street]
    HOBBIES: Body[hobbies][0]
    RESPONSE_DATE: Header[Date]
    EMPLOYMENT_START_DATE: Body[employment][details][start_date]
```

Here, `Body` and `Header` are special keys for accessing response body and response headers respectively. For accessing, we can use `Body[key]` or `Header[key]` syntax. If you have array like structure, you can access it using `Body[key][index]` or `Header[key][index]`. If value is not found, it will write empty string in your env file.

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
