# Environment Support
By default, restler will load `.env` file or `.env.local` file on the `cwd`, if both present `.env.local` will be priorized over `.env`.
- if config.yaml file is on cwd and it has Env filed, it will be used to load the env file.

## todo DEC 15 2024
- [ ] Add support for EnvPath on request file flag (get from anywhere - relative or absolute)
- [ ] Load .env.local or .env file on the cwd
- [ ] Add support for envpath flag
- [ ] Add support for EnvPath on After Hook. (set anywhere)
- [ ] Add support for config.yaml file (optional)

### Suppport for EnvPath on request file ( get from anywhere - relative or absolute)
- If Envpath is present on the request file, it should load only that env file.
- If EnvPath is not present then it should look for .env.local or .env file on the cwd.
- If not found It should ignore the error and continue, but if request file have env used on the request, it should throw error.

## Questions

### How can i choose what env file to load?
1. config.yaml (search for config.yaml file)
2. env flag (--env=local or -e local) (search for .env.local file)
3. envpath flag (--envpath=.env.local or -e .env.local) (search for .env.local file)

### What if i need to run nested collections from root? which env file should be loaded?
1. If config.yaml is present:
- `Env: local` --> load `.env.local` file
- `EnvPath: .env.local` --> load `.env.local` file
- `EnvPath: ../something/.something.env` --> load `../something/.something.env` file

**Note:** `EnvPath` will be prioritized over `Env` if both are present.

In this way we can speficy what env file to load and where to load it from.

2. Request file can also have `EnvPath` and `Env` fields

#### What can be issue we need to look into?
  1. Can we load env file above the current folder?


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
- config.yaml can be used to set the differnet env files.
- config.yaml can be also used to specify the env file path to load.
- TODO: We can havee more public config options like API_URL, API_KEY which can be shared over git.
- for now we can use .env.example file to share env variables over git that can be copy and renamed to .env file as secret.
