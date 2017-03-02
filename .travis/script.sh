#!/usr/bin/env bash

main() {
  go build goheft.go

  if [[ $? -ne 0 ]] ; then
    exit 1
  fi

  mv goheft goheft-binary

  if [[ $(./goheft-binary goheft.go | wc -l) -lt 20 ]] ; then
    exit 1
  fi
}

main "$@"
