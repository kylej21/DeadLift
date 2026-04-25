// ============================================================
// Onboarding — 7-step stepper
// ============================================================

const ONBOARD_STEPS = [
  { key: 'gcp', label: 'Connect GCP' },
  { key: 'pubsub', label: 'Select Pub/Sub' },
  { key: 'context', label: 'Upload context' },
  { key: 'approval', label: 'Approval mode' },
  { key: 'batching', label: 'Batching' },
  { key: 'notify', label: 'Notifications' },
  { key: 'review', label: 'Review & finish' },
];

const Onboarding = () => {
  const [step, setStep] = React.useState(0);
  const [cfg, setCfg] = React.useState({
    projectId: '', dlqSub: '', mainTopic: '',
    files: [], githubUrl: '', webUrl: '',
    approvalMode: 'human', categoryOverrides: {},
    batchThreshold: 5,
    notifications: { slack: '', email: window.session.user?.email || '', pagerduty: '' },
  });
  const upd = (patch) => setCfg(c => ({ ...c, ...patch }));
  const next = () => setStep(s => Math.min(s + 1, ONBOARD_STEPS.length - 1));
  const prev = () => setStep(s => Math.max(s - 1, 0));

  const finish = async () => {
    await window.api.completeOnboarding(cfg);
    window.session.setConfig(cfg);
    location.hash = '#/app';
  };

  const panels = [
    <StepGCP cfg={cfg} upd={upd} />,
    <StepPubSub cfg={cfg} upd={upd} />,
    <StepContext cfg={cfg} upd={upd} />,
    <StepApproval cfg={cfg} upd={upd} />,
    <StepBatching cfg={cfg} upd={upd} />,
    <StepNotify cfg={cfg} upd={upd} />,
    <StepReview cfg={cfg} onFinish={finish} />,
  ];

  return (
    <div style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
      <TopNav variant="public" />
      <div className="container-narrow" style={{ flex: 1, padding: '48px 24px 80px', display: 'flex', gap: 40 }}>
        {/* Sidebar stepper */}
        <div style={{ flex: '0 0 200px' }}>
          <div className="eyebrow" style={{ marginBottom: 18 }}>Setup</div>
          {ONBOARD_STEPS.map((s, i) => (
            <div key={s.key} onClick={() => i < step && setStep(i)}
              style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '9px 0', cursor: i < step ? 'pointer' : 'default', borderLeft: '2px solid', borderColor: i === step ? 'var(--text)' : i < step ? 'var(--green)' : 'var(--line)', paddingLeft: 14, transition: 'all 150ms ease' }}>
              <span style={{ width: 22, height: 22, borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 11, fontWeight: 600,
                background: i < step ? 'var(--green-bg)' : i === step ? 'var(--surface-3)' : 'transparent',
                border: `1px solid ${i < step ? 'var(--green-line)' : i === step ? 'var(--line-strong)' : 'var(--line)'}`,
                color: i < step ? 'var(--green)' : i === step ? 'var(--text)' : 'var(--text-4)' }}>
                {i < step ? <CheckIcon size={11} color="var(--green)" /> : i + 1}
              </span>
              <span style={{ fontSize: 13, color: i === step ? 'var(--text)' : i < step ? 'var(--text-2)' : 'var(--text-4)', fontWeight: i === step ? 500 : 400 }}>{s.label}</span>
            </div>
          ))}
        </div>
        {/* Panel */}
        <div style={{ flex: 1, minWidth: 0 }} className="fade-in" key={step}>
          {panels[step]}
          {step < ONBOARD_STEPS.length - 1 && (
            <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 32 }}>
              <button className="btn" onClick={prev} disabled={step === 0}>Back</button>
              <button className="btn btn-primary" onClick={next}>Continue <ArrowIcon /></button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

// ---- individual steps ----
const StepGCP = ({ cfg, upd }) => (
  <div>
    <h2 className="h3" style={{ marginBottom: 6 }}>Connect your GCP project</h2>
    <p className="muted-2" style={{ fontSize: 14, marginBottom: 24 }}>We'll use OAuth to grant a least-privilege service account. Your token is never stored.</p>
    <label className="lbl">GCP Project ID</label>
    <input className="field" value={cfg.projectId} onChange={e => upd({ projectId: e.target.value })} placeholder="my-gcp-project-123" />
    <p className="muted" style={{ fontSize: 12, marginTop: 6 }}>Found in GCP Console → Project Settings</p>
    <div className="surface" style={{ padding: 16, marginTop: 20, background: 'var(--surface-2)' }}>
      <div className="eyebrow" style={{ marginBottom: 10, fontSize: 10 }}>Permissions we'll configure</div>
      {['roles/pubsub.subscriber on your DLQ subscription', 'roles/pubsub.publisher on your main topic', 'roles/logging.viewer on your project'].map((p, i) => (
        <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '5px 0', fontSize: 13, color: 'var(--text-2)' }}>
          <span style={{ color: 'var(--blue)', fontSize: 11 }}>→</span>
          <code className="mono" style={{ fontSize: 12, color: 'var(--blue)', background: 'var(--surface-3)', padding: '1px 6px', borderRadius: 4 }}>{p.split(' on ')[0]}</code>
          <span className="muted">{p.split(' on ')[1] ? 'on ' + p.split(' on ')[1] : ''}</span>
        </div>
      ))}
    </div>
  </div>
);

const StepPubSub = ({ cfg, upd }) => {
  const [resources, setResources] = React.useState(null);
  React.useEffect(() => {
    if (cfg.projectId) window.api.listPubsubResources({ projectId: cfg.projectId }).then(setResources);
    else setResources({ topics: [], subscriptions: [] });
  }, [cfg.projectId]);
  return (
    <div>
      <h2 className="h3" style={{ marginBottom: 6 }}>Select your Pub/Sub resources</h2>
      <p className="muted-2" style={{ fontSize: 14, marginBottom: 24 }}>Pick the DLQ subscription to monitor and the main topic to republish to.</p>
      <label className="lbl">DLQ Subscription</label>
      <select className="field" value={cfg.dlqSub} onChange={e => upd({ dlqSub: e.target.value })} style={{ marginBottom: 16 }}>
        <option value="">Select subscription…</option>
        {(resources?.subscriptions || []).filter(s => s.includes('dlq')).map(s => <option key={s} value={s}>{s}</option>)}
      </select>
      <label className="lbl">Main Topic (republish target)</label>
      <select className="field" value={cfg.mainTopic} onChange={e => upd({ mainTopic: e.target.value })}>
        <option value="">Select topic…</option>
        {(resources?.topics || []).filter(t => !t.includes('dlq')).map(t => <option key={t} value={t}>{t}</option>)}
      </select>
    </div>
  );
};

const StepContext = ({ cfg, upd }) => {
  const [uploading, setUploading] = React.useState(false);
  const addFile = async (e) => {
    const file = e.target.files[0];
    if (!file) return;
    setUploading(true);
    const res = await window.api.uploadContext(file);
    setUploading(false);
    upd({ files: [...cfg.files, { id: res.id, name: res.name, size: res.size }] });
  };
  return (
    <div>
      <h2 className="h3" style={{ marginBottom: 6 }}>Upload context for smarter repairs</h2>
      <p className="muted-2" style={{ fontSize: 14, marginBottom: 24 }}>Drop runbooks, post-mortems, source code, or paste a GitHub / web URL. We build a per-tenant GraphRAG index so the agent can cite real context.</p>
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12, marginBottom: 20 }}>
        <div>
          <label className="lbl">GitHub repository URL</label>
          <input className="field" value={cfg.githubUrl} onChange={e => upd({ githubUrl: e.target.value })} placeholder="https://github.com/acme/payments-api" />
        </div>
        <div>
          <label className="lbl">Web scrape URL</label>
          <input className="field" value={cfg.webUrl} onChange={e => upd({ webUrl: e.target.value })} placeholder="https://docs.acme.com/runbook" />
        </div>
      </div>
      <label className="lbl">Upload files</label>
      <div style={{ border: '1px dashed var(--line-strong)', borderRadius: 10, padding: 28, textAlign: 'center', background: 'var(--surface-1)', cursor: 'pointer', position: 'relative' }}>
        <input type="file" multiple onChange={addFile} style={{ position: 'absolute', inset: 0, opacity: 0, cursor: 'pointer' }} />
        <div style={{ fontSize: 13, color: 'var(--text-2)' }}>{uploading ? 'Uploading…' : 'Click or drag files here'}</div>
        <div className="muted" style={{ fontSize: 12, marginTop: 4 }}>PDF, Markdown, source archives, text</div>
      </div>
      {cfg.files.length > 0 && (
        <div style={{ marginTop: 12 }}>
          {cfg.files.map((f, i) => (
            <div key={f.id} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '8px 0', borderBottom: '1px solid var(--line)' }}>
              <span style={{ fontSize: 13 }}>{f.name}</span>
              <button className="btn btn-ghost btn-sm" onClick={() => upd({ files: cfg.files.filter((_, j) => j !== i) })}>Remove</button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

const FAILURE_CATEGORIES = ['Schema drift', 'Malformed JSON', 'Type mismatch', 'Missing field', 'Encoding', 'Downstream outage'];

const StepApproval = ({ cfg, upd }) => (
  <div>
    <h2 className="h3" style={{ marginBottom: 6 }}>Approval mode</h2>
    <p className="muted-2" style={{ fontSize: 14, marginBottom: 24 }}>Choose whether DeadLift auto-republishes fixes or waits for human approval. You can set a global default and override per failure category.</p>
    <div style={{ display: 'flex', gap: 12, marginBottom: 24 }}>
      {[{ key: 'human', label: 'Human-in-the-loop', desc: 'All fixes require manual approval before republish.' },
        { key: 'auto', label: 'Fully autonomous', desc: 'Auto-republish when confidence ≥ 90%. Queue the rest.' }].map(m => (
        <button key={m.key} className="surface" onClick={() => upd({ approvalMode: m.key })}
          style={{ flex: 1, padding: 16, cursor: 'pointer', textAlign: 'left', borderColor: cfg.approvalMode === m.key ? 'var(--text)' : 'var(--line)', background: cfg.approvalMode === m.key ? 'var(--surface-2)' : 'var(--surface-1)' }}>
          <div style={{ fontSize: 14, fontWeight: 500, marginBottom: 4 }}>{m.label}</div>
          <div className="muted-2" style={{ fontSize: 12.5 }}>{m.desc}</div>
        </button>
      ))}
    </div>
    <div className="eyebrow" style={{ marginBottom: 10 }}>Per-category overrides <span className="muted" style={{ textTransform: 'none', letterSpacing: 0 }}>(optional)</span></div>
    {FAILURE_CATEGORIES.map(cat => (
      <div key={cat} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '8px 0', borderBottom: '1px solid var(--line)' }}>
        <span style={{ fontSize: 13 }}>{cat}</span>
        <select className="field" style={{ width: 180, padding: '6px 10px', fontSize: 12.5 }}
          value={cfg.categoryOverrides[cat] || 'default'}
          onChange={e => upd({ categoryOverrides: { ...cfg.categoryOverrides, [cat]: e.target.value } })}>
          <option value="default">Use default</option>
          <option value="auto">Auto-approve</option>
          <option value="human">Require approval</option>
        </select>
      </div>
    ))}
  </div>
);

const StepBatching = ({ cfg, upd }) => (
  <div>
    <h2 className="h3" style={{ marginBottom: 6 }}>Batch threshold</h2>
    <p className="muted-2" style={{ fontSize: 14, marginBottom: 24 }}>When multiple messages fail for the same root cause, DeadLift groups them. Set the minimum count to create a batch (approve once, fix all).</p>
    <label className="lbl">Minimum messages to create a batch</label>
    <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
      <input type="range" min={2} max={100} step={1} value={cfg.batchThreshold}
        onChange={e => upd({ batchThreshold: Number(e.target.value) })}
        style={{ flex: 1, accentColor: 'var(--green)' }} />
      <span className="mono" style={{ fontSize: 22, fontWeight: 600, minWidth: 48, textAlign: 'right' }}>{cfg.batchThreshold}</span>
    </div>
    <p className="muted" style={{ fontSize: 12.5, marginTop: 10 }}>Lower = more batches (less manual work). Higher = fewer batches (more granular control).</p>
    <div className="surface" style={{ padding: 16, marginTop: 24, background: 'var(--surface-2)' }}>
      <div style={{ fontSize: 13, color: 'var(--text-2)' }}>Example: If a downstream node goes down and 200 messages fail, a batch threshold of <strong style={{ color: 'var(--text)' }}>{cfg.batchThreshold}</strong> means you'll see <strong style={{ color: 'var(--green)' }}>1 batch card</strong> instead of 200 individual fixes.</div>
    </div>
  </div>
);

const StepNotify = ({ cfg, upd }) => (
  <div>
    <h2 className="h3" style={{ marginBottom: 6 }}>Notification settings</h2>
    <p className="muted-2" style={{ fontSize: 14, marginBottom: 24 }}>Get alerted when fixes need approval or when a new root cause is detected.</p>
    <div style={{ display: 'grid', gap: 16 }}>
      <div>
        <label className="lbl">Slack webhook URL</label>
        <input className="field" value={cfg.notifications.slack} onChange={e => upd({ notifications: { ...cfg.notifications, slack: e.target.value } })} placeholder="https://hooks.slack.com/services/..." />
      </div>
      <div>
        <label className="lbl">Email</label>
        <input className="field" value={cfg.notifications.email} onChange={e => upd({ notifications: { ...cfg.notifications, email: e.target.value } })} />
      </div>
      <div>
        <label className="lbl">PagerDuty integration key <span className="muted" style={{ fontWeight: 400 }}>(optional)</span></label>
        <input className="field" value={cfg.notifications.pagerduty} onChange={e => upd({ notifications: { ...cfg.notifications, pagerduty: e.target.value } })} placeholder="e.g. R03F..." />
      </div>
    </div>
  </div>
);

const StepReview = ({ cfg, onFinish }) => {
  const [loading, setLoading] = React.useState(false);
  const go = async () => { setLoading(true); await onFinish(); };
  return (
    <div>
      <h2 className="h3" style={{ marginBottom: 6 }}>Review & finish</h2>
      <p className="muted-2" style={{ fontSize: 14, marginBottom: 24 }}>Confirm your settings. You can change everything later from the dashboard.</p>
      <div className="surface" style={{ padding: 18, marginBottom: 16 }}>
        <ReviewRow label="GCP Project" value={cfg.projectId || '—'} />
        <ReviewRow label="DLQ Subscription" value={cfg.dlqSub || '—'} />
        <ReviewRow label="Main Topic" value={cfg.mainTopic || '—'} />
        <ReviewRow label="Context files" value={cfg.files.length > 0 ? cfg.files.map(f => f.name).join(', ') : 'None uploaded'} />
        <ReviewRow label="GitHub URL" value={cfg.githubUrl || '—'} />
        <ReviewRow label="Approval mode" value={cfg.approvalMode === 'auto' ? 'Fully autonomous' : 'Human-in-the-loop'} />
        <ReviewRow label="Batch threshold" value={`≥ ${cfg.batchThreshold} messages`} />
        <ReviewRow label="Notifications" value={[cfg.notifications.slack && 'Slack', cfg.notifications.email && 'Email', cfg.notifications.pagerduty && 'PagerDuty'].filter(Boolean).join(', ') || '—'} last />
      </div>
      <button className="btn btn-green btn-lg" style={{ width: '100%' }} onClick={go} disabled={loading}>
        {loading ? 'Finishing setup…' : 'Launch DeadLift'}
      </button>
    </div>
  );
};

const ReviewRow = ({ label, value, last }) => (
  <div style={{ display: 'flex', justifyContent: 'space-between', padding: '9px 0', borderBottom: last ? 'none' : '1px solid var(--line)', gap: 16 }}>
    <span className="muted-2" style={{ fontSize: 13, flexShrink: 0 }}>{label}</span>
    <span className="mono" style={{ fontSize: 13, color: 'var(--text)', textAlign: 'right', wordBreak: 'break-all' }}>{value}</span>
  </div>
);

Object.assign(window, { Onboarding });
