#!/bin/bash

# get all namespaces in the cluster
namespaces=$(kubectl get namespaces -o jsonpath='{.items[*].metadata.name}')

# loop through each namespace
for namespace in $namespaces; do
  echo "Namespace: $namespace"

  # get all Helm applications in the namespace
  helm_apps=$(helm list --namespace $namespace --short)

  # loop through each Helm application
  for app in $helm_apps; do
    echo "  Helm App: $app"

    # get the notes for the Helm application
    notes=$(helm get notes $app --namespace $namespace)

    # filter out the URLs from the notes
    urls=$(echo "$notes" | grep -o 'http[s]\?://[^ ]\+')

    # print out the URLs
    if [ ! -z "$urls" ]; then
      echo "    URLs:"
      echo "$urls" | sed 's/^/      /'
    fi
  done
done
