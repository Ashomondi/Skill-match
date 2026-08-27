import React from 'react';
import { Navigate, Route, Routes, useLocation } from 'react-router-dom';
import { Dashboard } from './pages/Dashboard';
import { Login } from './pages/Login';
import { Register } from './pages/Register';
import { ResumePage } from './pages/Resume';
import { Profile } from './pages/Profile';
import { Applications } from './pages/Applications';
import { JobDetail, Jobs } from './pages/Jobs';
import { Tailor } from './pages/Tailor';
import { useAuth } from './hooks/useAuth';
import { Chat } from './pages/Chat';
import { SavedJobs } from './pages/SavedJobs';
import { Loader2 } from 'lucide-react';

const RouteLoading = () => <div className="flex min-h-screen items-center justify-center gap-2 bg-[var(--bg-primary)] text-sm text-[var(--text-muted)]"><Loader2 className="animate-spin" size={18} />Loading...</div>;
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => { const { isAuthenticated, loading } = useAuth(); if (loading) return <RouteLoading />; return isAuthenticated ? <>{children}</> : <Navigate to="/login" replace />; };
const PublicOnlyRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => { const { isAuthenticated, loading } = useAuth(); if (loading) return <RouteLoading />; return isAuthenticated ? <Navigate to="/dashboard" replace /> : <>{children}</>; };

export const AppRoutes: React.FC = () => {
  const location = useLocation();
  const routeKey = `${location.pathname}${location.search}`;

  return <div key={routeKey} className="route-transition" aria-live="polite"><Routes location={location}>
    <Route path="/login" element={<PublicOnlyRoute><Login /></PublicOnlyRoute>} />
    <Route path="/register" element={<PublicOnlyRoute><Register /></PublicOnlyRoute>} />
    <Route path="/dashboard" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
    <Route path="/chat" element={<ProtectedRoute><Chat /></ProtectedRoute>} />
    <Route path="/resume" element={<ProtectedRoute><ResumePage /></ProtectedRoute>} />
    <Route path="/cv-tailor" element={<ProtectedRoute><Navigate to="/discover" replace /></ProtectedRoute>} />
    <Route path="/profile" element={<ProtectedRoute><Profile /></ProtectedRoute>} />
    <Route path="/discover" element={<ProtectedRoute><Jobs /></ProtectedRoute>} />
    <Route path="/discover/:jobId" element={<ProtectedRoute><JobDetail /></ProtectedRoute>} />
    <Route path="/discover/:jobId/tailor" element={<ProtectedRoute><Tailor /></ProtectedRoute>} />
    <Route path="/applications" element={<ProtectedRoute><Applications /></ProtectedRoute>} />
    <Route path="/saved-jobs" element={<ProtectedRoute><SavedJobs /></ProtectedRoute>} />
    <Route path="/" element={<Navigate to="/dashboard" replace />} />
    <Route path="*" element={<Navigate to="/dashboard" replace />} />
  </Routes></div>;
};
