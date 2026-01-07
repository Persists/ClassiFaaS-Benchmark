function terminateInstanceAfter(invocationCount, maxInvocations) {
  if (invocationCount >= maxInvocations) setImmediate(() => process.exit(1));
}

module.exports = {
  terminateInstanceAfter,
};
