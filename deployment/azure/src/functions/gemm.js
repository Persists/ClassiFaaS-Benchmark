const { app } = require("@azure/functions");
const { gemm } = require("../factory/benchmark");

app.http("gemm", {
  methods: ["GET"],
  authLevel: "function",
  handler: gemm,
});
