# Go Compose
![Build](https://github.com/hadialqattan/go-compose/workflows/Build/badge.svg?branch=main)
![LICENSE](https://img.shields.io/badge/License-MIT-darkgreen.svg)


A lightweight services composer written in Golang for managing services (processes) in a development environment.

***
## Installation

```bash
$ go get github.com/hadialqattan/go-compose
```

## Usage

```bash
$ go-compose start # --config path/to/go-compose.yaml [default=./go-compose.yaml].
```

## `go-compose.yaml`

```yaml
services:
  service_name:
    ignore_failures: false
    sub_service: false
    auto_restart: true
    hooks:
        wait: 
            - another_service_name
        stop:
            - another_service_name
        start:
            - another_service_name
    cwd: .
    command: echo "shell script"
    environs:
      ENVIRONMENT_VARIABLE: true
```

* `services`: a set of services to run.
* `ignore_failures`: don't stop other services when this failed.
* `sub_service`: don't start this service. this may be started by another service using the `start hook`.
* `auto_restart`: automatically restart this service if it crashed.
* `hooks`:
    + `wait`: wait for other services to stop before starting (setup).
    + `stop`: stop other services on exit (teardown).
    + `start`: start other services/sub-services on exit (teardown).
* `cwd`: where the command will be executed (Command Work Directory).
* `command`: a unix shell command.
* `environs`: environment variables.

A real world example:

```yaml
services:

  webapp:
    ignore_failures: true
    auto_restart: true
    hooks:
      start:
        - cleanup
    cwd: .
    command: |
      docker build .
      docker run --rm \
        --name webapp \
        --link db \
        -v ${PWD}:/webapp -w /webapp webapp
    environs:
      - POSTGRES_USER=webapp_admin
      - POSTGRES_PASSWORD=db0123
      - POSTGRES_DB=db

  postgres_db:
    ignore_failures: true
    auto_restart: true
    cwd: .
    command: |
      sleep 3
      docker run --rm \
        --name db postgres:12.0-alpine
    environs:
      - POSTGRES_USER=webapp_admin
      - POSTGRES_PASSWORD=db0123
      - POSTGRES_DB=db

  cleanup:
    sub_service: true
    cwd: .
    command: python3 teardown.py
```

***
## Copyright ©

👤 **Hadi Alqattan**

* Github: [@hadialqattan](https://github.com/hadialqattan)
* Email: [alqattanhadizaki@gmail.com](<mailto:alqattanhadizaki@gmail.com>)

📝 **License**

Copyright © 2020 [Hadi Alqattan](https://github.com/hadialqattan).<br />
This project is [MIT](https://github.com/hadialqattan/go-compose/blob/master/LICENSE) licensed.

***
Give a ⭐️ if this project helped you!
