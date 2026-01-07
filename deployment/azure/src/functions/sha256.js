const { app } = require("@azure/functions");
const { sha256 } = require("../factory/benchmark");

app.http("sha256", {
  methods: ["GET"],
  authLevel: "function",
  handler: sha256,
});
