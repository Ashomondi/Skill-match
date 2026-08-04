import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

export const Navbar: React.FC = () => {
  const { user, isAuthenticated, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <header className="border-b border-[#C2BBB0] bg-[#EAE5DC] px-6 py-4 flex justify-between items-center font-mono">
      <div className="flex items-center gap-8">
        <Link to="/" className="font-bold text-lg text-[#2C2A29]">
          Skill-match
        </Link>
        {isAuthenticated && (
          <nav className="flex gap-6 text-sm">
            <Link to="/dashboard" className="text-[#2C2A29] hover:underline">Dashboard</Link>
            <Link to="/resume" className="text-[#2C2A29] hover:underline">CV Tailor</Link>
            <Link to="/jobs" className="text-[#2C2A29] hover:underline">Discover</Link>
            <Link to="/applications" className="text-[#2C2A29] hover:underline">Applications</Link>
            <Link to="/history" className="text-[#2C2A29] hover:underline">History</Link>
          </nav>
        )}
      </div>

      <div className="flex items-center gap-4 text-sm">
        {isAuthenticated ? (
          <>
            <span className="text-[#6B655D]">Hi, {user?.fullName || 'User'}</span>
            <button
              onClick={handleLogout}
              className="bg-[#E3DCD1] border border-[#8C8275] px-3 py-1.5 rounded cursor-pointer hover:bg-[#D8D0C3]"
            >
              Log out
            </button>
          </>
        ) : (
          <>
            <Link to="/login" className="text-[#2C2A29] hover:underline">Log in</Link>
            <Link
              to="/register"
              className="bg-[#594433] text-white px-4 py-1.5 rounded cursor-pointer hover:opacity-90"
            >
              Sign up
            </Link>
          </>
        )}
      </div>
    </header>
  );
};
