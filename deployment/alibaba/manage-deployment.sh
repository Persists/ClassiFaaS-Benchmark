#!/bin/bash

deploy() {
    local location="$1"
    if [ -z "$location" ]; then
        echo "Error: region argument is required."
        exit 1
    fi

    echo "🚀 Deploying functions to Alibaba Cloud region=$location"
    rm -rf ./code/shared
    cp -rf ../shared ./code/shared

    export REGION="$location"
    s deploy -y -o raw

    rm -rf ./code/shared
    echo "✅ Deployment completed for $location."
}

get_urls() {
    local location=$1
    export REGION="$location"

    info_output=$(s info --silent -o raw 2>/dev/null)
    if [ -z "$info_output" ]; then
        echo "❌ Could not retrieve function info."
        exit 1
    fi


    jq -r --arg region "$location" '
        to_entries[]
        | {
            url: .value.url.system_url,
            intranet_url: .value.url.system_intranet_url,
            memory: .value.memorySize,
            region: $region,
            benchmark: (.key | split("-")[1])
        }
        | select(.url != null)
        | if .benchmark == "sha" then .benchmark = "sha256" else . end
        | @json
    ' <<< "$info_output"
}


remove() {
    local location="$1"
    if [ -z "$location" ]; then
        echo "Error: region argument is required."
        exit 1
    fi

    echo "🧹 Removing functions from region=$location ..."
    export REGION="$location"
    s remove -y --silent
    echo "✅ Removal completed for $location."

    rm -rf ./code/shared
}

# Main entrypoint
case "$1" in
    deploy)
        shift
        deploy "$@"
        ;;
    get-urls)
        shift
        get_urls "$@"
        ;;
    remove|delete)
        shift
        remove "$@"
        ;;
    *)
        echo "Usage:"
        echo "  $0 deploy <region>"
        echo "  $0 get-urls <region> <memory...|all>"
        echo "  $0 delete <region>"
        echo
        echo "Examples:"
        echo "  $0 deploy cn-hangzhou"
        echo "  $0 get-urls cn-hangzhou 128 256"
        echo "  $0 get-urls cn-hangzhou all"
        echo "  $0 delete cn-hangzhou"
        exit 1
        ;;
esac
