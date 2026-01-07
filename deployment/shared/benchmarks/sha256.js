const { createHash, randomFillSync } = require("crypto");
const { performance } = require("perf_hooks");

function hashOnce(buffer) {
  const hash = createHash("sha256");
  hash.update(buffer);
  return hash.digest();
}

exports.runSha256Benchmark = function (iterations) {
  const buffer = Buffer.allocUnsafe(8 * 1024 * 1024);
  randomFillSync(buffer);

  // --- BENCHMARK ---
  const start = performance.now();
  for (let i = 0; i < iterations; i++) {
    hashOnce(buffer);
  }
  const end = performance.now();

  buffer.fill(0);

  return {
    type: "sha256",
    hashSizeMB: buffer.length / (1024 * 1024),
    hashTimeMs: end - start,
    iterations,
  };
};
