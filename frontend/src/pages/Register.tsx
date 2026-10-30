import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { AlertCircle, Check, Eye, EyeOff, Loader2 } from 'lucide-react';
import { AuthLayout } from './Login';
import { useAuth } from '../hooks/useAuth';
import { GoogleSignIn } from '../components/GoogleSignIn';

const isEmail = (value: string) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);

export const Register: React.FC = () => {
  const navigate = useNavigate();
  const { register, googleLogin, error, loading } = useAuth();
  const [fullName, setFullName] = useState(''); const [email, setEmail] = useState(''); const [password, setPassword] = useState(''); const [accepted, setAccepted] = useState(false); const [showPassword, setShowPassword] = useState(false); const [touched, setTouched] = useState({ fullName: false, email: false, password: false }); const [submitAttempted, setSubmitAttempted] = useState(false);
  const fullNameError = touched.fullName && !fullName.trim() ? 'Enter your full name.' : '';
  const emailError = touched.email && !isEmail(email) ? 'Enter a valid email address.' : '';
  const passwordError = touched.password && password.length < 8 ? 'Password must be at least 8 characters.' : '';
  const valid = Boolean(fullName.trim()) && isEmail(email) && password.length >= 8 && accepted;
  const isSubmitting = loading && submitAttempted;
  const submit = async (event: React.FormEvent) => { event.preventDefault(); setTouched({ fullName: true, email: true, password: true }); if (!valid) return; setSubmitAttempted(true); try { await register({ fullName: fullName.trim(), email, password }); navigate('/dashboard'); } catch { /* hook error renders below */ } };
  const handleGoogleCredential = async (credential: string) => { try { await googleLogin(credential); navigate('/dashboard'); } catch { /* hook error renders below */ } };
  const fieldClass = (invalid: boolean, isValid = false) => `h-11 w-full rounded border bg-[#EFE6D6] px-3 pr-10 text-[15px] text-[#3A2A1C] outline-none transition focus:ring-2 focus:ring-[#B08D57]/40 ${invalid ? 'border-[#B5573C]' : isValid ? 'border-[#7A8B6F]' : 'border-[#D8C9B2]'}`;
  return <AuthLayout register>
    <header><h1 className="font-serif text-[32px] font-semibold tracking-[-0.03em] text-[#3A2A1C]">Create your account.</h1><p className="mt-2 text-[15px] text-[#8A7B6B]">Start tailoring CVs that actually get responses.</p></header>
    <form className="mt-6 space-y-6" onSubmit={submit} noValidate>
      <div><label htmlFor="fullName" className="mb-2 block text-sm font-medium text-[#3A2A1C]">Full name</label><input id="fullName" autoComplete="name" value={fullName} onChange={(e) => setFullName(e.target.value)} onBlur={() => setTouched((x) => ({ ...x, fullName: true }))} placeholder="Jane Doe" className={fieldClass(Boolean(fullNameError), touched.fullName && Boolean(fullName.trim()))} />{fullNameError ? <p className="mt-1 text-xs text-[#B5573C]">{fullNameError}</p> : null}</div>
      <div><label htmlFor="email" className="mb-2 block text-sm font-medium text-[#3A2A1C]">Email address</label><div className="relative"><input id="email" type="email" autoComplete="email" value={email} onChange={(e) => setEmail(e.target.value)} onBlur={() => setTouched((x) => ({ ...x, email: true }))} placeholder="you@email.com" className={fieldClass(Boolean(emailError), touched.email && isEmail(email))} />{touched.email && isEmail(email) ? <Check className="absolute right-3 top-1/2 h-[18px] w-[18px] -translate-y-1/2 text-[#7A8B6F]" /> : null}</div>{emailError ? <p className="mt-1 text-xs text-[#B5573C]">{emailError}</p> : null}</div>
      <div><label htmlFor="password" className="mb-2 block text-sm font-medium text-[#3A2A1C]">Password</label><div className="relative"><input id="password" type={showPassword ? 'text' : 'password'} autoComplete="new-password" value={password} onChange={(e) => setPassword(e.target.value)} onBlur={() => setTouched((x) => ({ ...x, password: true }))} placeholder="••••••••" className={fieldClass(Boolean(passwordError))} /><button type="button" onClick={() => setShowPassword(!showPassword)} className="absolute right-3 top-1/2 -translate-y-1/2 text-[#8A7B6B]" aria-label={showPassword ? 'Hide password' : 'Show password'}>{showPassword ? <EyeOff className="h-[18px] w-[18px]" /> : <Eye className="h-[18px] w-[18px]" />}</button></div>{passwordError ? <p className="mt-1.5 text-xs text-[#B5573C]">{passwordError}</p> : <p className="mt-1.5 text-xs text-[#8A7B6B]">At least 8 characters.</p>}</div>
      <label className="flex cursor-pointer items-start gap-2 text-[13px] leading-5 text-[#3A2A1C]"><input type="checkbox" checked={accepted} onChange={(e) => setAccepted(e.target.checked)} className="mt-0.5 h-4 w-4 rounded-sm border-[#D8C9B2] accent-[#5C3A21]" />I agree to the <a href="/terms" className="underline">Terms of Service</a> and <a href="/privacy" className="underline">Privacy Policy.</a></label>
      {error ? <div role="alert" className="flex gap-2 rounded border border-[#B5573C] bg-[#B5573C]/10 p-3 text-sm text-[#3A2A1C]"><AlertCircle className="mt-0.5 h-4 w-4 shrink-0 text-[#B5573C]" />{error || 'That email is already in use.'}</div> : null}
      <button type="submit" disabled={!valid || isSubmitting} className="flex h-11 w-full items-center justify-center rounded bg-[#5C3A21] text-[15px] font-semibold text-[#F6F0E6] transition hover:bg-[#4A2F1A] hover:shadow-[0px_4px_16px_rgba(92,58,33,0.14)] disabled:pointer-events-none disabled:opacity-50">{isSubmitting ? <Loader2 className="h-4 w-4 animate-spin" /> : 'Create account'}</button>
    </form>
    <GoogleSignIn onCredential={handleGoogleCredential} busy={loading} /><p className="mt-7 text-center text-sm text-[#8A7B6B]">Already have an account? <Link to="/login" className="font-semibold text-[#3A2A1C] hover:underline">Log in</Link></p>
  </AuthLayout>;
};
export default Register;
