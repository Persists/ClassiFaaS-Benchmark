const Inspector = require("../shared/utils/inspector");
const { terminateInstanceAfter } = require("../shared/utils/terminator");
const uuidv4 = require("uuid/v4");

var invocationCount = 0;
var instanceId = uuidv4();


function extractParameter(request, defaultValue) {
  const url = new URL(request.url);
  const param = url.searchParams.get("parameter");
  return param ? parseInt(param, 10) : defaultValue;
}

function benchmarkFactory({ benchmarkFn, defaultParam }) {
  return async (request, context) => {
    const inspector = new Inspector();
    inspector.inspectAll();

    invocationCount++;
    terminateInstanceAfter(invocationCount, 4);

    inspector.addAttribute("provider", "azure");
    inspector.addAttribute("instanceId", instanceId);
    inspector.addAttribute("invocationCount", invocationCount);

    const parameter = extractParameter(request, defaultParam);
    const benchMetrics = benchmarkFn(parameter);

    inspector.addAttribute("benchmark", benchMetrics);
    inspector.inspectAllDeltas();

    return {
      status: 200, headers: {
        "Content-Type": "application/json",
        "azure-invocation-id": context.invocationId,
      }, body: JSON.stringify(inspector.finish())
    };
  };
}

module.exports = {
  gemm: benchmarkFactory({
    benchmarkFn: require("../shared/benchmarks/gemm")
      .runMatrixMultiplicationBenchmark,
    defaultParam: 100,
  }),
  sha256: benchmarkFactory({
    benchmarkFn: require("../shared/benchmarks/sha256").runSha256Benchmark,
    defaultParam: 2,
  }),
  aesCtr: benchmarkFactory({
    benchmarkFn: require("../shared/benchmarks/aesCtr").runAesCtrBenchmark,
    defaultParam: 2,
  }),
  gzip: benchmarkFactory({
    benchmarkFn: require("../shared/benchmarks/gzip").runGzipBenchmark,
    defaultParam: 2,
  }),
  json: benchmarkFactory({
    benchmarkFn: require("../shared/benchmarks/json").runJsonBenchmark,
    defaultParam: 500,
  }),
};
