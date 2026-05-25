// GET /hello — Always responds 200 with a friendly message.
// Use this endpoint to verify the monitoring template captures successful requests.

exports.handler = async (event) => {
  return {
    statusCode: 200,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      message: "Hello from example API",
      service: "example-api",
      timestamp: new Date().toISOString(),
      region: process.env.AWS_REGION || "unknown",
    }),
  };
};
