# Restler Readme for new version of API

1. Remove Env folder with config.yaml
- One more folder will be removed
- structure will be something like

```bash
Env: default


Envs:
    default:
        Key1: Value1
    dev:
        Key1: Value.dev.1

```

2. Config support
- if current request folder have config.yaml then it will read/write from that yaml file.
- otherwise it will keep searching for its immidiate parent upto RESTLER_PATH for config and load it.
- It will write accordingly.
- if config.yaml file is present but it doesn't have that element? error or search for params?
  - may be we can expect {{inherit}}

3. Git question
- How can we avoid to push the tokens and environments to the server??
- if we will keep prod envs on the same config.yaml file then we might accidently push the environments to git
- may be we can introduce RESTLER_ENV=prod that will load the config.prod.yaml file or expect that file
- if RESTLER_CONFIG=something then config.something.yaml file will be loaded so that config.prod.yaml file can be ignored on the git.
- it should support .env, .env.* file


3. What is the best way to run the request
- restler p article -- RESTLER_PATH/article/article.post.yaml
- restler article/article.post.yaml
- restler -p article post article
- retler filename.yaml -- method will define which one to execute?
- this will simplify all aspects of API.
- you can create `make` file to create whole flow of it.
- you can suggest to create/reuse the make file to use previous commands as dependencies so that you can do total flow.


4. Rename restler project to make more cacthy and unique
- `rl` is sounds like good one but its means reinforcement learning
- rest client library: rcl
- api client tool : acl : act

6. Generally necesary commands like
- random things like : https://learning.postman.com/docs/tests-and-scripts/write-scripts/variables-list/
- {{$guid}
- {{$timestamp}
- {{$isoTimestamp}
- {{$randomUUID}
- {{$randomFirstName} or {{$random}

5. Flow
- for now this flow thing can be done with make file but later we need to decide to create native one
- flow.yaml or any name like authentication.flow.yaml
```bash
actions:
        - name: create-user
          action: request-name to take action

        - name: login
          action: login-action

flows:
    -   name: new authentication flow
        actions:
            - action: create-user
              success:
                After:
                  Env:
                    userName: Body[user][userName]
              failure:
                action: display-error

    -   name: delay
        action: {{$sleep(500)}
    -   name : display-error
        action: {{$display("this is the final error message")}}

```


5. Versions of the response so that it can be cached on the local
- res.<timestamp>.md
