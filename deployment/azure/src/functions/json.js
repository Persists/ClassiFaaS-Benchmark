const { app } = require("@azure/functions");
const { json } = require("../factory/benchmark");

app.http("json", {
    methods: ["GET"],
    authLevel: "function",
    handler: json,
});
