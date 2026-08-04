import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Eye, EyeOff, Mail, Lock, User, Loader as Loader2, CircleAlert as AlertCircle, CircleCheck as CheckCircle2 } from 'lucide-react';
import { useAuth } from '../hooks/useAuth';

interface FormErrors {
  fullName?: string;
  email?: string;
  password?: string;
  confirmPassword?: string;
}

export const Register: React.FC = () => {
  const navigate = useNavigate();
  const { register } = useAuth();

  const [fullName, setFullName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [errors, setErrors] = useState<FormErrors>({});
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const passwordChecks = {
    length: password.length >= 8,
    hasUpper: /[A-Z]/.test(password),
    hasNumber: /[0-9]/.test(password),
  };

  const validate = (): FormErrors => {
    const newErrors: FormErrors = {};

    if (!fullName.trim()) {
      newErrors.fullName = 'Full name is required';
    } else if (fullName.trim().length < 2) {
      newErrors.fullName = 'Enter your full name';
    }

    if (!email.trim()) {
      newErrors.email = 'Email is required';
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      newErrors.email = 'Enter a valid email address';
    }

    if (!password) {
      newErrors.password = 'Password is required';
    } else if (password.length < 8) {
      newErrors.password = 'Password must be at least 8 characters';
    }

    if (!confirmPassword) {
      newErrors.confirmPassword = 'Please confirm your password';
    } else if (password !== confirmPassword) {
      newErrors.confirmPassword = 'Passwords do not match';
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
      await register({ email, password, fullName });
      navigate('/dashboard');
    } catch (err: any) {
      setSubmitError(err.message || 'Unable to create account. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  const clearError = (field: keyof FormErrors) => {
    if (errors[field]) setErrors({ ...errors, [field]: undefined });
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
              Start your journey with a career copilot that remembers.
            </h1>
            <p className="text-base text-[#C2BBB0] leading-relaxed">
              Create an account to upload your CV, chat with an AI assistant that learns your
              strengths, and get job recommendations that get smarter over time.
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
            <h2 className="text-2xl font-bold text-[#2C2A29]">Create your account</h2>
            <p className="mt-2 text-sm text-[#6B655D]">
              Join SkillMatch and let your career memory begin.
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
            {/* Full Name */}
            <div>
              <label htmlFor="fullName" className="block text-sm font-medium text-[#2C2A29] mb-1.5">
                Full name
              </label>
              <div className="relative">
                <User className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-[#8C8275]" />
                <input
                  id="fullName"
                  name="fullName"
                  type="text"
                  autoComplete="name"
                  value={fullName}
                  onChange={(e) => {
                    setFullName(e.target.value);
                    clearError('fullName');
                  }}
                  placeholder="Jane Doe"
                  aria-invalid={!!errors.fullName}
                  aria-describedby={errors.fullName ? 'fullName-error' : undefined}
                  className={`w-full rounded border bg-[#F5F1E9] py-2.5 pl-10 pr-3 text-sm text-[#2C2A29] placeholder:text-[#8C8275] focus:outline-none focus:ring-2 focus:ring-[#594433] focus:border-transparent transition ${
                    errors.fullName ? 'border-red-400' : 'border-[#C2BBB0]'
                  }`}
                />
              </div>
              {errors.fullName && (
                <p id="fullName-error" className="mt-1.5 text-xs text-red-600">
                  {errors.fullName}
                </p>
              )}
            </div>

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
                  onChange={(e) => {
                    setEmail(e.target.value);
                    clearError('email');
                  }}
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
                  autoComplete="new-password"
                  value={password}
                  onChange={(e) => {
                    setPassword(e.target.value);
                    clearError('password');
                  }}
                  placeholder="Create a password"
                  aria-invalid={!!errors.password}
                  aria-describedby={errors.password ? 'password-error' : 'password-strength'}
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
              {errors.password ? (
                <p id="password-error" className="mt-1.5 text-xs text-red-600">
                  {errors.password}
                </p>
              ) : (
                password.length > 0 && (
                  <ul id="password-strength" className="mt-2 space-y-1">
                    <li className="flex items-center gap-1.5 text-xs text-[#6B655D]">
                      <CheckCircle2 className={`h-3 w-3 ${passwordChecks.length ? 'text-emerald-600' : 'text-[#C2BBB0]'}`} />
                      At least 8 characters
                    </li>
                    <li className="flex items-center gap-1.5 text-xs text-[#6B655D]">
                      <CheckCircle2 className={`h-3 w-3 ${passwordChecks.hasUpper ? 'text-emerald-600' : 'text-[#C2BBB0]'}`} />
                      One uppercase letter
                    </li>
                    <li className="flex items-center gap-1.5 text-xs text-[#6B655D]">
                      <CheckCircle2 className={`h-3 w-3 ${passwordChecks.hasNumber ? 'text-emerald-600' : 'text-[#C2BBB0]'}`} />
                      One number
                    </li>
                  </ul>
                )
              )}
            </div>

            {/* Confirm Password */}
            <div>
              <label htmlFor="confirmPassword" className="block text-sm font-medium text-[#2C2A29] mb-1.5">
                Confirm password
              </label>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-[#8C8275]" />
                <input
                  id="confirmPassword"
                  name="confirmPassword"
                  type={showConfirmPassword ? 'text' : 'password'}
                  autoComplete="new-password"
                  value={confirmPassword}
                  onChange={(e) => {
                    setConfirmPassword(e.target.value);
                    clearError('confirmPassword');
                  }}
                  placeholder="Re-enter your password"
                  aria-invalid={!!errors.confirmPassword}
                  aria-describedby={errors.confirmPassword ? 'confirmPassword-error' : undefined}
                  className={`w-full rounded border bg-[#F5F1E9] py-2.5 pl-10 pr-10 text-sm text-[#2C2A29] placeholder:text-[#8C8275] focus:outline-none focus:ring-2 focus:ring-[#594433] focus:border-transparent transition ${
                    errors.confirmPassword ? 'border-red-400' : 'border-[#C2BBB0]'
                  }`}
                />
                <button
                  type="button"
                  onClick={() => setShowConfirmPassword((s) => !s)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-[#8C8275] hover:text-[#2C2A29] transition"
                  aria-label={showConfirmPassword ? 'Hide password' : 'Show password'}
                  tabIndex={0}
                >
                  {showConfirmPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </button>
              </div>
              {errors.confirmPassword && (
                <p id="confirmPassword-error" className="mt-1.5 text-xs text-red-600">
                  {errors.confirmPassword}
                </p>
              )}
            </div>

            {/* Submit */}
            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full flex items-center justify-center gap-2 rounded bg-[#594433] py-2.5 text-sm font-medium text-white hover:bg-[#3D3733] focus:outline-none focus:ring-2 focus:ring-[#594433] focus:ring-offset-2 focus:ring-offset-[#EAE5DC] disabled:opacity-60 disabled:cursor-not-allowed transition"
            >
              {isSubmitting && <Loader2 className="h-4 w-4 animate-spin" />}
              {isSubmitting ? 'Creating account...' : 'Create account'}
            </button>
          </form>

          <p className="mt-8 text-center text-sm text-[#6B655D]">
            Already have an account?{' '}
            <Link to="/login" className="font-medium text-[#594433] hover:underline">
              Sign in
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
};

export default Register;
