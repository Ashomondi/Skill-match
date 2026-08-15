import React from 'react';
import { Link, NavLink, useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

export const Navbar: React.FC = () => {
  const { user, isAuthenticated, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <header className="flex items-center justify-between border-b border-[var(--border-hairline)] bg-[var(--bg-nav)] px-5 py-4 sm:px-8">
      <div className="flex items-center gap-6">
        <Link to="/dashboard" className="font-serif text-xl font-bold text-[var(--text-heading)]">
          Skill-match
        </Link>
        {isAuthenticated && (
          <nav className="hidden gap-5 text-sm md:flex">
            {[['/dashboard','Dashboard'],['/chat','Chat'],['/cv-tailor','CV Tailor'],['/discover','Discover'],['/saved-jobs','Saved Jobs'],['/applications','Applications']].map(([to,label]) => <NavLink key={to} to={to} className={({isActive}) => `border-b-2 pb-1 ${isActive ? 'border-[var(--accent-gold)] text-[var(--text-heading)]' : 'border-transparent text-[var(--text-body)] hover:text-[var(--text-heading)]'}`}>{label}</NavLink>)}
          </nav>
        )}
      </div>

      <div className="flex items-center gap-3 text-sm">
        {isAuthenticated ? (
          <>
            <span className="hidden text-[var(--text-muted)] sm:block">Hi, {user?.fullName || 'User'}</span>
            <button
              onClick={handleLogout}
              className="rounded border border-[var(--text-button-fill)] px-3 py-1.5 text-[var(--text-button-fill)]"
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
