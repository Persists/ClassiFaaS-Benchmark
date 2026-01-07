const { app } = require("@azure/functions");
const { gzip } = require("../factory/benchmark");

app.http("gzip", {
  methods: ["GET"],
  authLevel: "function",
  handler: gzip,
});
