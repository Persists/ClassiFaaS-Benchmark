"use strict";

const { terminateInstanceAfter } = require("./shared/utils/terminator");
const Inspector = require("./shared/utils/inspector");
const uuidv4 = require("uuid/v4");

var invocationCount = 0;
var instanceId = uuidv4();

function extractParameter(event, defaultValue) {
  try {
    const parsedEvent = JSON.parse(event.toString());
    if (parsedEvent.queryParameters && parsedEvent.queryParameters.parameter) {
      return parseInt(parsedEvent.queryParameters.parameter, 10);
    }
  } catch (e) { }
  return defaultValue;
}

function benchmarkFactory({ benchmarkFn, defaultParam }) {
  return async (event, context, callback) => {
    const inspector = new Inspector();
    inspector.inspectAll();

    invocationCount++;
    terminateInstanceAfter(invocationCount, 4);

    inspector.addAttribute("provider", "alibaba");
    inspector.addAttribute("instanceId", instanceId);
    inspector.addAttribute("invocationCount", invocationCount);

    const parameter = extractParameter(event, defaultParam);
    const benchMetrics = benchmarkFn(parameter);

    inspector.addAttribute("benchmark", benchMetrics);
    inspector.inspectAllDeltas();

    callback(null, {
      statusCode: 200,
      headers: {
        "Content-Type": "application/json",
        "ali-request-id": context.requestId,
      },
      body: JSON.stringify(inspector.finish()),
    });
  };
}

const gemm = benchmarkFactory({
  benchmarkFn: require("./shared/benchmarks/gemm")
    .runMatrixMultiplicationBenchmark,
  defaultParam: 100,
});

const sha256 = benchmarkFactory({
  benchmarkFn: require("./shared/benchmarks/sha256").runSha256Benchmark,
  defaultParam: 2,
});

const aesCtr = benchmarkFactory({
  benchmarkFn: require("./shared/benchmarks/aesCtr").runAesCtrBenchmark,
  defaultParam: 2,
});

const gzip = benchmarkFactory({
  benchmarkFn: require("./shared/benchmarks/gzip").runGzipBenchmark,
  defaultParam: 2,
});

const json = benchmarkFactory({
  benchmarkFn: require("./shared/benchmarks/json").runJsonBenchmark,
  defaultParam: 500,
});

module.exports = {
  gemm,
  sha256,
  aesCtr,
  gzip,
  json,
};
