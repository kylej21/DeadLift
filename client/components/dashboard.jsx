// ============================================================
// Dashboard — shell, Fixes tab, Batches tab
// ============================================================

const TABS = [
  { key: 'fixes', label: 'Fixes' },
  { key: 'batches', label: 'Batches' },
  { key: 'analytics', label: 'Analytics' },
  { key: 'rca', label: 'Root cause' },
  { key: 'settings', label: 'Settings' },
];

const Dashboard = () => {
  const tweaks = window.__tweakState || {};
  const viewState = tweaks.dashboardState || 'populated'; // empty | loading | populated
  const [tab, setTab] = React.useState('fixes');
  const [fixes, setFixes] = React.useState(null);
  const [batches, setBatches] = React.useState(null);
  const [analytics, setAnalytics] = React.useState(null);
  const [rca, setRca] = React.useState(null);

  React.useEffect(() => {
    if (viewState === 'empty') { setFixes([]); setBatches([]); setAnalytics(null); setRca([]); return; }
    if (viewState === 'loading') { setFixes(null); setBatches(null); setAnalytics(null); setRca(null); return; }
    window.api.getFixes().then(setFixes);
    window.api.getBatches().then(setBatches);
    window.api.getAnalytics().then(setAnalytics);
    window.api.getRCAReports().then(setRca);
  }, [viewState]);

  const pendingCount = fixes ? fixes.filter(f => f.status === 'pending').length : 0;
  const batchPending = batches ? batches.filter(b => b.status === 'pending').length : 0;

  return (
    <div style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
      <TopNav variant="app" />
      <div className="container" style={{ flex: 1, padding: '24px 24px 80px' }}>
        {/* Tabs */}
        <div className="tab-bar" style={{ marginBottom: 24 }}>
          {TABS.map(t => (
            <button key={t.key} className={`tab ${tab === t.key ? 'active' : ''}`} onClick={() => setTab(t.key)}>
              {t.label}
              {t.key === 'fixes' && pendingCount > 0 && <span className="tab-count">{pendingCount}</span>}
              {t.key === 'batches' && batchPending > 0 && <span className="tab-count">{batchPending}</span>}
            </button>
          ))}
        </div>
        <div className="fade-in" key={tab + viewState}>
          {tab === 'fixes' && <FixesTab fixes={fixes} setFixes={setFixes} />}
          {tab === 'batches' && <BatchesTab batches={batches} setBatches={setBatches} />}
          {tab === 'analytics' && <AnalyticsTab data={analytics} />}
          {tab === 'rca' && <RCATab reports={rca} />}
          {tab === 'settings' && <SettingsTab />}
        </div>
      </div>
    </div>
  );
};

// ====== FIXES TAB ======
const FixesTab = ({ fixes, setFixes }) => {
  const [refreshing, setRefreshing] = React.useState(false);

  const refresh = async () => {
    setRefreshing(true);
    const fresh = await window.api.getFixes();
    setFixes(fresh);
    setRefreshing(false);
  };

  const handleApprove = async (id) => {
    await window.api.approveFix(id);
    setFixes(prev => prev.map(f => f.id === id ? { ...f, status: 'fixed', fixedAt: 'just now' } : f));
  };
  const handleDeny = async (id) => {
    await window.api.denyFix(id);
    setFixes(prev => prev.map(f => f.id === id ? { ...f, status: 'denied' } : f));
  };

  const inner = fixes === null ? <LoadingState /> :
    fixes.length === 0 ? <EmptyState title="No DLQ messages yet" desc="Once DeadLift detects failed messages, proposed fixes will appear here." /> :
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      {fixes.map(fix => <FixCard key={fix.id} fix={fix} onApprove={handleApprove} onDeny={handleDeny} />)}
    </div>;

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: 12 }}>
        <button className="btn btn-sm" onClick={refresh} disabled={refreshing}>
          {refreshing ? 'Refreshing…' : 'Refresh'}
        </button>
      </div>
      {inner}
    </div>
  );
};

const FixCard = ({ fix, onApprove, onDeny }) => {
  const [expanded, setExpanded] = React.useState(fix.status === 'pending');
  const statusPill = fix.status === 'fixed'
    ? <span className="pill pill-green"><span className="dot" style={{ background: 'var(--green)' }} />Fixed{fix.fixedAt ? ` · ${fix.fixedAt}` : ''}</span>
    : fix.status === 'denied'
    ? <span className="pill pill-red"><span className="dot" style={{ background: 'var(--red)' }} />Denied</span>
    : <span className="pill pill-amber"><span className="dot pulse" style={{ background: 'var(--amber)' }} />Awaiting approval</span>;

  return (
    <div className="surface" style={{ overflow: 'hidden' }}>
      {/* Header */}
      <div onClick={() => setExpanded(!expanded)} style={{ padding: '14px 18px', display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 12, cursor: 'pointer', flexWrap: 'wrap' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10, flexWrap: 'wrap' }}>
          {statusPill}
          <span className="pill">{fix.category}</span>
          <span className="mono muted" style={{ fontSize: 12 }}>{fix.id}</span>
          {fix.batch && <span className="pill pill-blue" style={{ fontSize: 11 }}>Batch: {fix.batch.count} msgs</span>}
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <ConfidenceRing value={fix.confidence} size={32} />
          <span className="mono muted" style={{ fontSize: 12 }}>{fix.subscription}</span>
          <span className="muted" style={{ fontSize: 12 }}>{fix.receivedAt}</span>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" style={{ transition: 'transform 150ms', transform: expanded ? 'rotate(180deg)' : 'none' }}>
            <path d="M6 9l6 6 6-6" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
        </div>
      </div>

      {expanded && (
        <>
          {/* Error */}
          <div style={{ padding: '0 18px 12px' }}>
            <div style={{ background: 'var(--red-bg)', border: '1px solid var(--red-line)', borderRadius: 8, padding: '8px 12px' }}>
              <code className="mono" style={{ fontSize: 12, color: 'var(--red)', wordBreak: 'break-all' }}>{fix.error}</code>
            </div>
          </div>

          {/* Diff */}
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', borderTop: '1px solid var(--line)', borderBottom: '1px solid var(--line)' }}>
            <FixDiffPane title="Before (original)" json={fix.before} otherJson={fix.after} variant="before" />
            <FixDiffPane title="After (proposed fix)" json={fix.after} otherJson={fix.before} variant="after" />
          </div>

          {/* Sources */}
          {fix.sources && fix.sources.length > 0 && (
            <div style={{ padding: '12px 18px', borderBottom: '1px solid var(--line)' }}>
              <div className="eyebrow" style={{ marginBottom: 8, fontSize: 10 }}>Cited sources</div>
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6 }}>
                {fix.sources.map((s, i) => (
                  <span key={i} className="pill" style={{ fontSize: 11 }}>
                    <SourceIcon kind={s.kind} />{s.name}
                  </span>
                ))}
              </div>
            </div>
          )}

          {/* Actions */}
          {fix.status === 'pending' && (
            <div style={{ padding: '12px 18px', display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
              <button className="btn btn-sm" onClick={() => onDeny(fix.id)}>Deny</button>
              <button className="btn btn-sm">Edit fix</button>
              <button className="btn btn-sm btn-green" onClick={() => onApprove(fix.id)}>
                {fix.batch ? `Approve all ${fix.batch.count}` : 'Approve & republish'}
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
};

const FixDiffPane = ({ title, json, otherJson, variant }) => {
  const lines = (json || '').split('\n');
  const otherLineSet = new Set((otherJson || '').split('\n').map(l => l.trim()));
  return (
    <div style={{ borderRight: variant === 'before' ? '1px solid var(--line)' : 'none' }}>
      <div style={{ padding: '8px 14px', borderBottom: '1px solid var(--line)', fontSize: 11, color: 'var(--text-3)', textTransform: 'uppercase', letterSpacing: '0.08em', display: 'flex', justifyContent: 'space-between' }}>
        <span>{title}</span>
        <span className="mono" style={{ fontSize: 10 }}>JSON</span>
      </div>
      <div style={{ padding: '6px 0', background: 'var(--bg-1)', overflowX: 'auto', maxHeight: 280 }}>
        {lines.map((l, i) => {
          const isChanged = l.trim() !== '' && !otherLineSet.has(l.trim());
          return (
            <div key={i} className={`diff-line ${isChanged ? (variant === 'before' ? 'diff-del' : 'diff-add') : ''}`}>
              <span className="diff-gutter">{i + 1}</span>
              <span className="diff-marker">{isChanged ? (variant === 'before' ? '−' : '+') : ' '}</span>
              <span className="diff-content">{l}</span>
            </div>
          );
        })}
      </div>
    </div>
  );
};

const SourceIcon = ({ kind }) => {
  const map = { runbook: '📄', code: '⌨', log: '📋', context: '🔍', deploy: '🚀', metric: '📊', doc: '📝' };
  return <span style={{ fontSize: 11 }}>{map[kind] || '•'}</span>;
};

// ====== BATCHES TAB ======
const BatchesTab = ({ batches, setBatches }) => {
  if (batches === null) return <LoadingState />;
  if (batches.length === 0) return <EmptyState title="No batches" desc="Batches appear when multiple DLQ messages share a root cause." />;

  const handleApprove = async (id) => {
    await window.api.approveBatch(id);
    setBatches(prev => prev.map(b => b.id === id ? { ...b, status: 'fixed' } : b));
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      {batches.map(b => <BatchCard key={b.id} batch={b} onApprove={handleApprove} />)}
    </div>
  );
};

const BatchCard = ({ batch, onApprove }) => {
  const [expanded, setExpanded] = React.useState(batch.status === 'pending');
  return (
    <div className="surface" style={{ overflow: 'hidden' }}>
      <div onClick={() => setExpanded(!expanded)} style={{ padding: '16px 18px', display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 12, cursor: 'pointer', flexWrap: 'wrap' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10, flexWrap: 'wrap' }}>
          {batch.status === 'fixed'
            ? <span className="pill pill-green"><span className="dot" style={{ background: 'var(--green)' }} />Fixed</span>
            : <span className="pill pill-amber"><span className="dot pulse" style={{ background: 'var(--amber)' }} />Pending</span>}
          <span className="pill">{batch.category}</span>
          <span style={{ fontSize: 14, fontWeight: 500 }}>{batch.title}</span>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <span className="pill pill-blue" style={{ fontSize: 12 }}>{batch.affectedCount} messages</span>
          <ConfidenceRing value={batch.confidence} size={32} />
        </div>
      </div>
      {expanded && (
        <div style={{ padding: '0 18px 16px' }}>
          <div className="surface" style={{ padding: 16, background: 'var(--surface-2)', marginBottom: 12 }}>
            <div className="eyebrow" style={{ marginBottom: 6, fontSize: 10 }}>Root cause</div>
            <p style={{ fontSize: 13.5, color: 'var(--text-2)', margin: 0, lineHeight: 1.55 }}>{batch.rootCause}</p>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 12, marginBottom: 12 }}>
            <MiniStat label="Affected messages" value={batch.affectedCount} />
            <MiniStat label="First seen" value={batch.firstSeen} />
            <MiniStat label="Topics" value={batch.affectedTopics.join(', ')} />
          </div>
          <div style={{ background: 'var(--green-bg)', border: '1px solid var(--green-line)', borderRadius: 8, padding: '10px 14px', marginBottom: 14 }}>
            <div className="eyebrow" style={{ fontSize: 10, color: 'var(--green)', marginBottom: 4 }}>Proposed fix</div>
            <div style={{ fontSize: 13, color: 'var(--text)' }}>{batch.fixSummary}</div>
          </div>
          {batch.status === 'pending' && (
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
              <button className="btn btn-sm">Deny batch</button>
              <button className="btn btn-sm btn-green" onClick={() => onApprove(batch.id)}>Approve all {batch.affectedCount}</button>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

const MiniStat = ({ label, value }) => (
  <div className="surface" style={{ padding: '10px 12px', background: 'var(--surface-2)' }}>
    <div className="eyebrow" style={{ fontSize: 10, marginBottom: 4 }}>{label}</div>
    <div className="mono" style={{ fontSize: 14, fontWeight: 600 }}>{value}</div>
  </div>
);

// ====== Shared states ======
const LoadingState = () => (
  <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
    {[1, 2, 3].map(i => <div key={i} className="skeleton" style={{ height: 100 }} />)}
  </div>
);

const EmptyState = ({ title, desc }) => (
  <div style={{ textAlign: 'center', padding: '80px 24px' }}>
    <div style={{ width: 56, height: 56, borderRadius: 14, background: 'var(--surface-2)', border: '1px solid var(--line)', display: 'inline-flex', alignItems: 'center', justifyContent: 'center', marginBottom: 16 }}>
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="var(--text-3)" strokeWidth="1.5"><path d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" strokeLinecap="round" strokeLinejoin="round"/></svg>
    </div>
    <h3 className="h3" style={{ marginBottom: 6 }}>{title}</h3>
    <p className="muted-2" style={{ fontSize: 14 }}>{desc}</p>
  </div>
);

Object.assign(window, { Dashboard, LoadingState, EmptyState, MiniStat, SourceIcon });
