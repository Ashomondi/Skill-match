import React, { useEffect, useState } from 'react';
import { Menu, X } from 'lucide-react';
import { Link, NavLink, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

const links = [['/dashboard','Dashboard'],['/chat','Chat'],['/cv-tailor','CV Tailor'],['/discover','Discover'],['/saved-jobs','Saved Jobs'],['/applications','Applications']];
const navClass = ({ isActive }: { isActive: boolean }) => `block rounded-md px-3 py-2 text-sm font-medium ${isActive ? 'bg-[var(--bg-card)] text-[var(--text-heading)] md:bg-transparent md:text-[var(--text-heading)]' : 'text-[var(--text-body)] hover:bg-[var(--bg-card)]/60 hover:text-[var(--text-heading)] md:hover:bg-transparent'}`;

export const Navbar: React.FC = () => {
  const { user, isAuthenticated, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [open, setOpen] = useState(false);
  useEffect(() => { setOpen(false); }, [location.pathname]);
  const handleLogout = () => { logout(); navigate('/login'); };

  return <header className="sticky top-0 z-30 border-b border-[var(--border-hairline)] bg-[var(--bg-nav)]/95 backdrop-blur">
    <div className="mx-auto flex min-h-16 max-w-7xl items-center justify-between gap-4 px-5 sm:px-8">
      <Link to="/dashboard" className="shrink-0 font-serif text-xl font-bold text-[var(--text-heading)]">Skill-match</Link>
      {isAuthenticated && <nav className="hidden items-center md:flex">{links.map(([to,label]) => <NavLink key={to} to={to} className={navClass}>{label}</NavLink>)}</nav>}
      <div className="flex items-center gap-2 text-sm">{isAuthenticated ? <><span className="hidden max-w-40 truncate text-[var(--text-muted)] lg:block">Hi, {user?.fullName || 'User'}</span><button type="button" onClick={handleLogout} className="hidden rounded-md border border-[var(--text-button-fill)] px-3 py-2 font-medium text-[var(--text-button-fill)] sm:block">Log out</button><button type="button" onClick={() => setOpen((value) => !value)} className="grid h-10 w-10 place-items-center rounded-md text-[var(--text-heading)] md:hidden" aria-label={open ? 'Close navigation' : 'Open navigation'} aria-expanded={open}>{open ? <X size={21} /> : <Menu size={21} />}</button></> : <><Link to="/login" className="px-2 py-2 text-[var(--text-heading)]">Log in</Link><Link to="/register" className="rounded-md bg-[var(--btn-primary-bg)] px-4 py-2 font-semibold text-[var(--btn-primary-text)]">Sign up</Link></>}</div>
    </div>
    {isAuthenticated && open && <nav className="border-t border-[var(--border-hairline)] px-4 py-3 md:hidden" aria-label="Mobile navigation">{links.map(([to,label]) => <NavLink key={to} to={to} className={navClass}>{label}</NavLink>)}<button type="button" onClick={handleLogout} className="mt-2 w-full rounded-md border border-[var(--text-button-fill)] px-3 py-2 text-left text-sm font-medium text-[var(--text-button-fill)] sm:hidden">Log out</button></nav>}
  </header>;
};
