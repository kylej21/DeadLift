// ============================================================
// Auth — sign in / sign up stubs
// ============================================================

const GoogleIcon = () => (
  <svg width="16" height="16" viewBox="0 0 24 24">
    <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
    <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
    <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
    <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
  </svg>
);

const AuthShell = ({ title, subtitle, children, footer }) => (
  <div style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
    <TopNav variant="public" />
    <div style={{ position: 'relative', flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', padding: '40px 24px' }}>
      <div className="hero-bg" />
      <div style={{ position: 'relative', zIndex: 1, width: '100%', maxWidth: 420 }}>
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <h1 style={{ fontSize: 28, letterSpacing: '-0.02em', marginBottom: 8 }}>{title}</h1>
          <p className="muted-2" style={{ fontSize: 14 }}>{subtitle}</p>
        </div>
        <div className="surface" style={{ padding: 28 }}>{children}</div>
        <div style={{ textAlign: 'center', marginTop: 20, fontSize: 13, color: 'var(--text-3)' }}>{footer}</div>
      </div>
    </div>
  </div>
);

const SignIn = () => (
  <AuthShell title="Welcome back" subtitle="Sign in to your DeadLift dashboard."
    footer={<>New here? <a href="#/signup" className="link">Create an account</a></>}>
    <button className="btn" style={{ width: '100%', background: '#fff', color: '#0a0b0d', borderColor: '#fff' }}
      onClick={() => {
        if (window.session.orgId) {
          location.hash = '#/app';
        } else {
          window.api.signInWithGoogle().catch(err => alert('Sign-in failed: ' + err.message));
        }
      }}>
      <GoogleIcon /> Continue with Google
    </button>
  </AuthShell>
);

const SignUp = () => (
  <AuthShell title="Get started with DeadLift" subtitle="Connect your GCP project and start repairing dead letters automatically."
    footer={<>Already have an account? <a href="#/signin" className="link">Sign in</a></>}>
    <button className="btn" style={{ width: '100%', background: '#fff', color: '#0a0b0d', borderColor: '#fff' }}
      onClick={() => { window.session.setUser({ email: 'new@acme.com', name: 'new', org: null }); location.hash = '#/onboarding'; }}>
      <GoogleIcon /> Sign up with Google
    </button>
    <p className="muted" style={{ fontSize: 11.5, marginTop: 14, textAlign: 'center' }}>By signing up you agree to our Terms and Privacy Policy.</p>
  </AuthShell>
);

Object.assign(window, { SignIn, SignUp, GoogleIcon });
