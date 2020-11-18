#!/usr/bin/env bash

main() {
  go build goheft.go

  if [[ $? -ne 0 ]] ; then
    exit 1
  fi

  mv goheft goheft-binary

  ./goheft-binary --raw goheft.go

  if [[ $? -ne 0 ]] ; then
    exit 1
  fi
}

main "$@"
