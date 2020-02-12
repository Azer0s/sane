[![Github Action](https://github.com/Azer0s/sane/workflows/Go/badge.svg)](https://github.com/Azer0s/sane/actions?query=workflow%3AGo) [![Go Report Card](https://goreportcard.com/badge/github.com/Azer0s/sane)](https://goreportcard.com/report/github.com/Azer0s/sane) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/Azer0s/sane/blob/master/LICENSE.md)

# sane

A package manager for sane configurations

## apply package list

`sane` supports aliasing. By pulling a list of aliases, one doesn't need to type out the full name of the repo.

```bash
sane apply Azer0s/config/index
```

## apply settings

```bash
sane apply vimsettings
```

## remove settings

```bash
sane remove vimsettings
```

## start docker containers

```bash
sane start kafka
```

## stop docker containers

```bash
sane stop kafka
```
