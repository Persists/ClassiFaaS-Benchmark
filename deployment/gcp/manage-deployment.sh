#!/bin/bash

microBenchmarks=("gemm" "sha256" "aesCtr" "gzip" "json")
instanceMemoryOptions=(128 512 2048)

deploy() {
    local location=$1
    if ! gcloud functions regions list --format="value(locationId)" | grep -qx "$location"; then
        echo "Error: '$location' is not a valid GCP Functions region."
        exit 1
    fi


    rm -rf ./shared
    cp -r ../shared ./shared

    pids=()
    for memory in "${instanceMemoryOptions[@]}"; do
        for benchmark in "${microBenchmarks[@]}"; do
            functionName="b-${benchmark}-${location}-${memory}"
            echo "→ Deploying $functionName ..."
            (
                gcloud functions deploy "$functionName" \
                    --region="$location" \
                    --runtime="nodejs20" \
                    --memory="$memory" \
                    --source="." \
                    --entry-point="$benchmark" \
                    --trigger-http \
                    --no-gen2 \
                    --no-allow-unauthenticated \
                    --quiet
            ) &
            pids+=($!)
        done
    done

    failed=0
    for pid in "${pids[@]}"; do
        wait "$pid" || failed=1
    done

    rm -rf ./shared

    if [ "$failed" -ne 0 ]; then
        echo "❌ One or more functions failed to deploy."
        exit 1
    fi

    echo "✅ Deployment completed for region=$location and memories: ${memories[*]}"
}

get_urls() {
    local location=$1
    if ! gcloud functions regions list --format="value(locationId)" | grep -qx "$location"; then
        echo "Error: '$location' is not a valid GCP Functions region."
        exit 1
    fi

    for memory in "${instanceMemoryOptions[@]}"; do
        for benchmark in "${microBenchmarks[@]}"; do
            functionName="b-${benchmark}-${location}-${memory}"
            url=$(gcloud functions describe "$functionName" --region="$location" --format="value(httpsTrigger.url)")
            if [ -z "$url" ]; then
                echo "⚠️  Could not retrieve URL for function '$functionName'."
                continue
            fi
            echo "{\"benchmark\":\"$benchmark\",\"url\":\"$url\",\"memory\":$memory,\"region\":\"$location\"}"
        done
    done
}

remove() {
    local location=$1

    pids=()
    for memory in "${instanceMemoryOptions[@]}"; do
        for benchmark in "${microBenchmarks[@]}"; do
            functionName="b-${benchmark}-${location}-${memory}"
            echo "→ Deleting $functionName ..."
            (
                gcloud functions delete "$functionName" \
                    --region="$location" \
                    --quiet
            ) &
            pids+=($!)
        done
    done

    failed=0
    for pid in "${pids[@]}"; do
        wait "$pid" || failed=1
    done

    if [ "$failed" -ne 0 ]; then
        echo "❌ One or more functions failed to delete."
        exit 1
    fi

    echo "✅ Deletion completed for region=$location and memories: ${memories[*]}"

    rm -rf ./shared
}

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
        echo "Usage: $0 {deploy|get-urls|delete} <region> <memory>"
        echo "Example: $0 deploy europe-west3 512MB"
        exit 1
        ;;
esac
