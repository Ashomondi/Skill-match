// frontend/src/hooks/useUpload.ts

import { useState, useCallback } from 'react';
import { Resume, resumeService } from '../services/resume';

const ALLOWED_TYPES = [
  'application/pdf',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  'text/plain',
];
const MAX_FILE_SIZE = 5 * 1024 * 1024; // 5MB

export function useUpload() {
  const [progress, setProgress] = useState<number>(0);
  const [isUploading, setIsUploading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [successData, setSuccessData] = useState<Resume | null>(null);

  const validateFile = (file: File): string | null => {
    if (!ALLOWED_TYPES.includes(file.type)) {
      return 'Invalid file type. Please upload a PDF, DOCX, or TXT format.';
    }
    if (file.size > MAX_FILE_SIZE) {
      return 'File is too large. Maximum allowed size is 5MB.';
    }
    return null;
  };

  const upload = useCallback(async (file: File, replaceId?: string) => {
    setError(null);
    setSuccessData(null);
    setProgress(0);

    const validationError = validateFile(file);
    if (validationError) {
      setError(validationError);
      return;
    }

    setIsUploading(true);

    try {
      const response = await resumeService.upload(file, {
        replaceId,
        onProgress: setProgress,
      });
      setSuccessData(response);
      setProgress(100);
    } catch (err: any) {
      setError(err.message || 'An unexpected error occurred during upload.');
    } finally {
      setIsUploading(false);
    }
  }, []);

  const resetUpload = useCallback(() => {
    setProgress(0);
    setIsUploading(false);
    setError(null);
    setSuccessData(null);
  }, []);

  return {
    upload,
    resetUpload,
    progress,
    isUploading,
    error,
    successData,
  };
}
