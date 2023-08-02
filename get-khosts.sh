#!/bin/bash

# Parse command-line arguments
namespace=""
while getopts ":n:" opt; do
  case $opt in
    n)
      namespace="$OPTARG"
      ;;
    \?)
      echo "Invalid option: -$OPTARG" >&2
      exit 1
      ;;
    :)
      echo "Option -$OPTARG requires an argument." >&2
      exit 1
      ;;
  esac
done

namespaces="$namespace"
# If the '-n' argument was not provided, get all namespaces in the cluster
if [ -z "$namespace" ]; then
  namespaces=$(kubectl get namespaces -o jsonpath='{.items[*].metadata.name}')
fi

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

    # get the deployment healthcheck URLs
    healthcheck_urls=$(kubectl describe deployment $app --namespace $namespace 2>/dev/null | grep -o 'http[s]\?://[^ ]\+')

    # get the ingress hosts
    ingress_hosts=$(kubectl get ingress --namespace $namespace -o jsonpath="{.items[*].spec.rules[*].host}" 2>/dev/null)

    # print out results
    if [ ! -z "$urls" ]; then
      echo "    Helm URLs:"
      echo "$urls" | sed 's/^/- /' | sed 's/^/      /'
    fi

    if [ ! -z "$healthcheck_urls" ]; then
      echo "    Deployment Healthcheck URLs:"
      echo "$healthcheck_urls" | sed 's/^/- /' | sed 's/^/      /'
    fi

    if [[ ! -z "$ingress_hosts" ]]; then
      echo "    Ingress Hosts:"
      # loop through each ingress host and print them on a new line
      for host in $ingress_hosts; do
        echo "      - $host"
      done
    fi
  done
done
