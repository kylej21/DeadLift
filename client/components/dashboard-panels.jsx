// ============================================================
// Analytics tab
// ============================================================

const AnalyticsTab = ({ data }) => {
  if (!data) return <LoadingState />;
  const k = data.kpis;
  return (
    <div>
      {/* KPI strip */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(190px, 1fr))', gap: 12, marginBottom: 28 }}>
        <KPI label="DLQ volume (24h)" value={k.dlqVolume24h.toLocaleString()} />
        <KPI label="Auto-fixed" value={k.autoFixed.toLocaleString()} accent="green" />
        <KPI label="Awaiting approval" value={k.awaitingApproval.toString()} accent="amber" />
        <KPI label="Unfixable" value={k.unfixable.toString()} accent="red" />
        <KPI label="MTTR before" value={k.mttrBefore} />
        <KPI label="MTTR with DeadLift" value={k.mttrAfter} delta={`${k.mttrDelta}%`} accent="green" />
        <KPI label="Est. savings (30d)" value={`$${(k.estSavings30d / 1000).toFixed(0)}K`} accent="green" />
      </div>

      {/* Volume chart */}
      <div className="surface" style={{ padding: 20, marginBottom: 20 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
          <div>
            <div className="eyebrow" style={{ marginBottom: 4 }}>DLQ volume · 24 hours</div>
            <div className="muted-2" style={{ fontSize: 12.5 }}>Messages entering DLQ vs. repaired</div>
          </div>
          <div style={{ display: 'flex', gap: 14, fontSize: 11.5 }}>
            <span style={{ display: 'inline-flex', alignItems: 'center', gap: 5 }}><span className="dot" style={{ background: 'var(--text-3)' }} /> DLQ in</span>
            <span style={{ display: 'inline-flex', alignItems: 'center', gap: 5 }}><span className="dot" style={{ background: 'var(--green)' }} /> Fixed</span>
            <span style={{ display: 'inline-flex', alignItems: 'center', gap: 5 }}><span className="dot" style={{ background: 'var(--red)' }} /> Unfixable</span>
          </div>
        </div>
        <MiniBarChart series={data.series} />
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
        {/* Category breakdown */}
        <div className="surface" style={{ padding: 20 }}>
          <div className="eyebrow" style={{ marginBottom: 16 }}>Failure categories</div>
          {data.categories.map(c => (
            <div key={c.name} style={{ marginBottom: 12 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12.5, marginBottom: 5 }}>
                <span style={{ color: 'var(--text-2)' }}>{c.name}</span>
                <span className="mono" style={{ color: 'var(--text)' }}>{c.count} <span className="muted">({c.pct}%)</span></span>
              </div>
              <div style={{ height: 5, background: 'var(--surface-3)', borderRadius: 999, overflow: 'hidden' }}>
                <div style={{ height: '100%', width: `${c.pct}%`, background: c.color, borderRadius: 999, transition: 'width 500ms ease' }} />
              </div>
            </div>
          ))}
        </div>

        {/* Per-topic table */}
        <div className="surface" style={{ padding: 20 }}>
          <div className="eyebrow" style={{ marginBottom: 16 }}>Per-topic breakdown</div>
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 13 }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--line)' }}>
                <th style={{ textAlign: 'left', padding: '6px 0', color: 'var(--text-3)', fontWeight: 500, fontSize: 11.5 }}>Topic</th>
                <th style={{ textAlign: 'right', padding: '6px 8px', color: 'var(--text-3)', fontWeight: 500, fontSize: 11.5 }}>DLQ</th>
                <th style={{ textAlign: 'right', padding: '6px 8px', color: 'var(--text-3)', fontWeight: 500, fontSize: 11.5 }}>Fixed</th>
                <th style={{ textAlign: 'right', padding: '6px 0', color: 'var(--text-3)', fontWeight: 500, fontSize: 11.5 }}>MTTR</th>
              </tr>
            </thead>
            <tbody>
              {data.topics.map(t => (
                <tr key={t.name} style={{ borderBottom: '1px solid var(--line)' }}>
                  <td className="mono" style={{ padding: '8px 0', fontSize: 12.5 }}>{t.name}</td>
                  <td style={{ textAlign: 'right', padding: '8px 8px' }}>{t.dlq}</td>
                  <td style={{ textAlign: 'right', padding: '8px 8px', color: 'var(--green)' }}>{t.fixed}</td>
                  <td className="mono" style={{ textAlign: 'right', padding: '8px 0', fontSize: 12 }}>{t.mttr}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

const KPI = ({ label, value, delta, accent }) => {
  const color = accent === 'green' ? 'var(--green)' : accent === 'amber' ? 'var(--amber)' : accent === 'red' ? 'var(--red)' : 'var(--text)';
  return (
    <div className="surface" style={{ padding: '14px 16px' }}>
      <div className="eyebrow" style={{ marginBottom: 8, fontSize: 10 }}>{label}</div>
      <div style={{ display: 'flex', alignItems: 'baseline', gap: 6 }}>
        <span className="mono" style={{ fontSize: 26, fontWeight: 700, color, letterSpacing: '-0.02em' }}>{value}</span>
        {delta && <span className="mono" style={{ fontSize: 12, color }}>{delta}</span>}
      </div>
    </div>
  );
};

const MiniBarChart = ({ series }) => {
  const max = Math.max(...series.map(s => s.dlq));
  const h = 140;
  return (
    <div style={{ display: 'flex', alignItems: 'flex-end', gap: 3, height: h }}>
      {series.map((s, i) => {
        const total = s.dlq;
        const fixedH = (s.fixed / max) * h;
        const unfixH = (s.unfixable / max) * h;
        const awaitH = (s.awaiting / max) * h;
        return (
          <div key={i} style={{ flex: 1, display: 'flex', flexDirection: 'column', justifyContent: 'flex-end', height: h, position: 'relative' }} title={`${s.hour}:00 — DLQ: ${s.dlq}, Fixed: ${s.fixed}`}>
            <div style={{ height: unfixH, background: 'var(--red)', opacity: 0.5, borderRadius: '2px 2px 0 0' }} />
            <div style={{ height: awaitH, background: 'var(--amber)', opacity: 0.6 }} />
            <div style={{ height: fixedH, background: 'var(--green)', opacity: 0.7, borderRadius: '0 0 2px 2px' }} />
          </div>
        );
      })}
    </div>
  );
};

// ============================================================
// RCA tab
// ============================================================

const RCATab = ({ reports }) => {
  if (!reports) return <LoadingState />;
  if (reports.length === 0) return <EmptyState title="No root cause reports" desc="Reports are generated when the agent completes an investigation." />;
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      {reports.map(r => <RCACard key={r.id} report={r} />)}
    </div>
  );
};

const RCACard = ({ report }) => {
  const [expanded, setExpanded] = React.useState(true);
  return (
    <div className="surface" style={{ overflow: 'hidden' }}>
      <div onClick={() => setExpanded(!expanded)} style={{ padding: '16px 18px', display: 'flex', justifyContent: 'space-between', alignItems: 'center', cursor: 'pointer', gap: 12, flexWrap: 'wrap' }}>
        <div>
          <div style={{ fontSize: 16, fontWeight: 500, marginBottom: 2 }}>{report.title}</div>
          <div className="muted" style={{ fontSize: 12.5 }}>Run at {report.runAt} · {report.affected} messages affected</div>
        </div>
        <ConfidenceRing value={report.confidence} size={36} />
      </div>
      {expanded && (
        <div style={{ padding: '0 18px 18px' }}>
          {/* Summary */}
          <div style={{ background: 'var(--surface-2)', borderRadius: 10, padding: 16, marginBottom: 16, border: '1px solid var(--line)' }}>
            <div className="eyebrow" style={{ marginBottom: 6, fontSize: 10 }}>Summary</div>
            <p style={{ fontSize: 13.5, color: 'var(--text-2)', margin: 0, lineHeight: 1.55 }}>{report.summary}</p>
          </div>

          {/* Hypothesis tree */}
          <div className="eyebrow" style={{ marginBottom: 10 }}>Hypotheses</div>
          {report.hypotheses.map(h => (
            <div key={h.id} style={{ marginBottom: 12, padding: 14, borderRadius: 10, border: h.winner ? '1px solid var(--green-line)' : '1px solid var(--line)', background: h.winner ? 'var(--green-bg)' : 'var(--surface-1)' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  {h.winner && <span className="pill pill-green" style={{ fontSize: 10, padding: '2px 8px' }}>Most likely</span>}
                  <span style={{ fontSize: 13.5, fontWeight: 500 }}>{h.text}</span>
                </div>
                <span className="mono" style={{ fontSize: 12, color: h.confidence > 0.7 ? 'var(--green)' : 'var(--text-3)' }}>{Math.round(h.confidence * 100)}%</span>
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
                {h.evidence.map((ev, i) => (
                  <div key={i} style={{ display: 'flex', alignItems: 'flex-start', gap: 8, fontSize: 12.5, color: 'var(--text-2)' }}>
                    <SourceIcon kind={ev.kind} />
                    <div>
                      <span style={{ fontWeight: 500, color: 'var(--text)' }}>{ev.label}</span>
                      {ev.t && <span className="muted mono" style={{ marginLeft: 6, fontSize: 11 }}>{ev.t}</span>}
                      <div className="muted" style={{ marginTop: 1 }}>{ev.detail}</div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ))}

          {/* Recommendation */}
          <div style={{ background: 'var(--blue-bg)', border: '1px solid oklch(0.72 0.14 240 / 0.3)', borderRadius: 10, padding: 14, marginTop: 8 }}>
            <div className="eyebrow" style={{ fontSize: 10, color: 'var(--blue)', marginBottom: 4 }}>Recommendation</div>
            <div style={{ fontSize: 13.5, color: 'var(--text)' }}>{report.recommendation}</div>
          </div>
        </div>
      )}
    </div>
  );
};

// ============================================================
// Settings tab
// ============================================================

const SettingsTab = () => {
  const [showConfirm, setShowConfirm] = React.useState(false);
  const [terminating, setTerminating] = React.useState(false);

  const handleTerminate = async () => {
    setTerminating(true);
    await window.api.terminateService();
    setTerminating(false);
    setShowConfirm(false);
    window.session.setUser(null);
    location.hash = '#/';
  };

  return (
    <div style={{ maxWidth: 680 }}>
      <h3 className="h3" style={{ marginBottom: 6 }}>Settings</h3>
      <p className="muted-2" style={{ fontSize: 14, marginBottom: 32 }}>Manage your DeadLift configuration. Changes made here override onboarding settings.</p>

      {/* Connection info */}
      <div className="surface" style={{ padding: 18, marginBottom: 20 }}>
        <div className="eyebrow" style={{ marginBottom: 12, fontSize: 10 }}>Connection</div>
        <ReviewRow2 label="GCP Project" value="acme-payments-prod" />
        <ReviewRow2 label="DLQ Subscription" value="payments-events-dlq-sub" />
        <ReviewRow2 label="Main Topic" value="payments-events" />
        <ReviewRow2 label="Service account" value="deadlift-sa@acme-payments-prod.iam.gserviceaccount.com" last />
      </div>

      {/* Approval mode */}
      <div className="surface" style={{ padding: 18, marginBottom: 20 }}>
        <div className="eyebrow" style={{ marginBottom: 12, fontSize: 10 }}>Approval mode</div>
        <div style={{ display: 'flex', gap: 10 }}>
          <button className="btn btn-sm" style={{ borderColor: 'var(--text)', background: 'var(--surface-3)' }}>Human-in-the-loop</button>
          <button className="btn btn-sm btn-ghost">Fully autonomous</button>
        </div>
      </div>

      {/* Danger zone */}
      <div style={{ border: '1px solid var(--red-line)', borderRadius: 12, padding: 18, background: 'var(--red-bg)' }}>
        <div className="eyebrow" style={{ color: 'var(--red)', marginBottom: 8, fontSize: 10 }}>Danger zone</div>
        <h4 style={{ fontSize: 15, marginBottom: 6 }}>Terminate DeadLift</h4>
        <p className="muted-2" style={{ fontSize: 13, marginBottom: 14 }}>This will revoke all IAM permissions granted to our service account, stop monitoring your DLQ, and delete all data. This cannot be undone.</p>
        <button className="btn btn-danger btn-sm" onClick={() => setShowConfirm(true)}>Terminate service</button>
      </div>

      {/* Confirm modal */}
      {showConfirm && (
        <div className="modal-backdrop" onClick={() => setShowConfirm(false)}>
          <div className="modal" onClick={e => e.stopPropagation()}>
            <h3 style={{ fontSize: 18, marginBottom: 8 }}>Are you sure?</h3>
            <p className="muted-2" style={{ fontSize: 13.5, marginBottom: 20, lineHeight: 1.55 }}>
              This will remove <strong style={{ color: 'var(--text)' }}>roles/pubsub.subscriber</strong>, <strong style={{ color: 'var(--text)' }}>roles/pubsub.publisher</strong>, and <strong style={{ color: 'var(--text)' }}>roles/logging.viewer</strong> from our service account on your GCP project. All pending fixes will be discarded.
            </p>
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
              <button className="btn btn-sm" onClick={() => setShowConfirm(false)}>Cancel</button>
              <button className="btn btn-danger btn-sm" onClick={handleTerminate} disabled={terminating}>
                {terminating ? 'Terminating…' : 'Yes, terminate'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

const ReviewRow2 = ({ label, value, last }) => (
  <div style={{ display: 'flex', justifyContent: 'space-between', padding: '8px 0', borderBottom: last ? 'none' : '1px solid var(--line)', gap: 16, flexWrap: 'wrap' }}>
    <span className="muted-2" style={{ fontSize: 13 }}>{label}</span>
    <span className="mono" style={{ fontSize: 12.5, color: 'var(--text)', wordBreak: 'break-all' }}>{value}</span>
  </div>
);

Object.assign(window, { AnalyticsTab, RCATab, SettingsTab });
