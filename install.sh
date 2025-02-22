#!/bin/bash

# if curl doesn't exist then use wget

if [ -x "$(command -v curl)" ]; then
    # download ing binary from releases page
elif [ -x "$(command -v wget)" ]; then

else
  printf "curl or wget are required to install fluffyproxy.\nYou can also download the latest release from the releases page."
  exit 1
fi


