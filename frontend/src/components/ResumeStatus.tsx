import React, { useEffect, useState } from 'react';
import { CheckCircle2, Clock, Loader2, XCircle } from 'lucide-react';
import { resumeService, ResumeStatus as ResumeStatusType } from '../services/resume';

const POLL_INTERVAL_MS = 3000;

const styles: Record<ResumeStatusType, string> = {
  uploaded: 'bg-[#EDE1CE] text-[#5C3A21]',
  parsing: 'bg-[#B08D57]/15 text-[#795e32]',
  parsed: 'bg-[#7A8B6F]/15 text-[#52624a]',
  failed: 'bg-[#B5573C]/10 text-[#B5573C]',
};

interface ResumeStatusProps {
  resumeId: string;
  status: ResumeStatusType;
  failureReason?: string;
  onStatusChange?: (status: ResumeStatusType, failureReason?: string) => void;
}

/**
 * Displays a resume's lifecycle state (uploaded | parsing | parsed | failed)
 * and polls the backend while in-flight (uploaded or parsing) until a terminal
 * state (parsed or failed) is reached.
 */
export const ResumeStatus: React.FC<ResumeStatusProps> = ({
  resumeId,
  status,
  failureReason,
  onStatusChange,
}) => {
  const [pollError, setPollError] = useState(false);

  const isPending = status === 'uploaded' || status === 'parsing';

  useEffect(() => {
    if (!isPending || !resumeId) return;

    let cancelled = false;
    const poll = async () => {
      try {
        const latest = await resumeService.get(resumeId);
        if (cancelled) return;
        setPollError(false);
        if (latest.status !== status || latest.failureReason !== failureReason) {
          onStatusChange?.(latest.status, latest.failureReason);
        }
      } catch {
        if (!cancelled) setPollError(true);
      }
    };

    const interval = window.setInterval(() => {
      void poll();
    }, POLL_INTERVAL_MS);

    return () => {
      cancelled = true;
      window.clearInterval(interval);
    };
  }, [resumeId, status, failureReason, isPending, onStatusChange]);

  const Icon =
    status === 'parsing'
      ? Loader2
      : status === 'parsed'
      ? CheckCircle2
      : status === 'failed'
      ? XCircle
      : Clock;

  const titleText =
    status === 'failed'
      ? failureReason || 'Processing failed — try uploading again'
      : status === 'parsing'
      ? 'Parsing resume text and extracting skills…'
      : status === 'uploaded'
      ? 'Resume uploaded, queued for parsing…'
      : 'Resume ready and parsed';

  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-semibold capitalize ${styles[status]}`}
      role="status"
      aria-label={`Resume status: ${status}`}
      title={titleText}
    >
      <Icon className={status === 'parsing' ? 'animate-spin' : undefined} size={13} />
      {status}
      {pollError && <span className="font-normal normal-case">(refresh failed)</span>}
    </span>
  );
};

export default ResumeStatus;
