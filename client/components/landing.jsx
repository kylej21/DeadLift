// ============================================================
// Landing page — DeadLift
// ============================================================

const Landing = () => {
  return (
    <div>
      <TopNav variant="public" />
      <Hero />
      <StatsStrip />
      <GcpQuoteSection />
      <HowItWorks />
      <LiveDemoSection />
      <CtaSection />
      <Footer />
    </div>
  );
};

const Hero = () => (
  <section style={{ position: 'relative', overflow: 'hidden' }}>
    <div className="hero-bg" />
    <div className="container" style={{ position: 'relative', zIndex: 1, padding: '90px 24px 80px' }}>
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 8, padding: '5px 12px 5px 6px', borderRadius: 999, border: '1px solid var(--line-strong)', background: 'rgba(255,255,255,0.03)', marginBottom: 28 }}>
        <span style={{ background: 'var(--green-bg)', color: 'var(--green)', fontSize: 10, fontWeight: 600, padding: '2px 8px', borderRadius: 999, border: '1px solid var(--green-line)', letterSpacing: '0.06em' }}>NEW</span>
        <span style={{ fontSize: 12.5, color: 'var(--text-2)' }}>Now connecting GCP Pub/Sub projects</span>
        <span className="muted" style={{ fontSize: 12.5 }}>→</span>
      </div>

      <h1 className="display" style={{ maxWidth: 1000 }}>
        The on-call agent for<br />
        <span style={{ background: 'linear-gradient(180deg, oklch(0.95 0.10 145), oklch(0.70 0.18 145))', WebkitBackgroundClip: 'text', backgroundClip: 'text', color: 'transparent' }}>event-driven workflows.</span>
      </h1>

      <p style={{ fontSize: 18, color: 'var(--text-2)', maxWidth: 640, marginTop: 28, lineHeight: 1.55 }}>
        DeadLift sits on your Pub/Sub dead-letter queue, finds the root cause, drafts the fix, and republishes — automatically or with one tap. GCP forwards. We repair.
      </p>

      <div style={{ display: 'flex', gap: 12, marginTop: 36, flexWrap: 'wrap' }}>
        <a href="#/signup" className="btn btn-primary btn-lg">Connect your GCP project <ArrowIcon /></a>
        <a href="#demo" className="btn btn-lg">See it run</a>
      </div>

      <div style={{ display: 'flex', alignItems: 'center', gap: 24, marginTop: 36, color: 'var(--text-3)', fontSize: 12.5, flexWrap: 'wrap' }}>
        <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}><CheckIcon size={12} color="var(--green)" /> Read-only IAM by default</span>
        <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}><CheckIcon size={12} color="var(--green)" /> Hooks the DLQ — never the hot path</span>
        <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}><CheckIcon size={12} color="var(--green)" /> Human-in-the-loop or fully autonomous</span>
      </div>

      {/* Hero visual: live DLQ animation preview */}
      <div style={{ marginTop: 64 }}>
        <HeroVisual />
      </div>
    </div>
  </section>
);

// Animated hero visual: messages flowing → DLQ → DeadLift → republish
const HeroVisual = () => {
  const [tick, setTick] = React.useState(0);
  React.useEffect(() => {
    const id = setInterval(() => setTick(t => t + 1), 2200);
    return () => clearInterval(id);
  }, []);
  return (
    <div className="surface" style={{ padding: 24, position: 'relative', overflow: 'hidden' }}>
      <div style={{ display: 'flex', alignItems: 'stretch', justifyContent: 'space-between', gap: 16, minHeight: 200 }}>
        <PipelineNode label="Producer" sub="payments-api" status="ok" />
        <PipelineFlow tick={tick} fromOk />
        <PipelineNode label="Topic" sub="payments-events" status="ok" />
        <PipelineFlow tick={tick} variant="branch" />
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12, flex: '0 0 auto', justifyContent: 'space-between' }}>
          <PipelineNode label="Consumer" sub="ledger-worker" status="failed" small />
          <PipelineNode label="DLQ" sub="payments-events-dlq" status="dlq" small />
        </div>
        <PipelineFlow tick={tick} variant="repair" />
        <PipelineNode label="DeadLift" sub="agent" status="agent" highlight />
        <PipelineFlow tick={tick} variant="republish" />
        <PipelineNode label="Republished" sub="payments-events" status="fixed" />
      </div>
      <div style={{ marginTop: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: 12, color: 'var(--text-3)' }}>
        <span>Real-time DLQ ⇒ repair pipeline · synthetic data</span>
        <span style={{ display: 'inline-flex', alignItems: 'center', gap: 8 }}>
          <span className="dot pulse" style={{ background: 'var(--green)' }} /> Replaying every 2.2s
        </span>
      </div>
    </div>
  );
};

const PipelineNode = ({ label, sub, status, highlight, small }) => {
  const map = {
    ok: { fg: 'var(--text-2)', bg: 'var(--surface-2)', dot: 'var(--text-3)' },
    failed: { fg: 'var(--red)', bg: 'var(--red-bg)', dot: 'var(--red)' },
    dlq: { fg: 'var(--amber)', bg: 'var(--amber-bg)', dot: 'var(--amber)' },
    agent: { fg: 'var(--green)', bg: 'var(--green-bg)', dot: 'var(--green)' },
    fixed: { fg: 'var(--green)', bg: 'var(--green-bg)', dot: 'var(--green)' },
  }[status] || {};
  return (
    <div style={{
      flex: '1 1 0', minWidth: 110,
      background: 'var(--surface-1)',
      border: highlight ? '1px solid var(--green-line)' : '1px solid var(--line)',
      borderRadius: 10, padding: small ? '10px 12px' : '14px 14px',
      boxShadow: highlight ? '0 0 0 4px oklch(0.78 0.17 145 / 0.08)' : 'none',
      display: 'flex', flexDirection: 'column', gap: 6, justifyContent: 'center',
    }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
        <span className="dot" style={{ background: map.dot }} />
        <span style={{ fontSize: 11.5, color: 'var(--text-3)', textTransform: 'uppercase', letterSpacing: '0.08em', fontWeight: 500 }}>{label}</span>
      </div>
      <div className="mono" style={{ fontSize: 12.5, color: map.fg || 'var(--text)' }}>{sub}</div>
    </div>
  );
};

const PipelineFlow = ({ tick, variant }) => {
  const phase = tick % 2;
  return (
    <div style={{ flex: '0 0 56px', display: 'flex', alignItems: 'center', justifyContent: 'center', position: 'relative' }}>
      <svg width="56" height="100%" viewBox="0 0 56 200" preserveAspectRatio="none" style={{ overflow: 'visible' }}>
        {variant === 'branch' ? (
          <>
            <line x1="0" y1="100" x2="56" y2="60" stroke="var(--red-line)" strokeWidth="1" strokeDasharray="2 3" />
            <line x1="0" y1="100" x2="56" y2="140" stroke="var(--amber-line)" strokeWidth="1" strokeDasharray="2 3" />
          </>
        ) : variant === 'repair' ? (
          <line x1="0" y1="140" x2="56" y2="100" stroke="var(--amber-line)" strokeWidth="1" strokeDasharray="2 3" />
        ) : variant === 'republish' ? (
          <line x1="0" y1="100" x2="56" y2="100" stroke="var(--green-line)" strokeWidth="1.5" />
        ) : (
          <line x1="0" y1="100" x2="56" y2="100" stroke="var(--line-strong)" strokeWidth="1" strokeDasharray="2 3" />
        )}
        <circle r="3" fill={
          variant === 'repair' ? 'var(--amber)' :
          variant === 'republish' ? 'var(--green)' :
          variant === 'branch' ? (phase ? 'var(--amber)' : 'var(--text-2)') :
          'var(--text-2)'
        }>
          <animate attributeName="cx" values="0;56" dur="1.6s" repeatCount="indefinite" />
          <animate attributeName="cy" values={
            variant === 'branch' ? (phase ? '100;140' : '100;60') :
            variant === 'repair' ? '140;100' :
            '100;100'
          } dur="1.6s" repeatCount="indefinite" />
        </circle>
      </svg>
    </div>
  );
};

// ===== Stats strip =====
const StatsStrip = () => (
  <section id="stats" className="container" style={{ padding: '80px 24px 40px' }}>
    <div className="eyebrow" style={{ marginBottom: 12 }}>The cost of not having this</div>
    <h2 className="h2" style={{ maxWidth: 800, marginBottom: 48 }}>
      The bill for manual DLQ triage is hiding in plain sight.
    </h2>
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 0, border: '1px solid var(--line)', borderRadius: 14, overflow: 'hidden' }}>
      <Stat big="$300K" unit="/min" label="Average enterprise loss during a production incident" />
      <Stat big="60–80%" label="of incident time spent on investigation, not the fix" border />
      <Stat big="$40K" unit="weekend" label="A single misconfigured retry policy can burn this in cloud" border />
      <Stat big="$3.45M" unit="/yr" label="Estimated savings per enterprise from automating fixable failures" border />
    </div>
  </section>
);

const Stat = ({ big, unit, label, border }) => (
  <div style={{
    padding: '28px 24px',
    borderLeft: border ? '1px solid var(--line)' : 'none',
    background: 'var(--surface-1)',
  }}>
    <div style={{ display: 'flex', alignItems: 'baseline', gap: 6, marginBottom: 10 }}>
      <span style={{ fontSize: 44, fontWeight: 600, letterSpacing: '-0.03em', lineHeight: 1 }}>{big}</span>
      {unit && <span className="muted" style={{ fontSize: 14 }}>{unit}</span>}
    </div>
    <div style={{ fontSize: 13.5, color: 'var(--text-2)', lineHeight: 1.5 }}>{label}</div>
  </div>
);

// ===== GCP doc quote =====
const GcpQuoteSection = () => (
  <section className="container" style={{ padding: '60px 24px' }}>
    <div className="surface" style={{ padding: '40px 40px 36px', position: 'relative', overflow: 'hidden' }}>
      <div className="eyebrow" style={{ marginBottom: 16, color: 'var(--text-3)' }}>From cloud.google.com/pubsub/docs · dead-letter topics</div>
      <blockquote style={{ margin: 0, fontSize: 22, lineHeight: 1.45, fontWeight: 400, letterSpacing: '-0.012em', color: 'var(--text)', maxWidth: 980 }}>
        <span className="muted" style={{ fontFamily: 'Georgia, serif', fontSize: 56, lineHeight: 0.4, marginRight: 4, verticalAlign: '-0.3em' }}>“</span>
        When Pub/Sub forwards an undeliverable message, it wraps the original message in a new one and adds attributes that identify the source subscription. The message is then sent to the specified dead-letter topic. <span style={{ color: 'var(--text-3)' }}>A separate subscription attached to the dead-letter topic can then receive these forwarded messages for analysis and offline debugging.</span>
      </blockquote>
      <div style={{ marginTop: 32, display: 'flex', alignItems: 'center', gap: 16, flexWrap: 'wrap' }}>
        <div style={{ height: 1, flex: 1, background: 'var(--line)' }} />
        <span className="pill pill-green" style={{ fontSize: 12, padding: '5px 12px' }}>DeadLift fills the gap</span>
        <div style={{ height: 1, flex: 1, background: 'var(--line)' }} />
      </div>
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16, marginTop: 24 }}>
        <div className="surface-2 surface" style={{ padding: 18, background: 'var(--surface-2)' }}>
          <div className="eyebrow" style={{ color: 'var(--red)', marginBottom: 8 }}>What GCP gives you</div>
          <div style={{ fontSize: 14, color: 'var(--text-2)', lineHeight: 1.55 }}>A second queue. A pile of dead letters. The expectation that an engineer will eventually pull, parse, and republish them by hand.</div>
        </div>
        <div className="surface-2 surface" style={{ padding: 18, background: 'var(--surface-2)' }}>
          <div className="eyebrow" style={{ color: 'var(--green)', marginBottom: 8 }}>What DeadLift gives you</div>
          <div style={{ fontSize: 14, color: 'var(--text-2)', lineHeight: 1.55 }}>An agent that diagnoses the root cause, drafts the fix, batches duplicates, and republishes — with a paper trail and human approval where you want it.</div>
        </div>
      </div>
    </div>
  </section>
);

// ===== How it works =====
const HowItWorks = () => (
  <section id="features" className="container" style={{ padding: '80px 24px' }}>
    <div className="eyebrow" style={{ marginBottom: 12 }}>How it works</div>
    <h2 className="h2" style={{ marginBottom: 12, maxWidth: 800 }}>An on-call triage agent — without the on-call.</h2>
    <p style={{ fontSize: 16, color: 'var(--text-2)', maxWidth: 680, marginBottom: 48 }}>
      DeadLift hooks only the dead-letter queue, so it never touches the hot path. Every other request stays at full Pub/Sub speed.
    </p>
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(260px, 1fr))', gap: 16 }}>
      <Feature
        n="01" title="Connect"
        body="OAuth into GCP. We grant a least-privilege service account: subscribe on your DLQ, publish on your main topic, log viewer."
      />
      <Feature
        n="02" title="Index your context"
        body="Drop in runbooks, post-mortems, source archives, GitHub URLs. We build a per-tenant GraphRAG index for smart retrieval."
      />
      <Feature
        n="03" title="Diagnose"
        body="Each DLQ message is correlated with logs, recent deploys, and source. The agent picks a hypothesis and explains it."
      />
      <Feature
        n="04" title="Fix or batch"
        body="If 47 messages share a root cause, you approve once. Auto-republish or human-in-the-loop, configurable per category."
      />
    </div>

    {/* Why this can't run on the hot path */}
    <div className="surface" style={{ padding: 28, marginTop: 56, background: 'linear-gradient(180deg, var(--surface-1), var(--bg-1))' }}>
      <div style={{ display: 'flex', alignItems: 'flex-start', gap: 24, flexWrap: 'wrap' }}>
        <div style={{ flex: '1 1 360px' }}>
          <div className="eyebrow" style={{ marginBottom: 10 }}>Why we sit on the DLQ</div>
          <h3 className="h3" style={{ marginBottom: 10 }}>Logging every message would explode in cost. Repairing every message would slow them all down.</h3>
          <p className="muted-2" style={{ fontSize: 14, lineHeight: 1.6, marginBottom: 12 }}>
            At 20K msgs/sec — Roblox’s Pub/Sub scale — there’s no room for an LLM in the request path, and persistent per-message logging is a six-figure cloud bill on its own. The DLQ is the natural sampling boundary: a message only arrives there after retries are exhausted, and that’s exactly when the cost of an LLM-shaped repair is finally worth it.
          </p>
        </div>
        <div style={{ flex: '0 0 320px', display: 'grid', gap: 10 }}>
          <CostRow label="Log every msg @ 20k/s" value="$$$$" pct={1.0} bad />
          <CostRow label="LLM-fix every msg" value="∞ latency" pct={1.0} bad />
          <CostRow label="DeadLift (DLQ-only)" value="< 1% volume" pct={0.05} good />
        </div>
      </div>
    </div>
  </section>
);

const CostRow = ({ label, value, pct, good, bad }) => (
  <div>
    <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12.5, color: 'var(--text-2)', marginBottom: 6 }}>
      <span>{label}</span>
      <span className="mono" style={{ color: good ? 'var(--green)' : bad ? 'var(--red)' : 'var(--text-2)' }}>{value}</span>
    </div>
    <div style={{ height: 6, background: 'var(--surface-3)', borderRadius: 999, overflow: 'hidden' }}>
      <div style={{ height: '100%', width: `${pct*100}%`, background: good ? 'var(--green)' : bad ? 'var(--red)' : 'var(--text-2)', borderRadius: 999 }} />
    </div>
  </div>
);

const Feature = ({ n, title, body }) => (
  <div className="surface" style={{ padding: 22 }}>
    <div className="mono" style={{ fontSize: 11, color: 'var(--text-3)', marginBottom: 16 }}>{n}</div>
    <h3 style={{ fontSize: 17, marginBottom: 8 }}>{title}</h3>
    <p style={{ fontSize: 13.5, color: 'var(--text-2)', lineHeight: 1.55, margin: 0 }}>{body}</p>
  </div>
);

// ===== Live demo =====
const LiveDemoSection = () => {
  const [step, setStep] = React.useState(0);
  React.useEffect(() => {
    const id = setInterval(() => setStep(s => (s + 1) % 4), 2400);
    return () => clearInterval(id);
  }, []);
  const steps = [
    { label: 'DLQ message received', detail: 'TypeError on amount_cents · payments-events-dlq', tone: 'amber' },
    { label: 'Correlated with deploy', detail: 'payments-api v3.1.0 → v3.2.1 · Cloud Build #8821 · 14:18 UTC', tone: 'blue' },
    { label: 'Repair drafted', detail: 'Rename amount → amount_cents · 47 similar messages batched', tone: 'green' },
    { label: 'Republished', detail: 'All 47 → payments-events · ack’d · 1.4s end-to-end', tone: 'green' },
  ];
  return (
    <section id="demo" className="container" style={{ padding: '80px 24px' }}>
      <div className="eyebrow" style={{ marginBottom: 12 }}>Live demo</div>
      <h2 className="h2" style={{ marginBottom: 16, maxWidth: 800 }}>What an incident looks like with DeadLift on duty.</h2>
      <p className="muted-2" style={{ fontSize: 16, marginBottom: 40, maxWidth: 600 }}>From DLQ ingest to republished payload, on a real-shaped payments event.</p>
      <div style={{ display: 'grid', gridTemplateColumns: '320px 1fr', gap: 16 }}>
        <div className="surface" style={{ padding: 16 }}>
          <div className="eyebrow" style={{ marginBottom: 12 }}>Timeline</div>
          {steps.map((s, i) => (
            <div key={i} style={{
              display: 'flex', gap: 12, padding: '10px 4px', alignItems: 'flex-start',
              opacity: i <= step ? 1 : 0.35, transition: 'opacity 300ms ease',
              borderTop: i === 0 ? 'none' : '1px solid var(--line)',
            }}>
              <span className={`dot ${i === step ? 'pulse' : ''}`} style={{
                marginTop: 6, background: i <= step ? `var(--${s.tone})` : 'var(--text-4)',
              }} />
              <div>
                <div style={{ fontSize: 13, fontWeight: 500 }}>{s.label}</div>
                <div className="muted" style={{ fontSize: 12, marginTop: 2 }}>{s.detail}</div>
              </div>
            </div>
          ))}
        </div>
        <DemoFixCard reveal={step >= 2} fully={step >= 3} />
      </div>
    </section>
  );
};

const DemoFixCard = ({ reveal, fully }) => (
  <div className="surface" style={{ padding: 0, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
    <div style={{ padding: '14px 18px', borderBottom: '1px solid var(--line)', display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 12, flexWrap: 'wrap' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
        <span className={fully ? 'pill pill-green' : reveal ? 'pill pill-amber' : 'pill'}>
          <span className="dot" style={{ background: fully ? 'var(--green)' : reveal ? 'var(--amber)' : 'var(--text-3)' }} />
          {fully ? 'Republished' : reveal ? 'Awaiting approval' : 'Diagnosing'}
        </span>
        <span className="pill">Schema drift</span>
        <span className="muted mono" style={{ fontSize: 12 }}>fix_8a3c91</span>
      </div>
      <span className="muted mono" style={{ fontSize: 12 }}>payments-events-dlq-sub</span>
    </div>
    <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', borderBottom: '1px solid var(--line)' }}>
      <DiffPane title="Before" lines={[
        { type: 'normal', text: '{' },
        { type: 'normal', text: '  "event": "charge.completed",' },
        { type: 'normal', text: '  "id": "ch_2nL9KQ",' },
        { type: 'del', text: '  "amount": 4999,' },
        { type: 'normal', text: '  "currency": "usd"' },
        { type: 'normal', text: '}' },
      ]} />
      <DiffPane title="After" lines={[
        { type: 'normal', text: '{' },
        { type: 'normal', text: '  "event": "charge.completed",' },
        { type: 'normal', text: '  "id": "ch_2nL9KQ",' },
        { type: 'add', text: '  "amount_cents": 4999,' },
        { type: 'normal', text: '  "currency": "usd"' },
        { type: 'normal', text: '}' },
      ]} dim={!reveal} />
    </div>
    <div style={{ padding: '14px 18px', display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 12, flexWrap: 'wrap' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, flex: '1 1 auto', minWidth: 0 }}>
        <ConfidenceRing value={0.94} />
        <div style={{ minWidth: 0 }}>
          <div style={{ fontSize: 13, fontWeight: 500 }}>Producer renamed <span className="mono">amount</span> → <span className="mono">amount_cents</span></div>
          <div className="muted" style={{ fontSize: 12.5 }}>This fix applies to <strong style={{ color: 'var(--text)' }}>47 similar messages</strong></div>
        </div>
      </div>
      <div style={{ display: 'flex', gap: 8 }}>
        <button className="btn btn-sm">Edit</button>
        <button className={fully ? 'btn btn-sm btn-green' : 'btn btn-sm btn-green'} disabled>{fully ? '✓ Approved' : 'Approve all 47'}</button>
      </div>
    </div>
  </div>
);

const DiffPane = ({ title, lines, dim }) => (
  <div style={{ borderRight: title === 'Before' ? '1px solid var(--line)' : 'none', opacity: dim ? 0.4 : 1, transition: 'opacity 300ms ease' }}>
    <div style={{ padding: '10px 14px', borderBottom: '1px solid var(--line)', fontSize: 11.5, color: 'var(--text-3)', textTransform: 'uppercase', letterSpacing: '0.08em' }}>{title}</div>
    <div style={{ padding: '8px 0', background: 'var(--surface-1)' }}>
      {lines.map((l, i) => (
        <div key={i} className={`diff-line ${l.type === 'add' ? 'diff-add' : l.type === 'del' ? 'diff-del' : ''}`}>
          <span className="diff-gutter">{i + 1}</span>
          <span className="diff-marker">{l.type === 'add' ? '+' : l.type === 'del' ? '−' : ' '}</span>
          <span className="diff-content">{l.text}</span>
        </div>
      ))}
    </div>
  </div>
);

// ===== CTA =====
const CtaSection = () => (
  <section className="container" style={{ padding: '60px 24px 100px' }}>
    <div className="surface" style={{
      padding: '56px 40px',
      background: 'radial-gradient(ellipse 60% 80% at 100% 0%, oklch(0.55 0.18 145 / 0.18), transparent 70%), var(--surface-1)',
      textAlign: 'center',
    }}>
      <h2 className="h2" style={{ marginBottom: 14 }}>Stop paying engineers to copy-paste from a DLQ.</h2>
      <p style={{ fontSize: 16, color: 'var(--text-2)', maxWidth: 560, margin: '0 auto 28px' }}>
        Connect your GCP project in under five minutes. Read-only by default. Cancel any time.
      </p>
      <div style={{ display: 'flex', justifyContent: 'center', gap: 12, flexWrap: 'wrap' }}>
        <a href="#/signup" className="btn btn-primary btn-lg">Get started <ArrowIcon /></a>
        <a href="#/signin" className="btn btn-lg">I already have an account</a>
      </div>
    </div>
  </section>
);

const Footer = () => (
  <footer style={{ borderTop: '1px solid var(--line)', padding: '32px 24px', marginTop: 24 }}>
    <div className="container" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 16, flexWrap: 'wrap' }}>
      <Logo />
      <div className="muted" style={{ fontSize: 12.5 }}>© 2026 DeadLift · Built at LA Hacks</div>
      <div style={{ display: 'flex', gap: 18, fontSize: 12.5, color: 'var(--text-3)' }}>
        <a href="#features">How it works</a>
        <a href="#stats">The gap</a>
        <a href="#demo">Demo</a>
      </div>
    </div>
  </footer>
);

Object.assign(window, { Landing });
