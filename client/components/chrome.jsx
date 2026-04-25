// ============================================================
// Shared chrome: TopNav, Footer, Logo, etc.
// ============================================================

const Logo = ({ size = 'md' }) => (
  <a href="#/" className="logo" style={{ fontSize: size === 'lg' ? 17 : 15 }}>
    <span className="logo-mark" style={{ width: size === 'lg' ? 26 : 22, height: size === 'lg' ? 26 : 22, fontSize: size === 'lg' ? 14 : 13 }}>D</span>
    <span>DeadLift</span>
  </a>
);

const TopNav = ({ variant = 'public' }) => {
  const onSignOut = async () => {
    await window.api.signOut();
    window.session.setUser(null);
    location.hash = '#/';
  };
  return (
    <nav className="topnav">
      <div className="container" style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', height: 60 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 28 }}>
          <Logo />
          {variant === 'public' && (
            <div style={{ display: 'flex', gap: 4 }}>
              <a href="#features" className="btn-ghost btn btn-sm">How it works</a>
              <a href="#stats" className="btn-ghost btn btn-sm">The gap</a>
              <a href="#demo" className="btn-ghost btn btn-sm">Live demo</a>
            </div>
          )}
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          {variant === 'public' && (<>
            <a href="#/signin" className="btn btn-ghost btn-sm">Sign in</a>
            <a href="#/signup" className="btn btn-primary btn-sm">Get started <ArrowIcon /></a>
          </>)}
          {variant === 'app' && (<>
            <span className="pill"><span className="dot pulse" style={{ background: 'var(--green)' }} />Live</span>
            <span className="muted-2" style={{ fontSize: 13 }}>{window.session.user?.email}</span>
            <button className="btn btn-sm btn-ghost" onClick={onSignOut}>Sign out</button>
          </>)}
        </div>
      </div>
    </nav>
  );
};

const ArrowIcon = ({ size = 14 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
    <path d="M5 12h14M13 6l6 6-6 6" />
  </svg>
);

const CheckIcon = ({ size = 14, color = 'currentColor' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d="M20 6L9 17l-5-5" />
  </svg>
);

const XIcon = ({ size = 14 }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
    <path d="M18 6L6 18M6 6l12 12" />
  </svg>
);

// minimal asJSX confidence ring
const ConfidenceRing = ({ value, size = 36 }) => {
  const r = size / 2 - 3;
  const c = 2 * Math.PI * r;
  const off = c * (1 - value);
  const color = value > 0.9 ? 'var(--green)' : value > 0.7 ? 'var(--amber)' : 'var(--red)';
  return (
    <div style={{ position: 'relative', width: size, height: size }}>
      <svg width={size} height={size}>
        <circle cx={size/2} cy={size/2} r={r} fill="none" stroke="rgba(255,255,255,0.08)" strokeWidth="3" />
        <circle cx={size/2} cy={size/2} r={r} fill="none" stroke={color} strokeWidth="3"
          strokeDasharray={c} strokeDashoffset={off} strokeLinecap="round"
          transform={`rotate(-90 ${size/2} ${size/2})`} />
      </svg>
      <div style={{ position: 'absolute', inset: 0, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 10, fontWeight: 600, color: 'var(--text)' }}>
        {Math.round(value*100)}
      </div>
    </div>
  );
};

Object.assign(window, { Logo, TopNav, ArrowIcon, CheckIcon, XIcon, ConfidenceRing });
