#!/bin/bash
deploy() {
    local region=$1

    if [ -z "$region" ]; then
        echo "Error: region argument is required."
        exit 1
    fi

    echo "🚀 Deploying to AWS region=$region..."
    sls deploy --param="region=$region"
    echo "✅ Deployment completed for region=$region."
}

get_urls() {
    local region=$1

    info_output=$(sls info --param="region=$region" 2>/dev/null || true)
    if [ -z "$info_output" ]; then
        echo "❌ Could not retrieve sls info output."
        exit 1
    fi

    api_key=$(echo "$info_output" | awk '/apiKey:/ {print $2; exit}')
    endpoints=$(echo "$info_output" | grep -Eo "https://[^[:space:]]+")

    if [ -z "$endpoints" ]; then
        echo "⚠️  No endpoints found in sls info output."
        exit 1
    fi

    while IFS= read -r url; do
        suffix=$(echo "$url" | grep -oE '[0-9]+$')
        benchmark=$(basename "$url" | sed -E "s/[0-9]+$//")

        if [[ "$benchmark" == "sha" ]]; then
            benchmark="sha256"
        fi

        echo "{\"url\":\"$url\",\"auth\":\"$api_key\",\"memory\":$suffix,\"region\":\"$region\",\"benchmark\":\"$benchmark\"}"
    done <<< "$endpoints"
}


remove() {
    local region=$1
    if [ -z "$region" ]; then
        echo "Error: region argument is required."
        exit 1
    fi

    echo "🧹 Removing AWS deployment for region: $region"
    sls remove --param="region=$region"
    echo "✅ Removal completed for region=$region."
}

# Entrypoint
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
        echo "  $0 deploy <region> <memory...|all>"
        echo "  $0 get-urls <region> <memory...|all>"
        echo "  $0 delete <region>"
        echo
        echo "Examples:"
        echo "  $0 deploy us-east-1 128 256"
        echo "  $0 deploy us-east-1 all"
        echo "  $0 get-urls us-east-1 512"
        echo "  $0 delete us-east-1"
        exit 1
        ;;
esac
