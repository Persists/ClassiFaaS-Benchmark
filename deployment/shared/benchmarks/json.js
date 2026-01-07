const { performance } = require("perf_hooks");

// Helper to generate a deeply nested complex object
function generateComplexObject(depth, breadth) {
    if (depth === 0) return "Leaf string data " + Math.random();

    const obj = {};
    for (let i = 0; i < breadth; i++) {
        obj[`key_${i}`] = generateComplexObject(depth - 1, breadth);
    }
    obj.id = Math.random();
    obj.isActive = true;
    obj.tags = [1, 2, 3, "tag"];
    return obj;
}

exports.runJsonBenchmark = function (iterations = 500) {
    const data = generateComplexObject(5, 4);

    // --- BENCHMARK ---
    const start = performance.now();
    let totalLength = 0;

    for (let i = 0; i < iterations; i++) {
        // Stringify the object
        const str = JSON.stringify(data);
        totalLength += str.length;

        // parse it back
        const obj = JSON.parse(str);
        if (obj.id === undefined) {
            throw new Error("Parsing failed");
        }
    }
    const end = performance.now();

    return {
        type: "json",
        throughputMBps: (totalLength / (1024 * 1024)) / ((end - start) / 1000),
        jsonTimeMs: end - start,
        iterations,
    };
};