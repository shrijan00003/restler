# Environment Support
support for .env.* files

## How to Access?
```yaml
URL: ${API_URL}
```
- os pacakge can be used to replace the environmetn variable natively we don't have to parse the string ourselves
- we can setup the env on the fly with os.Setenv and unset it with os.Unsetenv if needed
- env varaibles are not suppossed to be shared with other developers, so that for sharable configurations we can use config.yaml

## How to Set?
```yaml
After:
  Env:
    API_URL: https://api.example.com
```

# Config Support
support for config.yaml file???
Do we need config.yaml? what is difference between env and config?


sample config.yaml

```yaml
API_URL: https://api.example.com
X-API-KEY: ${API_KEY}

```

## How to Access?
In request file like get-token.yaml
Ref: https://stackoverflow.com/questions/69657289/use-environment-variables-in-yaml-file

```yaml
URL: {{.API_URL}}
Headers:
  X-API-KEY: {{.X-API-KEY}}
```

## How to Set?


## Different scenarios for using env and config?


## How can user can use different envrionment files? if they have to use .env.dev or .env.prod or .env.test?
Option 1: Do not support .env.* files and only support .env file
Option 2: Use one more file like config.yaml and defined which env file to use like env=dev
Oprion 3: Inreoduce set command to set the env file to use like `restler set env dev` and `restler set env prod`

# How does set command should work?
- command `set env dev | test | prod`
  -- create .restler/config.yaml file
  sample config.yaml

  ```yaml
  env: dev | test | prod
  ```
- if we want to support config.yaml file, the immidiate config.yaml file will be used to determine the env and other configurations.
- Do we want to add a support for config vars?
- But we can't set values on the config.yaml file, as it will/can be pushed to git that is not expected on most of the cases.

## So how after should work?
```yaml
After:
  Env:
    API_URL: https://api.example.com # will be written to current env file
  Config:
    API_KEY: 1234567890 # will be written to .restler/config.yaml
```

### If we set .restler/config.yaml file what will be our precedence?
- .restler should be considered as the current cache folder
- it should be loaded

flow:
1. call ser1 -> get token
2. use that token -> api 2
3. use some data from there -> use in api 3


- There should not be more than one config file for sure
.env or .env.local
RESETLER_ENV=prod

.restler
  - .env
  - .env.dev
  - .env.prod
  - .env.test
  - .env.<whatever>
  - cache
    -- cache-structure-1



## After scripts are crazy?
- simple use case

- Since we can't provide the ts or js event on that
After:
  script:
    - language: js/ts
      action: scripts/functionName  ->  (api) -> ({req, res, env})
    - language: go
      action: go-scripts/funcName -> (api *r.Api) ->
