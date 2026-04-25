// ============================================================
// DeadLift client-side API
// ============================================================

const PROXY_URL = 'https://deadlift-proxy-f47qsb66lq-uc.a.run.app';

const __delay = (ms = 250) => new Promise(r => setTimeout(r, ms));

// ---------- AUTH ----------
window.api = window.api || {};

window.api.signIn = async ({ email, password }) => {
  await __delay(400);
  // TODO: POST /api/auth/signin
  return { ok: true, user: { email, name: email.split('@')[0], org: 'acme-payments' } };
};

window.api.signUp = async ({ email, password }) => {
  await __delay(400);
  // TODO: POST /api/auth/signup
  return { ok: true, user: { email, name: email.split('@')[0], org: null } };
};

window.api.signOut = async () => {
  await __delay(150);
  // TODO: POST /api/auth/signout
  return { ok: true };
};

// ---------- ONBOARDING ----------
window.api.connectGCP = async ({ projectId }) => {
  await __delay(700);
  // TODO: kick off OAuth → /auth/google handled by Go service
  return { ok: true, projectId, account: 'sre@acme.com' };
};

window.api.listPubsubResources = async ({ projectId }) => {
  await __delay(500);
  // TODO: GET /api/gcp/{projectId}/pubsub
  return {
    topics: [
      'payments-events', 'payments-events-dlq',
      'orders-fulfillment', 'orders-fulfillment-dlq',
      'inventory-sync', 'inventory-sync-dlq',
      'notifications-out', 'notifications-out-dlq',
    ],
    subscriptions: [
      'payments-events-sub', 'payments-events-dlq-sub',
      'orders-fulfillment-sub', 'orders-fulfillment-dlq-sub',
      'inventory-sync-sub', 'inventory-sync-dlq-sub',
      'notifications-out-sub', 'notifications-out-dlq-sub',
    ],
  };
};

window.api.uploadContext = async (file) => {
  await __delay(800);
  // TODO: POST multipart /api/context/upload → GraphRAG ingestor
  return { ok: true, id: 'ctx_' + Math.random().toString(36).slice(2, 9), name: file.name, size: file.size };
};

window.api.startOnboarding = async (config) => {
  const autoRepublish = {};
  const CATS = ['Schema drift', 'Malformed JSON', 'Type mismatch', 'Missing field', 'Encoding', 'Downstream outage'];
  CATS.forEach(cat => {
    autoRepublish[cat] = (config.categoryOverrides[cat] || config.approvalMode) === 'auto';
  });
  const res = await fetch(`${PROXY_URL}/api/onboard/connect`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      project_id: config.projectId,
      dlq_subscription: config.dlqSub,
      main_topic: config.mainTopic,
      notification_email: config.notifications?.email || '',
      batching_threshold: config.batchThreshold,
      auto_republish: autoRepublish,
    }),
  });
  if (!res.ok) throw new Error(await res.text());
  return res.json(); // { oauth_url: '...' }
};

// ---------- DASHBOARD ----------
window.api.getFixes = async () => {
  await __delay(300);
  return MOCK_FIXES;
};

window.api.approveFix = async (id) => {
  await __delay(300);
  return { ok: true, id, status: 'fixed' };
};

window.api.denyFix = async (id, reason) => {
  await __delay(200);
  return { ok: true, id, status: 'denied' };
};

window.api.getBatches = async () => {
  await __delay(300);
  return MOCK_BATCHES;
};

window.api.approveBatch = async (id) => {
  await __delay(400);
  return { ok: true, id, status: 'fixed' };
};

window.api.getAnalytics = async () => {
  await __delay(300);
  return MOCK_ANALYTICS;
};

window.api.getRCAReports = async () => {
  await __delay(300);
  return MOCK_RCA;
};

window.api.terminateService = async () => {
  await __delay(900);
  return { ok: true };
};

// ============================================================
// MOCK DATA
// ============================================================

const MOCK_FIXES = [
  {
    id: 'fix_8a3c91',
    status: 'pending', // pending | fixed | denied
    category: 'Schema drift',
    confidence: 0.94,
    subscription: 'payments-events-dlq-sub',
    topic: 'payments-events',
    receivedAt: '2 min ago',
    error: 'TypeError: Cannot read property "amount_cents" of undefined at processCharge (charge.ts:42:18)',
    batch: { count: 47, label: 'producer v3.2.1 schema rollout' },
    before: `{
  "event": "charge.completed",
  "id": "ch_2nL9KQ8aBcDeFgH",
  "amount": 4999,
  "currency": "usd",
  "customer_id": "cus_NKpL2x8aBcD",
  "metadata": { "order_id": "ord_8839" }
}`,
    after: `{
  "event": "charge.completed",
  "id": "ch_2nL9KQ8aBcDeFgH",
  "amount_cents": 4999,
  "currency": "usd",
  "customer_id": "cus_NKpL2x8aBcD",
  "metadata": { "order_id": "ord_8839" }
}`,
    sources: [
      { kind: 'runbook', name: 'payments-runbook.md §4.2', line: 'amount renamed to amount_cents in v3.2.0' },
      { kind: 'code', name: 'src/handlers/charge.ts:42', line: 'destructure amount_cents from event' },
      { kind: 'log', name: 'cloud-logging 14:22:08 UTC', line: 'TypeError on undefined property amount_cents' },
    ],
  },
  {
    id: 'fix_4e72bd',
    status: 'fixed',
    category: 'Malformed JSON',
    confidence: 0.99,
    subscription: 'orders-fulfillment-dlq-sub',
    topic: 'orders-fulfillment',
    receivedAt: '11 min ago',
    fixedAt: '11 min ago',
    error: 'json.parse failed at position 287: unexpected token "," at end of object',
    batch: null,
    before: `{
  "order_id": "ord_8927",
  "items": [
    { "sku": "SKU-887", "qty": 2, },
    { "sku": "SKU-441", "qty": 1, },
  ],
  "total_cents": 12998,
}`,
    after: `{
  "order_id": "ord_8927",
  "items": [
    { "sku": "SKU-887", "qty": 2 },
    { "sku": "SKU-441", "qty": 1 }
  ],
  "total_cents": 12998
}`,
    sources: [
      { kind: 'log', name: 'cloud-logging 14:13 UTC', line: 'JSON parse error at offset 287' },
    ],
  },
  {
    id: 'fix_2fa1cc',
    status: 'pending',
    category: 'Type mismatch',
    confidence: 0.87,
    subscription: 'inventory-sync-dlq-sub',
    topic: 'inventory-sync',
    receivedAt: '23 min ago',
    error: 'ValidationError: in_stock expected boolean, got string "true"',
    batch: null,
    before: `{
  "sku": "SKU-19284",
  "warehouse": "us-west-2",
  "in_stock": "true",
  "qty_on_hand": "143"
}`,
    after: `{
  "sku": "SKU-19284",
  "warehouse": "us-west-2",
  "in_stock": true,
  "qty_on_hand": 143
}`,
    sources: [
      { kind: 'code', name: 'proto/inventory.proto:14', line: 'in_stock bool, qty_on_hand int32' },
      { kind: 'context', name: 'graphrag/inventory-schema', line: 'producer-side serializer is stringifying primitives' },
    ],
  },
  {
    id: 'fix_9c14ee',
    status: 'fixed',
    category: 'Missing field',
    confidence: 0.91,
    subscription: 'notifications-out-dlq-sub',
    topic: 'notifications-out',
    receivedAt: '38 min ago',
    fixedAt: '38 min ago',
    error: 'required field "channel" missing',
    batch: null,
    before: `{
  "user_id": "usr_8a2b",
  "template": "order_shipped",
  "params": { "tracking": "1Z999AA1" }
}`,
    after: `{
  "user_id": "usr_8a2b",
  "template": "order_shipped",
  "channel": "email",
  "params": { "tracking": "1Z999AA1" }
}`,
    sources: [
      { kind: 'context', name: 'past 7d telemetry', line: 'usr_8a2b has email channel preference (98% of prior sends)' },
    ],
  },
  {
    id: 'fix_77adb1',
    status: 'pending',
    category: 'Encoding',
    confidence: 0.96,
    subscription: 'payments-events-dlq-sub',
    topic: 'payments-events',
    receivedAt: '1 hr ago',
    error: 'UTF-8 decode error: invalid byte 0xC3 at position 142',
    batch: null,
    before: `{ "customer_name": "Andr\\xe9 Müller", "country": "DE" }`,
    after: `{ "customer_name": "André Müller", "country": "DE" }`,
    sources: [{ kind: 'log', name: 'cloud-logging 13:42 UTC', line: 'invalid utf-8 at byte 142' }],
  },
];

const MOCK_BATCHES = [
  {
    id: 'batch_a1',
    title: 'producer v3.2.1 schema rollout',
    rootCause: 'Producer service "payments-api" deployed v3.2.1 at 14:18 UTC, renaming `amount` → `amount_cents`. Consumer "ledger-worker" still on v3.1.0.',
    affectedCount: 47,
    affectedTopics: ['payments-events'],
    category: 'Schema drift',
    confidence: 0.94,
    status: 'pending',
    firstSeen: '14:22 UTC',
    fixSummary: 'Rename `amount` → `amount_cents` in all queued messages.',
  },
  {
    id: 'batch_a2',
    title: 'inventory producer serializing primitives as strings',
    rootCause: 'After commit `e8a31c` to `inventory-producer`, all bool/int fields are being JSON.stringify-ed twice.',
    affectedCount: 18,
    affectedTopics: ['inventory-sync'],
    category: 'Type mismatch',
    confidence: 0.87,
    status: 'pending',
    firstSeen: '13:58 UTC',
    fixSummary: 'Cast `in_stock` to bool, `qty_on_hand` to int across batch.',
  },
  {
    id: 'batch_a3',
    title: 'fulfillment-svc trailing-comma serialization bug',
    rootCause: 'Older Go binary on `fulfillment-svc-v2` is using a custom JSON encoder that emits trailing commas. PR #4421 fixes upstream.',
    affectedCount: 12,
    affectedTopics: ['orders-fulfillment'],
    category: 'Malformed JSON',
    confidence: 0.99,
    status: 'fixed',
    firstSeen: '13:14 UTC',
    fixSummary: 'Strip trailing commas, re-emit valid JSON.',
  },
];

const MOCK_ANALYTICS = {
  kpis: {
    dlqVolume24h: 1284,
    autoFixed: 891,
    awaitingApproval: 73,
    unfixable: 320,
    mttrBefore: '47m',
    mttrAfter: '2.4m',
    mttrDelta: -94,
    estSavings30d: 287400,
  },
  // 24h hourly volume (DLQ in / fixed / unfixable)
  series: Array.from({ length: 24 }, (_, i) => {
    const hour = i;
    const base = 30 + Math.sin(i / 3) * 18 + Math.cos(i / 5) * 10;
    const dlq = Math.max(8, Math.round(base + (i === 14 ? 95 : 0) + Math.random() * 8));
    const fixed = Math.round(dlq * (0.62 + Math.random() * 0.18));
    const awaiting = Math.round(dlq * 0.06);
    const unfixable = dlq - fixed - awaiting;
    return { hour, dlq, fixed, awaiting, unfixable };
  }),
  categories: [
    { name: 'Schema drift', count: 412, pct: 32, color: 'oklch(0.78 0.17 145)' },
    { name: 'Malformed JSON', count: 318, pct: 25, color: 'oklch(0.72 0.14 240)' },
    { name: 'Type mismatch', count: 207, pct: 16, color: 'oklch(0.82 0.16 78)' },
    { name: 'Missing field', count: 152, pct: 12, color: 'oklch(0.72 0.16 290)' },
    { name: 'Encoding', count: 95, pct: 7, color: 'oklch(0.68 0.18 25)' },
    { name: 'Downstream outage', count: 72, pct: 6, color: 'oklch(0.62 0.04 240)' },
    { name: 'Other', count: 28, pct: 2, color: 'oklch(0.55 0.02 240)' },
  ],
  topics: [
    { name: 'payments-events', dlq: 487, fixed: 412, mttr: '1.8m' },
    { name: 'orders-fulfillment', dlq: 318, fixed: 261, mttr: '2.1m' },
    { name: 'inventory-sync', dlq: 211, fixed: 157, mttr: '3.4m' },
    { name: 'notifications-out', dlq: 184, fixed: 168, mttr: '1.2m' },
    { name: 'analytics-events', dlq: 84, fixed: 71, mttr: '4.7m' },
  ],
};

const MOCK_RCA = [
  {
    id: 'rca_1',
    title: 'producer v3.2.1 schema rollout',
    runAt: '14:24 UTC',
    affected: 47,
    confidence: 0.94,
    summary: 'A backward-incompatible deploy of `payments-api` at 14:18 UTC renamed the `amount` field to `amount_cents`. The consumer `ledger-worker` is still pinned to v3.1.0 and crashes on the missing field.',
    hypotheses: [
      {
        id: 'h1',
        text: 'Producer schema change (amount → amount_cents)',
        confidence: 0.94,
        winner: true,
        evidence: [
          { kind: 'deploy', label: 'payments-api deploy', detail: 'v3.1.0 → v3.2.1 at 14:18:02 UTC (Cloud Build #8821)', t: '14:18 UTC' },
          { kind: 'log', label: 'first DLQ entry', detail: 'TypeError on undefined property `amount_cents` at 14:22:08', t: '14:22 UTC' },
          { kind: 'code', label: 'PR #2204', detail: 'feat(payments): rename amount → amount_cents (#2204)', t: '14:11 UTC' },
          { kind: 'doc', label: 'payments-runbook.md §4.2', detail: 'Note: amount renamed to amount_cents in v3.2.0' },
        ],
      },
      {
        id: 'h2',
        text: 'Consumer rollback in progress',
        confidence: 0.18,
        evidence: [{ kind: 'log', label: 'no rollback events', detail: 'No deploy events for ledger-worker in the past 6 hours.' }],
      },
      {
        id: 'h3',
        text: 'Network partition / broker hiccup',
        confidence: 0.04,
        evidence: [{ kind: 'metric', label: 'broker p99 latency', detail: '14ms (nominal). No partition events.' }],
      },
    ],
    recommendation: 'Apply the proposed batch fix (rename field) and pin ledger-worker upgrade to next release window. Add a schema gate to CI.',
  },
  {
    id: 'rca_2',
    title: 'inventory producer serializing primitives as strings',
    runAt: '14:01 UTC',
    affected: 18,
    confidence: 0.87,
    summary: 'Commit `e8a31c` to `inventory-producer` introduced a double JSON.stringify call. All boolean / int fields now arrive as strings.',
    hypotheses: [
      {
        id: 'h1',
        text: 'Double-stringify regression',
        confidence: 0.87,
        winner: true,
        evidence: [
          { kind: 'code', label: 'commit e8a31c', detail: 'inventory-producer: cleanup helpers — accidentally added second JSON.stringify wrap', t: '13:42 UTC' },
          { kind: 'log', label: 'validation errors', detail: '18 messages, all with `in_stock: "true"` and `qty_on_hand: "143"`' },
        ],
      },
      { id: 'h2', text: 'Schema downgrade on consumer', confidence: 0.10, evidence: [{ kind: 'metric', label: 'consumer version', detail: 'unchanged for 7 days' }] },
    ],
    recommendation: 'Revert e8a31c on inventory-producer. Apply batch fix to drain queue. Add type-roundtrip test to producer CI.',
  },
];

// ============================================================
// SESSION (in-memory; replace with cookie/JWT)
// ============================================================
window.session = window.session || {
  user: null,
  config: null,
  setUser(u) { this.user = u; },
  setConfig(c) { this.config = c; },
};
