import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { AlertCircle, Check, Eye, EyeOff, Loader2, ShieldCheck, Settings, Zap } from 'lucide-react';
import { useAuth } from '../hooks/useAuth';
import { GoogleSignIn } from '../components/GoogleSignIn';

type AuthLayoutProps = {
  children: React.ReactNode;
  register?: boolean;
};

const Brand = ({ mobile = false }: { mobile?: boolean }) => (
  <Link to="/" className={mobile ? 'inline-block text-center' : 'inline-block'}>
    <span className="font-serif text-[24px] tracking-[-0.03em] text-[#3A2A1C]">Skill-match</span>
    <span className="mt-1 block border-b-2 border-[#B08D57]" />
  </Link>
);

const ArcDecoration = () => (
  <svg className="pointer-events-none absolute -bottom-28 -left-28 h-72 w-72 text-[#B08D57] opacity-20" viewBox="0 0 300 300" fill="none" aria-hidden="true">
    <circle cx="150" cy="150" r="72" stroke="currentColor" />
    <circle cx="150" cy="150" r="105" stroke="currentColor" />
    <circle cx="150" cy="150" r="138" stroke="currentColor" />
  </svg>
);

export const AuthLayout: React.FC<AuthLayoutProps> = ({ children, register = false }) => {
  const rows = register
    ? [
        ['Your data stays private', 'Enterprise-grade encryption keeps your history yours alone.', ShieldCheck],
        ['Gets smarter with every application', 'Our engine adapts to industry trends as you apply.', Settings],
        ['Tailored in seconds, not hours.', 'Generate precise, targeted resumes without the manual rework.', Zap],
      ]
    : [
        ['Your data stays private', '', ShieldCheck],
        ['Gets smarter with every application', '', Settings],
        ['Tailored in seconds, not hours', '', Zap],
      ];

  return (
    <div className="flex min-h-screen bg-[#F6F0E6]">
      <aside className="relative hidden min-h-screen w-[45%] flex-col overflow-hidden bg-[#E3D7C4] p-12 lg:flex xl:p-16">
        <Brand />
        <div className="mt-auto mb-[12%] max-w-[380px]">
          <h2 className="font-serif text-[34px] font-semibold leading-[1.16] tracking-[-0.03em] text-[#3A2A1C]">Your career, with a memory that never forgets what worked.</h2>
          <p className="mt-4 text-[15px] text-[#8A7B6B]">Every application teaches your CV something new.</p>
          <div className="mt-9 space-y-5">
            {rows.map(([title, description, Icon]) => {
              const FeatureIcon = Icon as typeof ShieldCheck;
              return (
                <div className="flex gap-3" key={title as string}>
                  <FeatureIcon className="mt-0.5 h-5 w-5 shrink-0 text-[#B08D57]" strokeWidth={1.7} />
                  <div className="text-[14px] leading-5 text-[#3A2A1C]">
                    <p className={register ? 'font-semibold' : ''}>{title as string}</p>
                    {description ? <p className="mt-0.5 text-[13px] leading-[1.45] text-[#8A7B6B]">{description as string}</p> : null}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
        <ArcDecoration />
      </aside>
      <main className="flex min-h-screen w-full items-center justify-center px-6 py-8 lg:w-[55%] lg:px-12">
        <div className="w-full max-w-[400px]">
          <div className="pb-8 text-center lg:hidden"><Brand mobile /></div>
          {children}
        </div>
      </main>
    </div>
  );
};

const isEmail = (value: string) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);

export const Login: React.FC = () => {
  const navigate = useNavigate();
  const { login, googleLogin, error, loading } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [touched, setTouched] = useState({ email: false, password: false });
  const [submitAttempted, setSubmitAttempted] = useState(false);

  const emailError = touched.email && !isEmail(email) ? 'Enter a valid email address.' : '';
  const passwordError = touched.password && !password ? 'Enter your password.' : '';
  const isSubmitting = loading && submitAttempted;

  const submit = async (event: React.FormEvent) => {
    event.preventDefault();
    setTouched({ email: true, password: true });
    if (!isEmail(email) || !password) return;
    setSubmitAttempted(true);
    try {
      await login({ email, password });
      navigate('/dashboard');
    } catch {
      // The hook exposes the backend error for the inline banner.
    }
  };

  const handleGoogleCredential = async (credential: string) => {
    try {
      await googleLogin(credential);
      navigate('/dashboard');
    } catch {
      // The hook exposes the backend error for the inline banner.
    }
  };

  return (
    <AuthLayout>
      <header><h1 className="font-serif text-[32px] font-semibold tracking-[-0.03em] text-[#3A2A1C]">Welcome back.</h1><p className="mt-2 text-[15px] text-[#8A7B6B]">Log in to keep building your tailored applications.</p></header>
      <form className="mt-6 space-y-6" onSubmit={submit} noValidate>
        <div>
          <label htmlFor="email" className="mb-2 block text-sm font-medium text-[#3A2A1C]">Email</label>
          <div className="relative"><input id="email" type="email" autoComplete="email" value={email} onChange={(e) => setEmail(e.target.value)} onBlur={() => setTouched((value) => ({ ...value, email: true }))} placeholder="you@example.com" className={`h-11 w-full rounded border bg-[#EFE6D6] px-3 pr-10 text-[15px] text-[#3A2A1C] outline-none transition focus:ring-2 focus:ring-[#B08D57]/40 ${emailError ? 'border-[#B5573C]' : touched.email && isEmail(email) ? 'border-[#7A8B6F]' : 'border-[#D8C9B2]'}`} />{touched.email && isEmail(email) ? <Check className="absolute right-3 top-1/2 h-[18px] w-[18px] -translate-y-1/2 text-[#7A8B6F]" /> : null}</div>
          {emailError ? <p className="mt-1 text-xs text-[#B5573C]">{emailError}</p> : null}
        </div>
        <div>
          <div className="mb-2 flex justify-between"><label htmlFor="password" className="text-sm font-medium text-[#3A2A1C]">Password</label><a href="#" className="text-sm font-medium text-[#5C3A21] hover:underline">Forgot password?</a></div>
          <div className="relative"><input id="password" type={showPassword ? 'text' : 'password'} autoComplete="current-password" value={password} onChange={(e) => setPassword(e.target.value)} onBlur={() => setTouched((value) => ({ ...value, password: true }))} placeholder="••••••••" className={`h-11 w-full rounded border bg-[#EFE6D6] px-3 pr-10 text-[15px] text-[#3A2A1C] outline-none transition focus:ring-2 focus:ring-[#B08D57]/40 ${passwordError ? 'border-[#B5573C]' : 'border-[#D8C9B2]'}`} /><button type="button" onClick={() => setShowPassword(!showPassword)} className="absolute right-3 top-1/2 -translate-y-1/2 text-[#8A7B6B]" aria-label={showPassword ? 'Hide password' : 'Show password'}>{showPassword ? <EyeOff className="h-[18px] w-[18px]" /> : <Eye className="h-[18px] w-[18px]" />}</button></div>
          {passwordError ? <p className="mt-1 text-xs text-[#B5573C]">{passwordError}</p> : null}
        </div>
        {error ? <div role="alert" className="flex gap-2 rounded border border-[#B5573C] bg-[#B5573C]/10 p-3 text-sm text-[#3A2A1C]"><AlertCircle className="mt-0.5 h-4 w-4 shrink-0 text-[#B5573C]" />{error || "That email and password don't match."}</div> : null}
        <button type="submit" disabled={isSubmitting} className="flex h-11 w-full items-center justify-center rounded bg-[#5C3A21] text-[15px] font-semibold text-[#F6F0E6] transition hover:bg-[#4A2F1A] hover:shadow-[0px_4px_16px_rgba(92,58,33,0.14)] disabled:pointer-events-none disabled:opacity-70">{isSubmitting ? <Loader2 className="h-4 w-4 animate-spin" /> : 'Log in'}</button>
      </form>
      <GoogleSignIn onCredential={handleGoogleCredential} busy={loading} />
      <p className="mt-7 text-center text-sm text-[#8A7B6B]">Don&apos;t have an account? <Link to="/register" className="font-semibold text-[#3A2A1C] hover:underline">Sign up</Link></p>
    </AuthLayout>
  );
};

export default Login;
