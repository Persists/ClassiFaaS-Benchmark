const { createCipheriv, randomFillSync, randomBytes } = require("crypto");
const { performance } = require("perf_hooks");

function encryptOnce(buffer, key, iv, keySize) {
  const cipher = createCipheriv(`aes-${keySize}-ctr`, key, iv);
  cipher.update(buffer);
  cipher.final();
}

exports.runAesCtrBenchmark = function (iterations, keySize = 128) {
  const buffer = Buffer.allocUnsafe(8 * 1024 * 1024);
  randomFillSync(buffer);

  const key = randomBytes(keySize / 8);
  const iv = Buffer.alloc(16, 0);

  const start = performance.now();
  for (let i = 0; i < iterations; i++) {
    encryptOnce(buffer, key, iv, keySize);
  }
  const end = performance.now();

  buffer.fill(0);

  return {
    type: "aesCtr",
    encryptSizeMB: buffer.length / (1024 * 1024),
    encryptTimeMs: end - start,
    keySize,
    iterations,
  };
};
