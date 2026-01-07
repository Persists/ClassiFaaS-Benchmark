
exports.seededRandomBuffer = function (sizeMB, seed = 0x12345678) {
    const buffer = Buffer.allocUnsafe(sizeMB * 1024 * 1024);
    let x = seed >>> 0;
    for (let i = 0; i < buffer.length; i++) {
        // pseudo-random generator (LCG)
        x = (1664525 * x + 1013904223) >>> 0;
        buffer[i] = x & 0xff;
    }
    return buffer;
}
