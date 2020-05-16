<br />
<p align="center">
  <h3 align="center">go-lock</h3>

  <p align="center">
    gRPC server providing DLM functionality using the redlock algorithm. (https://redis.io/topics/distlock)
    <br />
    <br />
    <a href="https://github.com/stoex/go-lock/issues">Report Bug</a>
    ·
    <a href="https://github.com/stoex/go-lock/issues">Request Feature</a>
  </p>

## Table of Contents

* [Built With](#built-with)
* [Getting Started](#getting-started)
  * [Prerequisites](#prerequisites)
  * [Installation](#installation)
  * [Configuration](#configuration)
* [Usage](#usage)
* [License](#license)
* [Contact](#contact)

## Built With

* Golang
* gRPC
* Redis

## Getting Started

To get a local copy up and running follow these simple steps.

### Prerequisites

This project provides a **Makefile** for testing / generating / building the project.

### Installation
 
1. Clone the repo
```sh
git clone https://github.com/stoex/go-lock.git
```
2. Install dependencies
```sh
make deps
```
3. Generate Code
```sh
make gen
```
4. Start up a redis instance
```sh
make redis
``` 
5. Build the project
```sh
make all
```

Please see `make help` for a list of available commands.

### Configuration

The program uses environment variables and CLI flags for configuration. To specify which redis clients should be used to connect please specify the following variable:

```sh
REDIS_CLIENTS=redis://...,redis://...,redis://...
```

The variable should be used as a comma separated list - you can specify however many clients you wish to connect to.
During development you can also use a `.env` file like so:

```sh
echo 'REDIS_CLIENTS=...' > .env
``` 

The `.env` file should be placed in the same directory the server is run from.

## Usage

See the servers available parameters with `go-lock -h`.

> Note: if `-cert_file` or `-key_file` are not used in conjuction with `-tls`the server uses certificates from `google.golang.org/grpc/testdata` instead.

## Contributing

Contributions are what make the open source community such an amazing place to be learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

Distributed under the MIT License. See `LICENSE` for more information.

## Contact

Chris Nicola - [@chrisnicola_](https://twitter.com/chrisnicola_) - hi@chrisnicola.de 

Project Link: [https://github.com/stoex/go-lock](https://github.com/stoex/go-lock)
