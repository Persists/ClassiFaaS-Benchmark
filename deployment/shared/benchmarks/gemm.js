const { performance } = require("perf_hooks");

function createMatrix(size, seed) {
  const matrix = [];
  for (let i = 0; i < size; i++) {
    matrix[i] = [];
    for (let j = 0; j < size; j++) {
      matrix[i][j] = (i + 1) * (j + 1) + seed;
    }
  }
  return matrix;
}

function multiplyMatrices(A, B) {
  const result = [];
  for (let i = 0; i < A.length; i++) {
    result[i] = [];
    for (let j = 0; j < B[0].length; j++) {
      result[i][j] = 0;
      for (let k = 0; k < A[0].length; k++) {
        result[i][j] += A[i][k] * B[k][j];
      }
    }
  }
  return result;
}

exports.runMatrixMultiplicationBenchmark = function (matrixSize) {
  if (matrixSize <= 0) return { matrixSize: 0, multiplicationTimeMs: 0 };

  const A = createMatrix(matrixSize, 42);
  const B = createMatrix(matrixSize, 99);

  // --- BENCHMARK ---
  const start = performance.now();
  multiplyMatrices(A, B);
  const end = performance.now();

  return {
    type: "gemm",
    matrixSize,
    multiplicationTimeMs: end - start,
  };
};
