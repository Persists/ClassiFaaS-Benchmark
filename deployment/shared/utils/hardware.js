const { readFileSync } = require("fs");

exports.getCpuInfo = function () {
  const data = readFileSync("/proc/cpuinfo", "utf8");
  const firstProcessor = data.split("\n\n")[0];
  const cpuFingerprint = {};

  firstProcessor.split("\n").forEach((line) => {
    const [key, value] = line.split(":").map((s) => s.trim());
    if (key && value) {
      cpuFingerprint[key] = value;
    }
  });

  return cpuFingerprint;
};
