import React from 'react';
import { FileText, Loader2, RefreshCw, Trash2 } from 'lucide-react';

export type Resume = { id: string; name: string; uploadedAt: string; status: 'processing' | 'ready' | 'failed'; size?: number };
type ResumeListProps = { resumes: Resume[]; onReplace: (resume: Resume) => void; onDelete: (resume: Resume) => void; isDeletingId?: string | null };
const statusStyles = { ready: 'bg-[#7A8B6F]/15 text-[#52624a]', processing: 'bg-[#B08D57]/15 text-[#795e32]', failed: 'bg-[#B5573C]/10 text-[#B5573C]' };
const formatDate = (date: string) => new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' }).format(new Date(date));
const formatSize = (size?: number) => size ? `${(size / 1024 / 1024).toFixed(1)} MB` : 'Resume file';

export const ResumeList: React.FC<ResumeListProps> = ({ resumes, onReplace, onDelete, isDeletingId }) => <section className="mt-8">
  <div className="mb-4 flex items-end justify-between"><div><p className="text-xs font-semibold uppercase tracking-[0.16em] text-[#B08D57]">Your library</p><h2 className="mt-1 font-serif text-[25px] font-semibold text-[#3A2A1C]">Uploaded resumes</h2></div><span className="text-sm text-[#8A7B6B]">{resumes.length} {resumes.length === 1 ? 'file' : 'files'}</span></div>
  {resumes.length === 0 ? <div className="rounded-lg border border-[#D8C9B2] bg-[#F6F0E6] px-6 py-10 text-center text-sm text-[#8A7B6B]">Your uploaded resumes will appear here.</div> : <div className="overflow-hidden rounded-lg border border-[#D8C9B2] bg-[#F6F0E6] shadow-[0px_2px_8px_rgba(92,58,33,0.08)]">{resumes.map((resume) => <article key={resume.id} className="flex flex-col gap-4 border-b border-[#D8C9B2] p-5 last:border-b-0 sm:flex-row sm:items-center sm:justify-between"><div className="flex min-w-0 items-center gap-3"><div className="flex h-10 w-10 shrink-0 items-center justify-center rounded bg-[#EFE6D6] text-[#5C3A21]"><FileText className="h-5 w-5" /></div><div className="min-w-0"><p className="truncate text-sm font-semibold text-[#3A2A1C]">{resume.name}</p><p className="mt-0.5 text-xs text-[#8A7B6B]">Uploaded {formatDate(resume.uploadedAt)} · {formatSize(resume.size)}</p></div></div><div className="flex items-center gap-3"><span className={`rounded-full px-2.5 py-1 text-xs font-semibold capitalize ${statusStyles[resume.status]}`}>{resume.status}</span><button type="button" onClick={() => onReplace(resume)} className="inline-flex items-center gap-1.5 text-sm font-semibold text-[#5C3A21] hover:underline"><RefreshCw className="h-4 w-4" />Replace</button><button type="button" onClick={() => onDelete(resume)} disabled={isDeletingId === resume.id} className="text-[#8A7B6B] transition hover:text-[#B5573C] disabled:opacity-50" aria-label={`Delete ${resume.name}`}>{isDeletingId === resume.id ? <Loader2 className="h-4 w-4 animate-spin" /> : <Trash2 className="h-4 w-4" />}</button></div></article>)}</div>}
</section>;

export default ResumeList;
