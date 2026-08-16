import React, { DragEvent, useRef, useState } from 'react';
import { AlertCircle, FileUp, Loader2, UploadCloud } from 'lucide-react';

const ACCEPTED_TYPES = ['application/pdf', 'application/msword', 'application/vnd.openxmlformats-officedocument.wordprocessingml.document', 'text/plain'];
const MAX_FILE_SIZE = 5 * 1024 * 1024;

type ResumeUploaderProps = {
  onUpload: (file: File) => Promise<void>;
  isUploading?: boolean;
};

export const ResumeUploader: React.FC<ResumeUploaderProps> = ({ onUpload, isUploading = false }) => {
  const inputRef = useRef<HTMLInputElement>(null);
  const [isDragging, setIsDragging] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const selectFile = async (file?: File) => {
    if (!file || isUploading) return;
    if (!ACCEPTED_TYPES.includes(file.type)) { setError('Choose a PDF, DOC, DOCX, or TXT file.'); return; }
    if (file.size > MAX_FILE_SIZE) { setError('Your resume must be 5 MB or smaller.'); return; }
    setError(null);
    try { await onUpload(file); } catch (uploadError) { setError(uploadError instanceof Error ? uploadError.message : 'Your resume could not be uploaded. Please try again.'); }
  };
  const drop = (event: DragEvent<HTMLDivElement>) => { event.preventDefault(); setIsDragging(false); void selectFile(event.dataTransfer.files[0]); };

  return <section className="rounded-lg border border-[#D8C9B2] bg-[#F6F0E6] p-6 shadow-[0px_2px_8px_rgba(92,58,33,0.08)] sm:p-8">
    <div className="mb-6"><p className="text-xs font-semibold uppercase tracking-[0.16em] text-[#B08D57]">Add a resume</p><h2 className="mt-2 font-serif text-[28px] font-semibold tracking-[-0.02em] text-[#3A2A1C]">Upload your master CV.</h2><p className="mt-2 text-sm leading-6 text-[#8A7B6B]">We’ll use it as the foundation for every tailored application.</p></div>
    <div onDragOver={(event) => { event.preventDefault(); setIsDragging(true); }} onDragLeave={() => setIsDragging(false)} onDrop={drop} className={`rounded-lg border border-dashed px-6 py-10 text-center transition sm:px-10 ${isDragging ? 'border-[#5C3A21] bg-[#E3D7C4]/60' : 'border-[#B08D57] bg-[#EFE6D6]/50'}`}>
      <input ref={inputRef} type="file" accept=".pdf,.doc,.docx,.txt,application/pdf,application/msword,application/vnd.openxmlformats-officedocument.wordprocessingml.document,text/plain" className="sr-only" onChange={(event) => { void selectFile(event.target.files?.[0]); event.target.value = ''; }} />
      <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-[#E3D7C4] text-[#5C3A21]"><FileUp className="h-6 w-6" strokeWidth={1.7} /></div>
      <h3 className="mt-4 text-[15px] font-semibold text-[#3A2A1C]">Drag and drop your CV here</h3><p className="mt-1 text-sm text-[#8A7B6B]">or choose a file from your computer</p>
      <button type="button" disabled={isUploading} onClick={() => inputRef.current?.click()} className="mt-5 inline-flex h-10 items-center gap-2 rounded bg-[#5C3A21] px-4 text-sm font-semibold text-[#F6F0E6] transition hover:bg-[#4A2F1A] disabled:cursor-not-allowed disabled:opacity-60">{isUploading ? <Loader2 className="h-4 w-4 animate-spin" /> : <UploadCloud className="h-4 w-4" />}{isUploading ? 'Uploading…' : 'Browse files'}</button>
      <p className="mt-4 text-xs text-[#8A7B6B]">PDF, DOC, DOCX, or TXT · Maximum file size 5 MB</p>
    </div>
    {error ? <p role="alert" className="mt-3 flex items-center gap-2 text-sm text-[#B5573C]"><AlertCircle className="h-4 w-4" />{error}</p> : null}
  </section>;
};

export default ResumeUploader;
