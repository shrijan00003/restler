# Environment Support

By default, Restler supports the `.env` and `.env.local` files as environment. The `.env.local` file have precedence over `.env` file.
The `.env` file is loaded first and then the `.env.local` file is loaded in case both are present. You can specify the environment to choose by
setting `Env` and `EnvPath` in the `config.yaml` file in your root directory.

## Example

```yaml
Env: "local"
EnvPath: ".env"
```

## Env

If `Env` is set it will load the environment file as `cwd()+".env".${Env}`. For example, if `Env: local` is set, it will load the environment file as `cwd()+".env.local"`.
**Note**: EnvPath have higher precedence over Env. If both are set, the EnvPath will be used to load the environment file. But if you have unique environment
in `Env` that is not present in `EnvPath`, that environment will be loaded.

## EnvPath

EnvPath will expect absolute or relative path to the environment file. If `EnvPath` is set, it will load the environment file from the specified
path.
