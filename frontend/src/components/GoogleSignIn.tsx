import { useEffect, useRef, useState } from 'react';
import { Loader2 } from 'lucide-react';
import {
  googleClientId,
  isGoogleSignInConfigured,
  loadGoogleIdentityServices,
} from '../services/google';

type GoogleSignInProps = {
  onCredential: (credential: string) => void | Promise<void>;
  busy?: boolean;
};

// GoogleSignIn renders Google's official "Continue with Google" button using
// Google Identity Services, and hands the resulting ID token to the caller.
// It renders nothing when VITE_GOOGLE_CLIENT_ID is not configured.
export const GoogleSignIn: React.FC<GoogleSignInProps> = ({ onCredential, busy = false }) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const callbackRef = useRef(onCredential);
  const [error, setError] = useState<string | null>(null);

  callbackRef.current = onCredential;

  useEffect(() => {
    if (!isGoogleSignInConfigured) return;

    let cancelled = false;

    loadGoogleIdentityServices()
      .then(() => {
        const accounts = window.google?.accounts?.id;
        const container = containerRef.current;

        if (cancelled || !accounts || !container) return;

        accounts.initialize({
          client_id: googleClientId,
          ux_mode: 'popup',
          callback: (response) => {
            if (!response.credential) {
              setError('Google did not return a credential. Please try again.');
              return;
            }
            setError(null);
            void callbackRef.current(response.credential);
          },
        });

        container.innerHTML = '';
        accounts.renderButton(container, {
          type: 'standard',
          theme: 'outline',
          size: 'large',
          text: 'continue_with',
          shape: 'rectangular',
          logo_alignment: 'left',
          width: Math.min(container.clientWidth || 400, 400),
        });
      })
      .catch(() => {
        if (!cancelled) {
          setError('Google sign-in could not be loaded. Check your connection and try again.');
        }
      });

    return () => {
      cancelled = true;
    };
  }, []);

  if (!isGoogleSignInConfigured) return null;

  return (
    <>
      <div className="my-6 flex items-center gap-3">
        <span className="h-px flex-1 bg-[#D8C9B2]" />
        <span className="bg-[#F6F0E6] px-2 text-xs text-[#8A7B6B]">or</span>
        <span className="h-px flex-1 bg-[#D8C9B2]" />
      </div>
      <div className="relative flex min-h-[44px] justify-center">
        <div
          ref={containerRef}
          className={`flex w-full justify-center ${busy ? 'pointer-events-none opacity-60' : ''}`}
        />
        {busy ? (
          <span className="absolute inset-0 flex items-center justify-center">
            <Loader2 className="h-4 w-4 animate-spin text-[#3A2A1C]" />
          </span>
        ) : null}
      </div>
      {error ? (
        <p role="alert" className="mt-2 text-center text-xs text-[#B5573C]">
          {error}
        </p>
      ) : null}
    </>
  );
};

export default GoogleSignIn;
