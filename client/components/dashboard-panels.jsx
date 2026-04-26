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
        {/* Subscription breakdown */}
        <div className="surface" style={{ padding: 20 }}>
          <div className="eyebrow" style={{ marginBottom: 16 }}>By subscription</div>
          {data.categories.length === 0
            ? <p className="muted-2" style={{ fontSize: 13, margin: 0 }}>No data yet.</p>
            : data.categories.map(c => (
              <div key={c.name} style={{ marginBottom: 12 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12.5, marginBottom: 5 }}>
                  <span className="mono" style={{ color: 'var(--text-2)', fontSize: 12 }}>{c.name}</span>
                  <span className="mono" style={{ color: 'var(--text)' }}>{c.count} <span className="muted">({c.pct}%)</span></span>
                </div>
                <div style={{ height: 5, background: 'var(--surface-3)', borderRadius: 999, overflow: 'hidden' }}>
                  <div style={{ height: '100%', width: `${c.pct}%`, background: c.color, borderRadius: 999, transition: 'width 500ms ease' }} />
                </div>
              </div>
            ))
          }
        </div>

        {/* Per-topic table */}
        <div className="surface" style={{ padding: 20 }}>
          <div className="eyebrow" style={{ marginBottom: 16 }}>Per-topic breakdown</div>
          {data.topics.length === 0
            ? <p className="muted-2" style={{ fontSize: 13, margin: 0 }}>No data yet.</p>
            : (
              <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 13 }}>
                <thead>
                  <tr style={{ borderBottom: '1px solid var(--line)' }}>
                    <th style={{ textAlign: 'left', padding: '6px 0', color: 'var(--text-3)', fontWeight: 500, fontSize: 11.5 }}>Topic</th>
                    <th style={{ textAlign: 'right', padding: '6px 8px', color: 'var(--text-3)', fontWeight: 500, fontSize: 11.5 }}>DLQ</th>
                    <th style={{ textAlign: 'right', padding: '6px 8px', color: 'var(--text-3)', fontWeight: 500, fontSize: 11.5 }}>Fixed</th>
                    <th style={{ textAlign: 'right', padding: '6px 0', color: 'var(--text-3)', fontWeight: 500, fontSize: 11.5 }}>Fix rate</th>
                  </tr>
                </thead>
                <tbody>
                  {data.topics.map(t => (
                    <tr key={t.name} style={{ borderBottom: '1px solid var(--line)' }}>
                      <td className="mono" style={{ padding: '8px 0', fontSize: 12.5 }}>{t.name}</td>
                      <td style={{ textAlign: 'right', padding: '8px 8px' }}>{t.dlq}</td>
                      <td style={{ textAlign: 'right', padding: '8px 8px', color: 'var(--green)' }}>{t.fixed}</td>
                      <td className="mono" style={{ textAlign: 'right', padding: '8px 0', fontSize: 12, color: t.dlq > 0 && (t.fixed / t.dlq) >= 0.8 ? 'var(--green)' : 'var(--text-3)' }}>
                        {t.dlq > 0 ? `${Math.round((t.fixed / t.dlq) * 100)}%` : '—'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )
          }
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
  const max = Math.max(1, ...series.map(s => s.dlq));
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

const __RCA_CLASS_THEME = {
  missing_field:  { color: 'var(--amber)',  bg: 'var(--amber-bg)',  line: 'var(--amber-line)',  label: 'Missing field' },
  type_mismatch:  { color: 'var(--blue)',   bg: 'var(--blue-bg)',   line: 'oklch(0.72 0.14 240 / 0.30)', label: 'Type mismatch' },
  malformed_json: { color: 'var(--red)',    bg: 'var(--red-bg)',    line: 'var(--red-line)',    label: 'Malformed JSON' },
  schema_drift:   { color: 'var(--green)',  bg: 'var(--green-bg)',  line: 'var(--green-line)',  label: 'Schema drift' },
  encoding:       { color: 'oklch(0.72 0.14 290)', bg: 'oklch(0.72 0.14 290 / 0.10)', line: 'oklch(0.72 0.14 290 / 0.30)', label: 'Encoding' },
  unknown:        { color: 'var(--text-3)', bg: 'var(--surface-2)', line: 'var(--line)',        label: 'Unknown' },
};
const __rcaTheme = (cls) => __RCA_CLASS_THEME[cls] || __RCA_CLASS_THEME.unknown;

const MarkdownContent = ({ text }) => {
  if (!text) return null;

  const inlineFormat = (str, keyPrefix) => {
    const parts = [];
    let remaining = str;
    let k = 0;
    while (remaining.length > 0) {
      const bold = remaining.match(/^([\s\S]*?)\*\*(.+?)\*\*/);
      if (bold) {
        if (bold[1]) parts.push(<span key={keyPrefix + k++}>{bold[1]}</span>);
        parts.push(<strong key={keyPrefix + k++}>{bold[2]}</strong>);
        remaining = remaining.slice(bold[0].length);
        continue;
      }
      const italic = remaining.match(/^([\s\S]*?)\*(.+?)\*/);
      if (italic) {
        if (italic[1]) parts.push(<span key={keyPrefix + k++}>{italic[1]}</span>);
        parts.push(<em key={keyPrefix + k++}>{italic[2]}</em>);
        remaining = remaining.slice(italic[0].length);
        continue;
      }
      const code = remaining.match(/^([\s\S]*?)`(.+?)`/);
      if (code) {
        if (code[1]) parts.push(<span key={keyPrefix + k++}>{code[1]}</span>);
        parts.push(<code key={keyPrefix + k++} className="mono" style={{ fontSize: 12, background: 'var(--surface-3)', padding: '1px 5px', borderRadius: 3 }}>{code[2]}</code>);
        remaining = remaining.slice(code[0].length);
        continue;
      }
      parts.push(<span key={keyPrefix + k++}>{remaining}</span>);
      break;
    }
    return parts;
  };

  const lines = text.split('\n');
  const elements = [];
  let i = 0;

  while (i < lines.length) {
    const line = lines[i];

    // Triple-backtick code fence — consume until closing ```
    if (line.trimStart().startsWith('```')) {
      const codeLines = [];
      i++;
      while (i < lines.length && !lines[i].trimStart().startsWith('```')) {
        codeLines.push(lines[i]);
        i++;
      }
      if (i < lines.length) i++; // skip closing ```
      elements.push(
        <pre key={`pre${i}`} style={{ background: 'var(--surface-3)', borderRadius: 6, padding: '10px 14px', overflowX: 'auto', margin: '8px 0', border: '1px solid var(--line)' }}>
          <code className="mono" style={{ fontSize: 12, color: 'var(--text-2)', whiteSpace: 'pre' }}>{codeLines.join('\n')}</code>
        </pre>
      );
      continue;
    }

    // Headings — any depth (# through ######)
    const hMatch = line.match(/^(#{1,6})\s+(.*)/);
    if (hMatch) {
      const level = hMatch[1].length;
      const hStyle = level <= 2
        ? { fontSize: 14, fontWeight: 700, color: 'var(--text)', margin: '16px 0 6px' }
        : level === 3
          ? { fontSize: 13, fontWeight: 600, color: 'var(--text)', margin: '14px 0 4px' }
          : { fontSize: 12, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--text-3)', margin: '12px 0 4px' };
      elements.push(<p key={i} style={hStyle}>{inlineFormat(hMatch[2], `h${i}-`)}</p>);
      i++; continue;
    }

    // Bullet list — collect consecutive items
    if (line.match(/^[-*]\s+/)) {
      const items = [];
      while (i < lines.length && lines[i].match(/^[-*]\s+/)) {
        items.push(<li key={i} style={{ fontSize: 13, lineHeight: 1.65, color: 'var(--text-2)', marginBottom: 2 }}>{inlineFormat(lines[i].replace(/^[-*]\s+/, ''), `li${i}-`)}</li>);
        i++;
      }
      elements.push(<ul key={`ul${i}`} style={{ margin: '4px 0 8px', paddingLeft: 20 }}>{items}</ul>);
      continue;
    }

    // Numbered list — collect consecutive items
    if (line.match(/^\d+\.\s+/)) {
      const items = [];
      while (i < lines.length && lines[i].match(/^\d+\.\s+/)) {
        items.push(<li key={i} style={{ fontSize: 13, lineHeight: 1.65, color: 'var(--text-2)', marginBottom: 2 }}>{inlineFormat(lines[i].replace(/^\d+\.\s+/, ''), `oli${i}-`)}</li>);
        i++;
      }
      elements.push(<ol key={`ol${i}`} style={{ margin: '4px 0 8px', paddingLeft: 20 }}>{items}</ol>);
      continue;
    }

    // Blank line — skip
    if (line.trim() === '') { i++; continue; }

    // Regular paragraph
    elements.push(<p key={i} style={{ fontSize: 13, lineHeight: 1.7, color: 'var(--text-2)', margin: '4px 0' }}>{inlineFormat(line, `p${i}-`)}</p>);
    i++;
  }

  return <div>{elements}</div>;
};

const RCATab = ({ reports }) => {
  if (!reports) return <LoadingState />;
  if (reports.length === 0) return (
    <div style={{ textAlign: 'center', padding: '80px 24px' }}>
      <div style={{ width: 64, height: 64, borderRadius: 16, background: 'var(--surface-2)', border: '1px solid var(--line)', display: 'inline-flex', alignItems: 'center', justifyContent: 'center', marginBottom: 20 }}>
        <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="var(--text-3)" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
          <circle cx="11" cy="11" r="8"/><path d="M21 21l-4.35-4.35M11 8v3m0 3h.01"/>
        </svg>
      </div>
      <h3 className="h3" style={{ marginBottom: 8 }}>No root cause reports yet</h3>
      <p className="muted-2" style={{ fontSize: 14, maxWidth: 340, margin: '0 auto' }}>
        Open any fix card and click <strong style={{ color: 'var(--text-2)' }}>Generate root cause</strong> to kick off a deep analysis.
      </p>
    </div>
  );
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      {reports.map(r => <RCACard key={`${r.org_id}_${r.message_id}`} report={r} />)}
    </div>
  );
};

const RCACard = ({ report }) => {
  const [expanded, setExpanded] = React.useState(true);
  const theme = __rcaTheme(report.error_class);
  const date = report.created_at ? new Date(report.created_at).toLocaleString() : '—';

  return (
    <div style={{ borderRadius: 12, border: '1px solid var(--line)', background: 'var(--surface-1)', overflow: 'hidden' }}>
      <div
        style={{ padding: '12px 18px', display: 'flex', alignItems: 'center', justifyContent: 'space-between', borderBottom: '1px solid var(--line)', cursor: 'pointer', background: 'var(--surface-2)' }}
        onClick={() => setExpanded(e => !e)}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <span style={{ fontSize: 13, fontWeight: 600, color: theme.color }}>{theme.label}</span>
          <span className="mono" style={{ fontSize: 11, color: 'var(--text-3)' }}>{report.message_id}</span>
          <span style={{ fontSize: 11.5, color: 'var(--text-3)' }}>{date}</span>
        </div>
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5"
             style={{ transition: 'transform 150ms', transform: expanded ? 'rotate(180deg)' : 'none', color: 'var(--text-3)', flexShrink: 0 }}>
          <path d="M6 9l6 6 6-6" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      </div>
      {expanded && (
        <div style={{ padding: '16px 18px' }}>
          <MarkdownContent text={report.analysis} />
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
  const [user, setUser] = React.useState(null);

  React.useEffect(() => {
    window.api.getUser().then(u => setUser(u || false));
  }, []);

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
        {user === null ? (
          <div className="skeleton" style={{ height: 100 }} />
        ) : user === false ? (
          <p className="muted-2" style={{ fontSize: 13, margin: 0 }}>Not connected. Complete onboarding to see connection details.</p>
        ) : (
          <>
            <ReviewRow2 label="GCP Project" value={user.project_id || '—'} />
            <ReviewRow2 label="DLQ Subscription" value={user.dlq_subscription || '—'} />
            <ReviewRow2 label="Main Topic" value={user.main_topic || '—'} />
            <ReviewRow2 label="Email" value={user.email || '—'} />
            <ReviewRow2 label="Org ID" value={user.org_id || '—'} last />
          </>
        )}
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
