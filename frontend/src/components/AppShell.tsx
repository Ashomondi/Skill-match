import React from 'react';
import { Navbar } from './Navbar';
export const AppShell: React.FC<{children: React.ReactNode}> = ({ children }) => <div className="flex min-h-screen flex-col bg-[var(--bg-primary)]"><Navbar /><main className="mx-auto w-full max-w-6xl flex-1 px-5 py-10 sm:px-8">{children}</main><footer className="border-t border-[var(--border-hairline)] px-6 py-5 text-center text-xs text-[var(--text-muted)]">© 2024 Skill-match. Professional Tailoring. <span className="ml-4">Privacy&nbsp;&nbsp; Terms&nbsp;&nbsp; Help</span></footer></div>;
