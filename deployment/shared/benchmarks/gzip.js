const { gzipSync } = require("zlib");
const { performance } = require("perf_hooks");
const { randomFillSync } = require("crypto");

function compressOnce(buffer) {
  return gzipSync(buffer);
}

exports.runGzipBenchmark = function (iterations) {
  const buffer = Buffer.allocUnsafe(4 * 1024 * 1024);
  randomFillSync(buffer);


  // --- BENCHMARK ---
  const start = performance.now();
  for (let i = 0; i < iterations; i++) {
    compressOnce(buffer);
  }
  const end = performance.now();

  buffer.fill(0);

  return {
    type: "gzip",
    compressSizeMB: buffer.length / (1024 * 1024),
    compressTimeMS: end - start,
    iterations,
  };
};
