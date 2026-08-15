import React from 'react';
import { Navbar } from './Navbar';

export const AppShell: React.FC<{ children: React.ReactNode }> = ({ children }) => <div className="flex min-h-screen flex-col bg-[var(--bg-primary)]"><Navbar /><main className="mx-auto w-full max-w-7xl flex-1 px-4 py-8 sm:px-8 sm:py-10">{children}</main><footer className="border-t border-[var(--border-hairline)] px-5 py-5 text-xs text-[var(--text-muted)]"><div className="mx-auto flex max-w-7xl flex-col gap-2 text-center sm:flex-row sm:items-center sm:justify-between sm:text-left"><span>© {new Date().getFullYear()} Skill-match. Professional Tailoring.</span><span>Privacy&nbsp;&nbsp; Terms&nbsp;&nbsp; Help</span></div></footer></div>;
