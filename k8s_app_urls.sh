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

    # get the deployment healthcheck URLs
    healthcheck_urls=$(kubectl describe deployment $app --namespace $namespace 2>/dev/null | grep -o 'http[s]\?://[^ ]\+')

    # get the ingress hosts
    ingress_hosts=$(kubectl get ingress --namespace $namespace $app -o jsonpath="{.spec.rules[*].host}" 2>/dev/null)

    # print out results
    if [ ! -z "$urls" ]; then
      echo "    Helm URLs:"
      echo "$urls" | sed 's/^/- /' | sed 's/^/      /'
    fi

    if [ ! -z "$healthcheck_urls" ]; then
      echo "    Deployment Healthcheck URLs:"
      echo "$healthcheck_urls" | sed 's/^/- /' | sed 's/^/      /'
    fi

    if [[ ! -z "${ingress_hosts// }" ]]; then
      echo "    Ingress Hosts:"
      echo "$ingress_hosts" | sed 's/ /\n/g' | sed 's/^/      - /'
    fi
  done
done
