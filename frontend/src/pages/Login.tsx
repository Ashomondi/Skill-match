import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Eye, EyeOff, Mail, Lock, Loader as Loader2, CircleAlert as AlertCircle } from 'lucide-react';
import { useAuth } from '../hooks/useAuth';

interface FormErrors {
  email?: string;
  password?: string;
}

export const Login: React.FC = () => {
  const navigate = useNavigate();
  const { login } = useAuth();

  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [errors, setErrors] = useState<FormErrors>({});
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const validate = (): FormErrors => {
    const newErrors: FormErrors = {};

    if (!email.trim()) {
      newErrors.email = 'Email is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      newErrors.email = 'Enter a valid email address';
    }

    if (!password) {
      newErrors.password = 'Password is required';
    } else if (password.length < 6) {
      newErrors.password = 'Password must be at least 6 characters';
    }

    return newErrors;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitError(null);

    const validationErrors = validate();
    setErrors(validationErrors);

    if (Object.keys(validationErrors).length > 0) return;

    setIsSubmitting(true);
    try {
      await login({ email, password });
      navigate('/dashboard');
    } catch (err: any) {
      setSubmitError(err.message || 'Unable to sign in. Please check your credentials.');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleEmailChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setEmail(e.target.value);
    if (errors.email) setErrors({ ...errors, email: undefined });
  };

  const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setPassword(e.target.value);
    if (errors.password) setErrors({ ...errors, password: undefined });
  };

  return (
    <div className="min-h-screen flex">
      {/* Left panel — brand / image */}
      <div className="hidden lg:flex lg:w-1/2 relative bg-gradient-to-br from-[#2C2A29] to-[#3D3733]">
        <div className="absolute inset-0 flex flex-col justify-between p-12 text-[#EAE5DC]">
          <Link to="/" className="text-2xl font-bold tracking-tight">
            SkillMatch
          </Link>

          <div className="space-y-6 max-w-md">
            <h1 className="text-4xl font-bold leading-tight">
              Your career, with a memory that never forgets what worked.
            </h1>
            <p className="text-base text-[#C2BBB0] leading-relaxed">
              Sign in to access your tailored CVs, AI-powered job matches, and an application
              tracker that learns from every step you take.
            </p>
          </div>

          <div className="flex items-center gap-2 text-sm text-[#8C8275]">
            <span className="h-2 w-2 rounded-full bg-emerald-400" />
            Memory layer healthy
          </div>
        </div>
      </div>

      {/* Right panel — form */}
      <div className="flex-1 flex items-center justify-center bg-[#EAE5DC] px-6 py-12 sm:px-12">
        <div className="w-full max-w-md">
          {/* Mobile brand */}
          <Link to="/" className="lg:hidden block text-2xl font-bold text-[#2C2A29] mb-8">
            SkillMatch
          </Link>

          <div className="mb-8">
            <h2 className="text-2xl font-bold text-[#2C2A29]">Welcome back</h2>
            <p className="mt-2 text-sm text-[#6B655D]">
              Sign in to continue to your dashboard.
            </p>
          </div>

          {submitError && (
            <div
              role="alert"
              className="mb-6 flex items-start gap-2 rounded border border-red-300 bg-red-50 px-4 py-3 text-sm text-red-700"
            >
              <AlertCircle className="h-4 w-4 mt-0.5 flex-shrink-0" />
              <span>{submitError}</span>
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-5" noValidate>
            {/* Email */}
            <div>
              <label htmlFor="email" className="block text-sm font-medium text-[#2C2A29] mb-1.5">
                Email
              </label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-[#8C8275]" />
                <input
                  id="email"
                  name="email"
                  type="email"
                  autoComplete="email"
                  value={email}
                  onChange={handleEmailChange}
                  placeholder="you@example.com"
                  aria-invalid={!!errors.email}
                  aria-describedby={errors.email ? 'email-error' : undefined}
                  className={`w-full rounded border bg-[#F5F1E9] py-2.5 pl-10 pr-3 text-sm text-[#2C2A29] placeholder:text-[#8C8275] focus:outline-none focus:ring-2 focus:ring-[#594433] focus:border-transparent transition ${
                    errors.email ? 'border-red-400' : 'border-[#C2BBB0]'
                  }`}
                />
              </div>
              {errors.email && (
                <p id="email-error" className="mt-1.5 text-xs text-red-600">
                  {errors.email}
                </p>
              )}
            </div>

            {/* Password */}
            <div>
              <label htmlFor="password" className="block text-sm font-medium text-[#2C2A29] mb-1.5">
                Password
              </label>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-[#8C8275]" />
                <input
                  id="password"
                  name="password"
                  type={showPassword ? 'text' : 'password'}
                  autoComplete="current-password"
                  value={password}
                  onChange={handlePasswordChange}
                  placeholder="Enter your password"
                  aria-invalid={!!errors.password}
                  aria-describedby={errors.password ? 'password-error' : undefined}
                  className={`w-full rounded border bg-[#F5F1E9] py-2.5 pl-10 pr-10 text-sm text-[#2C2A29] placeholder:text-[#8C8275] focus:outline-none focus:ring-2 focus:ring-[#594433] focus:border-transparent transition ${
                    errors.password ? 'border-red-400' : 'border-[#C2BBB0]'
                  }`}
                />
                <button
                  type="button"
                  onClick={() => setShowPassword((s) => !s)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-[#8C8275] hover:text-[#2C2A29] transition"
                  aria-label={showPassword ? 'Hide password' : 'Show password'}
                  tabIndex={0}
                >
                  {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </button>
              </div>
              {errors.password && (
                <p id="password-error" className="mt-1.5 text-xs text-red-600">
                  {errors.password}
                </p>
              )}
            </div>

            <div className="flex items-center justify-between">
              <label className="flex items-center gap-2 text-sm text-[#6B655D] cursor-pointer">
                <input
                  type="checkbox"
                  className="h-4 w-4 rounded border-[#C2BBB0] text-[#594433] focus:ring-[#594433]"
                />
                Remember me
              </label>
              <a href="#" className="text-sm text-[#594433] hover:underline">
                Forgot password?
              </a>
            </div>

            {/* Submit */}
            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full flex items-center justify-center gap-2 rounded bg-[#594433] py-2.5 text-sm font-medium text-white hover:bg-[#3D3733] focus:outline-none focus:ring-2 focus:ring-[#594433] focus:ring-offset-2 focus:ring-offset-[#EAE5DC] disabled:opacity-60 disabled:cursor-not-allowed transition"
            >
              {isSubmitting && <Loader2 className="h-4 w-4 animate-spin" />}
              {isSubmitting ? 'Signing in...' : 'Sign in'}
            </button>
          </form>

          <p className="mt-8 text-center text-sm text-[#6B655D]">
            Don&apos;t have an account?{' '}
            <Link to="/register" className="font-medium text-[#594433] hover:underline">
              Create one
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
};

export default Login;
