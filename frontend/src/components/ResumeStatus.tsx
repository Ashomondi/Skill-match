import React, { useEffect, useState } from 'react';
import { CheckCircle2, Loader2, XCircle } from 'lucide-react';
import { resumeService, ResumeStatus as ResumeStatusType } from '../services/resume';

const POLL_INTERVAL_MS = 3000;

const styles: Record<ResumeStatusType, string> = {
  processing: 'bg-[#B08D57]/15 text-[#795e32]',
  active: 'bg-[#7A8B6F]/15 text-[#52624a]',
  failed: 'bg-[#B5573C]/10 text-[#B5573C]',
};

interface ResumeStatusProps {
  resumeId: string;
  status: ResumeStatusType;
  onStatusChange?: (status: ResumeStatusType) => void;
}

/**
 * Displays a resume's processing state and keeps it fresh by polling the
 * backend while the resume is still being processed — no full page reload
 * required. Stops polling once a terminal state (active/failed) is reached.
 */
export const ResumeStatus: React.FC<ResumeStatusProps> = ({ resumeId, status, onStatusChange }) => {
  const [pollError, setPollError] = useState(false);

  useEffect(() => {
    if (status !== 'processing' || !resumeId) return;

    let cancelled = false;
    const poll = async () => {
      try {
        const latest = await resumeService.get(resumeId);
        if (cancelled) return;
        setPollError(false);
        if (latest.status !== status) onStatusChange?.(latest.status);
      } catch {
        if (!cancelled) setPollError(true);
      }
    };

    const interval = window.setInterval(() => { void poll(); }, POLL_INTERVAL_MS);
    return () => {
      cancelled = true;
      window.clearInterval(interval);
    };
  }, [resumeId, status, onStatusChange]);

  const Icon = status === 'processing' ? Loader2 : status === 'active' ? CheckCircle2 : XCircle;

  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-semibold capitalize ${styles[status]}`}
      role="status"
      aria-label={`Resume status: ${status}`}
      title={status === 'failed' ? 'Processing failed — try uploading again' : undefined}
    >
      <Icon className={status === 'processing' ? 'animate-spin' : undefined} size={13} />
      {status}
      {pollError && <span className="font-normal normal-case">(refresh failed)</span>}
    </span>
  );
};

export default ResumeStatus;
