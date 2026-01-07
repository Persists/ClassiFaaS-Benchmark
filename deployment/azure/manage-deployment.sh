#!/bin/bash
microBenchmarks=("gemm" "sha256" "aesCtr" "gzip" "json")
instanceMemoryOptions=(512 2048)
functionRoot="src/functions"

deploy_function_app() {
    local appName=$1
    local instanceMemory=$2
    local location=$3
    local storage=$4
    local resourceGroup="b-rg-${location}"

    echo "🚀 Creating function app '$appName' with $instanceMemory MB memory..."

    az functionapp create \
        --name "$appName" \
        --storage-account "$storage" \
        --flexconsumption-location "$location" \
        --resource-group "$resourceGroup" \
        --os-type linux \
        --runtime node \
        --runtime-version 20 \
        --instance-memory "$instanceMemory" \
        --maximum-instance-count 1000 \
        --output none

    # Wait for availability
    for i in {1..10}; do
        sleep 5
        if az functionapp show --name "$appName" --resource-group "$resourceGroup" &>/dev/null; then
            echo "✅ Function app '$appName' is available."
            break
        fi
        echo "⏳ Waiting for '$appName'... retry ($i/10)"
        if [ $i -eq 10 ]; then
            echo "❌ Error: Function app '$appName' did not become available in time."
            exit 1
        fi
    done

    # One request per instance (avoid concurrency bias)
    az functionapp scale config set --resource-group "$resourceGroup" --name "$appName" --trigger-type http --trigger-settings perInstanceConcurrency=1 --output none
}

deploy() {
    local location=$1


    if ! az functionapp list-flexconsumption-locations --query "[].name" -o tsv | grep -qx "$location"; then
        echo "❌ '$location' is not a valid Azure Functions location."
        exit 1
    fi

    local resourceGroup="b-rg-${location}"
    local functionAppPrefix="b-fa-${location}"
    local randomIdentifier=$((RANDOM))$((RANDOM))
    local storage="storage${randomIdentifier}"

    echo "🛠️ Creating resource group '$resourceGroup'..."
    az group create --name "$resourceGroup" --location "$location" --output none

    echo "📦 Creating storage account '$storage'..."
    az storage account create \
        --name "$storage" \
        --location "$location" \
        --resource-group "$resourceGroup" \
        --sku Standard_LRS \
        --allow-blob-public-access false \
        --output none

    # Copy shared folder for deployment
    rm -rf ./src/shared
    cp -r ../shared ./src/shared

    echo "🚀 Deploying all function apps and benchmarks..."

    local failed=0

    for benchmark in "${microBenchmarks[@]}"; do
        jq '.main = "src/functions/'"$benchmark"'.js"' package.json > package.tmp && mv package.tmp package.json

        echo "🚀 Deploying $benchmark ..."
        pids=()
        for memory in "${instanceMemoryOptions[@]}"; do
            appName="${functionAppPrefix}-${benchmark}-${memory}mb"
            (
                deploy_function_app "$appName" "$memory" "$location" "$storage"
                echo "📤 Publishing $benchmark to $appName ..."
                func azure functionapp publish "$appName" --nozip >/dev/null
            ) &
            pids+=($!)
        done

        # Wait for all memory configs of this benchmark
        for pid in "${pids[@]}"; do
            wait "$pid" || failed=1
        done

    done

    rm -rf ./src/shared

    if [ "$failed" -ne 0 ]; then
        echo "❌ One or more functions failed to deploy."
        exit 1
    fi

    echo "✅ Deployment completed for region=$location and memories: ${memories[*]}"
}

get_urls() {
    local location=$1

    if ! az functionapp list-flexconsumption-locations --query "[].name" -o tsv | grep -qx "$location"; then
        echo "❌ '$location' is not a valid Azure Functions location."
        exit 1
    fi


    local resourceGroup="b-rg-${location}"
    local functionAppPrefix="b-fa-${location}"

    for memory in "${instanceMemoryOptions[@]}"; do
        for benchmark in "${microBenchmarks[@]}"; do
            appName="${functionAppPrefix}-${benchmark}-${memory}mb"

            funcUrl=$(az functionapp function show \
                --name "$appName" \
                --function-name "$benchmark" \
                --resource-group "$resourceGroup" \
                --query "invokeUrlTemplate" \
                --output tsv 2>/dev/null)

            if [ -z "$funcUrl" ]; then
                echo "⚠️  Could not retrieve URL for '$benchmark' in '$appName'"
                continue
            fi

            code=$(az functionapp function keys list \
                --name "$appName" \
                --function-name "$benchmark" \
                --resource-group "$resourceGroup" \
                --query "default" \
                --output tsv 2>/dev/null)

            echo "{\"benchmark\":\"$benchmark\",\"url\":\"$funcUrl\",\"auth\":\"$code\",\"memory\":$memory,\"region\":\"$location\"}"
        done
    done
}

remove() {
    local location=$1
    local resourceGroup="b-rg-${location}"

    echo "🧹 Deleting resource group '$resourceGroup'..."
    az group delete --name "$resourceGroup" -y
    echo "✅ Resource group '$resourceGroup' deleted."

    rm -rf ./src/shared
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
        echo "Usage: $0 {deploy|get-urls|delete} <location> <memory...>"
        echo "Example: $0 deploy westeurope 512 2048"
        exit 1
        ;;
esac
