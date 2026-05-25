// GET /work — Simulates real-world API behaviour for testing alarms.
//
// 70% of the time: returns 200 with variable latency (100ms - 2s)
// 15% of the time: returns 200 but slow (3-5 seconds) — triggers latency alarm
// 15% of the time: returns 500 error — triggers error rate alarm
//
// Use this endpoint to verify the monitoring template captures
// latency spikes and error rates correctly.

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

exports.handler = async (event) => {
  const roll = Math.random();

  // Slow response branch (triggers latency alarm)
  if (roll < 0.15) {
    const delay = 3000 + Math.random() * 2000;
    await sleep(delay);
    return {
      statusCode: 200,
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        status: "completed_slowly",
        latencyMs: Math.round(delay),
        timestamp: new Date().toISOString(),
      }),
    };
  }

  // Error branch (triggers error rate alarm)
  if (roll < 0.30) {
    return {
      statusCode: 500,
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        error: "InternalServerError",
        message: "Simulated upstream failure",
        timestamp: new Date().toISOString(),
      }),
    };
  }

  // Normal response (most requests)
  const delay = 100 + Math.random() * 1900;
  await sleep(delay);
  return {
    statusCode: 200,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      status: "completed",
      latencyMs: Math.round(delay),
      timestamp: new Date().toISOString(),
    }),
  };
};
