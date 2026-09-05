import React, { useCallback, useEffect, useRef, useState } from 'react';
import { AlertCircle, CheckCircle2 } from 'lucide-react';
import { Navbar } from '../components/Navbar';
import { ResumeList } from '../components/ResumeList';
import { ResumeUploader } from '../components/ResumeUploader';
import { Resume, ResumeStatus, resumeService } from '../services/resume';

export const ResumePage: React.FC = () => {
  const [resumes, setResumes] = useState<Resume[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isUploading, setIsUploading] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const replaceTarget = useRef<Resume | null>(null);

  const loadResumes = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      setResumes(await resumeService.list());
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : 'Unable to load your resumes.');
    } finally {
      setIsLoading(false);
    }
  }, []);
  useEffect(() => { void loadResumes(); }, [loadResumes]);

  const upload = async (file: File) => {
    setError(null);
    setMessage(null);
    setIsUploading(true);
    try {
      const uploaded = await resumeService.upload(file, { replaceId: replaceTarget.current?.id });
      setResumes((current) => replaceTarget.current
        ? [uploaded, ...current.filter((resume) => resume.id !== replaceTarget.current?.id)]
        : [uploaded, ...current]);
      setMessage(replaceTarget.current ? 'Your resume was replaced successfully.' : 'Your resume is uploading and will be ready shortly.');
      replaceTarget.current = null;
    } catch (uploadError) {
      setError(uploadError instanceof Error ? uploadError.message : 'Your resume could not be uploaded.');
    } finally {
      setIsUploading(false);
    }
  };

  const remove = async (resume: Resume) => {
    if (!window.confirm(`Delete ${resume.name}? This cannot be undone.`)) return;
    setDeletingId(resume.id);
    setError(null);
    try {
      await resumeService.deleteResume(resume.id);
      setResumes((current) => current.filter((item) => item.id !== resume.id));
      setMessage('Resume deleted.');
    } catch (deleteError) {
      setError(deleteError instanceof Error ? deleteError.message : 'Your resume could not be deleted.');
    } finally {
      setDeletingId(null);
    }
  };

  const updateStatus = (resume: Resume, status: ResumeStatus, failureReason?: string) => {
    setResumes((current) =>
      current.map((item) =>
        item.id === resume.id
          ? { ...item, status, ...(failureReason !== undefined ? { failureReason } : {}) }
          : item
      )
    );
  };

  return <div className="min-h-screen bg-[#F6F0E6] text-[#3A2A1C]"><Navbar /><main className="mx-auto w-full max-w-5xl px-6 py-10 sm:py-14"><header className="mb-8 max-w-2xl"><p className="text-xs font-semibold uppercase tracking-[0.16em] text-[#B08D57]">CV tailor</p><h1 className="mt-2 font-serif text-[38px] font-semibold tracking-[-0.03em]">Your master profile starts here.</h1><p className="mt-3 text-[15px] leading-6 text-[#8A7B6B]">Upload the resume that best represents your experience. You can replace it whenever your story evolves.</p></header>{message ? <div className="mb-5 flex items-center gap-2 rounded border border-[#7A8B6F] bg-[#7A8B6F]/10 p-3 text-sm text-[#3A2A1C]"><CheckCircle2 className="h-4 w-4 text-[#7A8B6F]" />{message}</div> : null}{error ? <div role="alert" className="mb-5 flex items-center gap-2 rounded border border-[#B5573C] bg-[#B5573C]/10 p-3 text-sm text-[#3A2A1C]"><AlertCircle className="h-4 w-4 text-[#B5573C]" />{error}</div> : null}<ResumeUploader onUpload={upload} isUploading={isUploading} />{isLoading ? <div className="mt-8 rounded-lg border border-[#D8C9B2] bg-[#F6F0E6] p-6 text-sm text-[#8A7B6B]">Loading your resumes…</div> : <ResumeList resumes={resumes} isDeletingId={deletingId} onDelete={remove} onStatusChange={updateStatus} onReplace={(resume) => { replaceTarget.current = resume; setMessage(`Choose a new file to replace ${resume.name}.`); window.scrollTo({ top: 0, behavior: 'smooth' }); }} />}</main></div>;
};

export default ResumePage;
