import React, { useRef, useState } from 'react';
import { AlertCircle, FileText, Loader2, UploadCloud, RefreshCw, ExternalLink } from 'lucide-react';
import { Link } from 'react-router-dom';
import { Resume, isResumeReady } from '../services/resume';
import { ACCEPTED_RESUME_EXTENSIONS, ACCEPTED_RESUME_TYPES, MAX_RESUME_SIZE } from '../services/resume';

interface CvPickerProps {
  resumes: Resume[];
  active: Resume | null;
  loading?: boolean;
  uploading?: boolean;
  error?: string | null;
  onUpload: (file: File, replaceId?: string) => Promise<unknown>;
  onSelect?: (id: string) => void;
  title?: string;
  description?: string;
}

const statusMeta: Record<Resume['status'], { label: string; className: string }> = {
  ready: { label: 'Ready', className: 'bg-[var(--status-offer)]/15 text-[var(--status-offer)]' },
  processing: { label: 'Processing', className: 'bg-[var(--accent-gold)]/20 text-[var(--text-heading)]' },
  failed: { label: 'Failed', className: 'bg-[var(--status-rejected)]/15 text-[var(--status-rejected)]' },
};

export const CvPicker: React.FC<CvPickerProps> = ({
  resumes,
  active,
  loading = false,
  uploading = false,
  error = null,
  onUpload,
  onSelect,
  title = 'Your CV',
  description = 'Upload your master CV. SkillMatch parses it so the assistant and the CV tailor can work from your real experience.',
}) => {
  const inputRef = useRef<HTMLInputElement>(null);
  const [replaceMode, setReplaceMode] = useState(false);
  const [localError, setLocalError] = useState<string | null>(null);

  const shownError = localError || error;

  const beginPick = (replace: boolean) => {
    setLocalError(null);
    setReplaceMode(replace);
    inputRef.current?.click();
  };

  const handleFile = async (file?: File) => {
    if (!file || uploading) return;
    if (!ACCEPTED_RESUME_TYPES.includes(file.type)) {
      setLocalError('Choose a PDF, DOC, DOCX, or TXT file.');
      return;
    }
    if (file.size > MAX_RESUME_SIZE) {
      setLocalError('Your CV must be 5 MB or smaller.');
      return;
    }
    setLocalError(null);
    const replaceId = replaceMode && active ? active.id : undefined;
    await onUpload(file, replaceId);
    setReplaceMode(false);
  };

  const activeMeta = active ? statusMeta[active.status] : null;

  return (
    <section className="rounded-lg border border-[var(--border-hairline)] bg-[var(--bg-secondary)] p-4 shadow-sm sm:p-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="text-xs font-semibold uppercase tracking-[0.16em] text-[var(--accent-gold)]">{title}</p>
          {active ? (
            <div className="mt-1 flex min-w-0 flex-wrap items-center gap-2">
              <span className="grid h-8 w-8 shrink-0 place-items-center rounded-full bg-[var(--bg-card)] text-[var(--text-heading)]"><FileText size={15} /></span>
              <span className="min-w-0 truncate text-sm font-semibold text-[var(--text-heading)]">{active.name}</span>
              {activeMeta && <span className={`shrink-0 rounded-full px-2 py-0.5 text-xs font-semibold ${activeMeta.className}`}>{activeMeta.label}</span>}
            </div>
          ) : (
            <p className="mt-1 text-sm font-semibold text-[var(--text-heading)]">No CV uploaded yet</p>
          )}
        </div>

        {resumes.length > 1 && onSelect && (
          <label className="flex min-w-0 items-center gap-2 text-xs text-[var(--text-muted)]">
            <span className="shrink-0">Use</span>
            <select
              value={active?.id ?? ''}
              onChange={(event) => onSelect(event.target.value)}
              className="h-9 min-w-0 flex-1 rounded-md border border-[var(--border-hairline)] bg-[var(--bg-input)] px-2 text-sm text-[var(--text-heading)] outline-none focus:border-[var(--accent-gold)] sm:flex-none"
              aria-label="Select CV"
            >
              {resumes.map((resume) => (
                <option key={resume.id} value={resume.id}>{resume.name}</option>
              ))}
            </select>
          </label>
        )}
      </div>

      <p className="mt-2 text-xs leading-5 text-[var(--text-muted)]">{description}</p>

      {active && active.status === 'failed' && (
        <p role="alert" className="mt-3 flex items-start gap-2 rounded border border-[var(--status-rejected)]/30 bg-[var(--status-rejected)]/10 p-2.5 text-xs leading-5 text-[var(--text-heading)]">
          <AlertCircle className="mt-0.5 h-4 w-4 shrink-0 text-[var(--status-rejected)]" />
          <span>{active.failureReason || 'This CV could not be read. Try a different file (PDF, DOCX, or TXT works best).'}</span>
        </p>
      )}

      <input
        ref={inputRef}
        type="file"
        accept={ACCEPTED_RESUME_EXTENSIONS}
        className="sr-only"
        onChange={(event) => { void handleFile(event.target.files?.[0]); event.target.value = ''; }}
      />

      {!loading && resumes.length === 0 && (
        <button
          type="button"
          onClick={() => beginPick(false)}
          disabled={uploading}
          className="mt-4 flex w-full items-center justify-center gap-2 rounded-md border border-dashed border-[var(--border-dashed-gold)] bg-[var(--bg-primary)] px-4 py-3 text-sm font-semibold text-[var(--text-heading)] transition hover:border-[var(--accent-gold)] disabled:opacity-60"
        >
          {uploading ? <Loader2 className="h-4 w-4 animate-spin" /> : <UploadCloud size={17} />}
          {uploading ? 'Uploading…' : 'Upload your CV'}
        </button>
      )}

      {!loading && resumes.length > 0 && (
        <div className="mt-4 flex flex-wrap items-center gap-2">
          <button
            type="button"
            onClick={() => beginPick(true)}
            disabled={uploading}
            className="inline-flex items-center gap-2 rounded-md bg-[var(--btn-primary-bg)] px-3.5 py-2 text-sm font-semibold text-[var(--btn-primary-text)] transition-opacity disabled:opacity-60"
          >
            {uploading ? <Loader2 className="h-4 w-4 animate-spin" /> : <RefreshCw size={15} />}
            {uploading ? 'Uploading…' : 'Replace CV'}
          </button>
          <Link to="/resume" className="inline-flex items-center gap-1.5 text-sm font-semibold text-[var(--text-button-fill)]">
            Manage CVs <ExternalLink size={14} />
          </Link>
        </div>
      )}

      {loading && <div className="mt-4 flex items-center gap-2 text-sm text-[var(--text-muted)]"><Loader2 className="animate-spin" size={15} />Loading CVs…</div>}
      {shownError && !isResumeReady(active) && (
        <p role="alert" className="mt-3 flex items-center gap-2 text-xs text-[var(--status-rejected)]"><AlertCircle className="h-4 w-4 shrink-0" />{shownError}</p>
      )}
    </section>
  );
};

export default CvPicker;
