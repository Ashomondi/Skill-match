import React from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';
import { Dashboard } from './pages/Dashboard';
import { Login } from './pages/Login';
import { Register } from './pages/Register';
import { ResumePage } from './pages/Resume';
import { useAuth } from './hooks/useAuth';

const RouteLoading = () => <div className="flex min-h-screen items-center justify-center bg-[#F6F0E6] text-sm text-[#8A7B6B]">Loading...</div>;
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => { const { isAuthenticated, loading } = useAuth(); if (loading) return <RouteLoading />; return isAuthenticated ? <>{children}</> : <Navigate to="/login" replace />; };
const PublicOnlyRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => { const { isAuthenticated, loading } = useAuth(); if (loading) return <RouteLoading />; return isAuthenticated ? <Navigate to="/dashboard" replace /> : <>{children}</>; };

export const AppRoutes: React.FC = () => <Routes>
  <Route path="/login" element={<PublicOnlyRoute><Login /></PublicOnlyRoute>} />
  <Route path="/register" element={<PublicOnlyRoute><Register /></PublicOnlyRoute>} />
  <Route path="/dashboard" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
  <Route path="/resume" element={<ProtectedRoute><ResumePage /></ProtectedRoute>} />
  <Route path="/" element={<Navigate to="/dashboard" replace />} />
  <Route path="*" element={<Navigate to="/dashboard" replace />} />
</Routes>;
