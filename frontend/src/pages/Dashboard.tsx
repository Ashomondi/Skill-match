import React from 'react';
import { useAuth } from '../hooks/useAuth';
import { Navbar } from '../components/Navbar';

export const Dashboard: React.FC = () => {
  const { user } = useAuth();

  return (
    <div className="min-h-screen bg-[#EAE5DC] font-mono text-[#2C2A29] flex flex-col">
      <Navbar />

      <main className="flex-1 max-w-5xl w-full mx-auto p-6">
        <div className="mb-8">
          <div className="text-xs text-[#6B655D] mb-1">Memory: healthy</div>
          <h1 className="text-2xl font-bold">Welcome back, {user?.fullName || 'Professional'}.</h1>
          <p className="text-sm text-[#6B655D] mt-1">
            Your career, with a memory that never forgets what worked.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          <div className="border border-[#C2BBB0] bg-[#F5F1E9] p-5 rounded">
            <h3 className="text-xs text-[#6B655D] uppercase tracking-wider mb-2">Active Tailored CVs</h3>
            <div className="text-3xl font-bold">12</div>
          </div>
          <div className="border border-[#C2BBB0] bg-[#F5F1E9] p-5 rounded">
            <h3 className="text-xs text-[#6B655D] uppercase tracking-wider mb-2">Applications Tracked</h3>
            <div className="text-3xl font-bold">6</div>
          </div>
          <div className="border border-[#C2BBB0] bg-[#F5F1E9] p-5 rounded">
            <h3 className="text-xs text-[#6B655D] uppercase tracking-wider mb-2">Match Efficiency</h3>
            <p className="text-sm mt-1">Gets smarter with every application.</p>
          </div>
        </div>

        <div className="border border-[#C2BBB0] bg-[#F5F1E9] p-6 rounded">
          <h2 className="text-lg font-bold mb-4 border-b border-[#C2BBB0] pb-2">Quick Actions</h2>
          <div className="flex flex-wrap gap-4">
            <a
              href="/resume"
              className="bg-[#E3DCD1] border border-[#8C8275] px-4 py-2 rounded text-sm hover:bg-[#D8D0C3]"
            >
              Upload & Tailor New CV
            </a>
            <a
              href="/jobs"
              className="bg-[#E3DCD1] border border-[#8C8275] px-4 py-2 rounded text-sm hover:bg-[#D8D0C3]"
            >
              Discover Matching Roles
            </a>
            <a
              href="/applications"
              className="bg-[#E3DCD1] border border-[#8C8275] px-4 py-2 rounded text-sm hover:bg-[#D8D0C3]"
            >
              View Application Status
            </a>
          </div>
        </div>
      </main>

      <footer className="border-t border-[#C2BBB0] py-4 px-6 text-xs text-[#6B655D] flex justify-between">
        <div>2026 Skill-match. Professional Tailoring.</div>
        <div className="flex gap-4">
          <a href="#" className="hover:underline">Privacy</a>
          <a href="#" className="hover:underline">Terms</a>
          <a href="#" className="hover:underline">Help</a>
        </div>
      </footer>
    </div>
  );
};
