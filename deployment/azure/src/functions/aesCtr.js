const { app } = require("@azure/functions");
const { aesCtr } = require("../factory/benchmark");

app.http("aesCtr", {
  methods: ["GET"],
  authLevel: "function",
  handler: aesCtr,
});
