import { useCallback, useEffect, useRef, useState, type CSSProperties, type FC } from 'react';
import { AlertCircle, CheckCircle2, LayoutGrid, Loader2, Palette, RefreshCw, Type } from 'lucide-react';
import {
  CV_ACCENT_COLORS,
  CV_SPACING,
  CV_TYPOGRAPHY,
  CvDocument,
  CvThemeConfig,
  DEFAULT_CV_THEME,
  cvService,
  normalizeThemeConfig,
} from '../../services/cv';

type SaveState = 'idle' | 'saving' | 'saved' | 'error';

export interface ThemeCustomizerProps {
  /** Document whose `theme_config` should be loaded/saved. */
  cvId?: string;
  /** Controlled theme. When omitted the component manages its own state. */
  theme?: CvThemeConfig;
  /** Initial theme for uncontrolled usage. */
  defaultTheme?: CvThemeConfig;
  /** Fires immediately on every selection so a preview can update instantly. */
  onChange?: (theme: CvThemeConfig) => void;
  /** Fires after the API confirms a saved theme. */
  onSaved?: (document: CvDocument) => void;
  /** Load `theme_config` for `cvId` on mount. Defaults to true. */
  autoLoad?: boolean;
  /** Persist changes to `cvId` automatically. Defaults to true. */
  autoSave?: boolean;
  /** Render the bundled live preview. Defaults to true. */
  showPreview?: boolean;
  disabled?: boolean;
  className?: string;
}

const optionClass = (selected: boolean): string =>
  [
    'flex items-center gap-2 rounded-lg border px-3 py-2 text-left text-sm transition-colors',
    selected
      ? 'border-[var(--accent-gold)] bg-[var(--bg-card)] text-[var(--text-heading)] ring-1 ring-[var(--accent-gold)]'
      : 'border-[var(--border-hairline)] bg-[var(--bg-input)] text-[var(--text-body)] hover:border-[var(--accent-gold)]',
    'disabled:cursor-not-allowed disabled:opacity-60',
  ].join(' ');

export const ThemeCustomizer: FC<ThemeCustomizerProps> = ({
  cvId,
  theme,
  defaultTheme,
  onChange,
  onSaved,
  autoLoad = true,
  autoSave = true,
  showPreview = true,
  disabled = false,
  className = '',
}) => {
  const [internalTheme, setInternalTheme] = useState<CvThemeConfig>(() =>
    normalizeThemeConfig(defaultTheme ?? DEFAULT_CV_THEME),
  );
  const [loading, setLoading] = useState(Boolean(cvId && autoLoad && theme === undefined));
  const [loadError, setLoadError] = useState<string | null>(null);
  const [reloadKey, setReloadKey] = useState(0);
  const [saveState, setSaveState] = useState<SaveState>('idle');
  const [saveError, setSaveError] = useState<string | null>(null);

  const selected = normalizeThemeConfig(theme ?? internalTheme);
  const onChangeRef = useRef(onChange);
  const onSavedRef = useRef(onSaved);
  const mountedRef = useRef(true);
  const savingRef = useRef(false);
  const pendingThemeRef = useRef<CvThemeConfig | null>(null);

  useEffect(() => {
    onChangeRef.current = onChange;
    onSavedRef.current = onSaved;
  }, [onChange, onSaved]);

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
    };
  }, []);

  // Restore the saved theme when an existing CV is opened.
  useEffect(() => {
    if (!cvId || !autoLoad || theme !== undefined) {
      setLoading(false);
      return;
    }

    let cancelled = false;
    setLoading(true);
    setLoadError(null);

    void cvService
      .get(cvId)
      .then((document) => {
        if (cancelled) return;
        const restored = normalizeThemeConfig(document.themeConfig);
        setInternalTheme(restored);
        onChangeRef.current?.(restored);
      })
      .catch((error: unknown) => {
        if (cancelled) return;
        setLoadError(error instanceof Error ? error.message : 'The saved theme could not be loaded.');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [cvId, autoLoad, theme, reloadKey]);

  const flushTheme = useCallback(async () => {
    if (!cvId || savingRef.current) return;

    savingRef.current = true;
    try {
      while (pendingThemeRef.current) {
        const next = pendingThemeRef.current;
        pendingThemeRef.current = null;

        if (mountedRef.current) {
          setSaveState('saving');
          setSaveError(null);
        }

        try {
          const document = await cvService.updateTheme(cvId, next);
          if (mountedRef.current) {
            setSaveState('saved');
            onSavedRef.current?.(document);
          }
        } catch (error) {
          if (mountedRef.current) {
            setSaveState('error');
            setSaveError(error instanceof Error ? error.message : 'The theme could not be saved.');
          }
        }
      }
    } finally {
      savingRef.current = false;
    }
  }, [cvId]);

  const selectTheme = useCallback(
    (patch: Partial<CvThemeConfig>) => {
      const next = normalizeThemeConfig({ ...selected, ...patch });

      if (theme === undefined) setInternalTheme(next);
      onChangeRef.current?.(next);

      if (cvId && autoSave && !disabled) {
        pendingThemeRef.current = next;
        void flushTheme();
      }
    },
    [autoSave, cvId, disabled, flushTheme, selected, theme],
  );

  const selectedAccent = CV_ACCENT_COLORS.find((option) => option.id === selected.accentColor) ?? CV_ACCENT_COLORS[0];
  const selectedTypography = CV_TYPOGRAPHY.find((option) => option.id === selected.typography) ?? CV_TYPOGRAPHY[0];
  const selectedSpacing = CV_SPACING.find((option) => option.id === selected.spacing) ?? CV_SPACING[1];

  const previewStyle: CSSProperties = {
    '--cv-accent': selectedAccent.value,
    '--cv-space': selectedSpacing.value,
    fontFamily: selectedTypography.fontFamily,
  } as CSSProperties;

  return (
    <section
      className={`rounded-xl border border-[var(--border-hairline)] bg-[var(--bg-secondary)] p-5 ${className}`}
      aria-label="CV theme customizer"
    >
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="font-serif text-2xl font-semibold text-[var(--text-heading)]">Style your CV</h2>
          <p className="mt-1 text-sm text-[var(--text-muted)]">
            Personalise colour, type, and density. Your content stays exactly the same.
          </p>
        </div>
        <div className="flex min-h-5 items-center gap-2 text-xs" aria-live="polite">
          {loading ? (
            <span role="status" className="inline-flex items-center gap-1.5 text-[var(--text-muted)]">
              <Loader2 className="animate-spin" size={14} /> Loading saved theme…
            </span>
          ) : null}
          {saveState === 'saving' ? (
            <span role="status" className="inline-flex items-center gap-1.5 text-[var(--text-muted)]">
              <Loader2 className="animate-spin" size={14} /> Saving theme…
            </span>
          ) : null}
          {saveState === 'saved' ? (
            <span className="inline-flex items-center gap-1.5 text-[var(--status-offer)]">
              <CheckCircle2 size={14} /> Theme saved
            </span>
          ) : null}
          {saveState === 'error' ? (
            <span role="alert" className="inline-flex items-center gap-1.5 text-[var(--status-rejected)]">
              <AlertCircle size={14} /> {saveError}
            </span>
          ) : null}
        </div>
      </header>

      {loadError ? (
        <div
          role="alert"
          className="mt-4 flex flex-wrap items-center justify-between gap-3 rounded-lg border border-[var(--status-rejected)] bg-[var(--status-rejected)]/10 p-3 text-sm text-[var(--text-heading)]"
        >
          <span className="inline-flex items-center gap-2">
            <AlertCircle size={16} /> {loadError}
          </span>
          <button
            type="button"
            onClick={() => setReloadKey((key) => key + 1)}
            className="inline-flex items-center gap-1.5 rounded border border-[var(--status-rejected)] px-2.5 py-1 text-xs font-semibold"
          >
            <RefreshCw size={13} /> Retry
          </button>
        </div>
      ) : null}

      <div className={`mt-5 grid gap-6 ${showPreview ? 'lg:grid-cols-[minmax(0,1fr)_minmax(260px,0.9fr)]' : ''}`}>
        <div className="space-y-6">
          <div>
            <p className="flex items-center gap-2 text-xs font-bold uppercase tracking-[0.18em] text-[var(--text-muted)]">
              <Palette size={14} /> Accent colour
            </p>
            <div className="mt-2.5 grid grid-cols-2 gap-2 sm:grid-cols-3" role="group" aria-label="Accent colour">
              {CV_ACCENT_COLORS.map((option) => {
                const active = selected.accentColor === option.id;
                return (
                  <button
                    key={option.id}
                    type="button"
                    aria-pressed={active}
                    disabled={disabled || loading}
                    onClick={() => selectTheme({ accentColor: option.id })}
                    className={optionClass(active)}
                  >
                    <span
                      className="h-4 w-4 shrink-0 rounded-full border border-black/10"
                      style={{ backgroundColor: option.value }}
                      aria-hidden="true"
                    />
                    {option.label}
                  </button>
                );
              })}
            </div>
          </div>

          <div>
            <p className="flex items-center gap-2 text-xs font-bold uppercase tracking-[0.18em] text-[var(--text-muted)]">
              <Type size={14} /> Typography
            </p>
            <div className="mt-2.5 grid gap-2 sm:grid-cols-3" role="group" aria-label="Typography">
              {CV_TYPOGRAPHY.map((option) => {
                const active = selected.typography === option.id;
                return (
                  <button
                    key={option.id}
                    type="button"
                    aria-label={`${option.label} (${option.family})`}
                    aria-pressed={active}
                    disabled={disabled || loading}
                    onClick={() => selectTheme({ typography: option.id })}
                    className={optionClass(active)}
                  >
                    <span className="min-w-0">
                      <span className="block font-semibold" style={{ fontFamily: option.fontFamily }}>
                        {option.label}
                      </span>
                      <span className="block truncate text-xs opacity-70">{option.family}</span>
                    </span>
                  </button>
                );
              })}
            </div>
          </div>

          <div>
            <p className="flex items-center gap-2 text-xs font-bold uppercase tracking-[0.18em] text-[var(--text-muted)]">
              <LayoutGrid size={14} /> Spacing
            </p>
            <div className="mt-2.5 grid grid-cols-3 gap-2" role="group" aria-label="Spacing">
              {CV_SPACING.map((option) => {
                const active = selected.spacing === option.id;
                return (
                  <button
                    key={option.id}
                    type="button"
                    aria-label={option.label}
                    aria-pressed={active}
                    disabled={disabled || loading}
                    onClick={() => selectTheme({ spacing: option.id })}
                    className={optionClass(active)}
                  >
                    <span className="min-w-0">
                      <span className="block font-semibold">{option.label}</span>
                      <span className="block text-xs opacity-70">{option.value}</span>
                    </span>
                  </button>
                );
              })}
            </div>
          </div>
        </div>

        {showPreview ? (
          <div className="lg:sticky lg:top-4 lg:self-start">
            <p className="mb-2 text-xs font-bold uppercase tracking-[0.18em] text-[var(--text-muted)]">Live preview</p>
            <article
              data-testid="cv-theme-preview"
              className="overflow-hidden rounded-lg border border-[var(--border-hairline)] bg-white text-[#3A2A1C] shadow-sm"
              style={previewStyle}
            >
              <div style={{ borderTop: `5px solid ${selectedAccent.value}`, padding: selectedSpacing.value }}>
                <p className="text-[10px] font-semibold uppercase tracking-[0.22em]" style={{ color: selectedAccent.value }}>
                  Curriculum vitae
                </p>
                <h3 className="mt-1 text-xl font-bold" style={{ color: selectedAccent.value }}>
                  Alex Morgan
                </h3>
                <p className="text-[11px] text-[#6B5B4B]">Product Designer · Nairobi, Kenya</p>

                <div style={{ marginTop: selectedSpacing.value, display: 'grid', gap: selectedSpacing.value }}>
                  <section>
                    <h4
                      className="text-[10px] font-bold uppercase tracking-[0.16em]"
                      style={{ color: selectedAccent.value }}
                    >
                      Experience
                    </h4>
                    <p className="mt-1 text-[11px] font-semibold">Senior Product Designer · Acme Corp</p>
                    <p className="text-[10px] leading-relaxed text-[#6B5B4B]">
                      Led the redesign of the enterprise dashboard and built the shared design system.
                    </p>
                  </section>

                  <section>
                    <h4
                      className="text-[10px] font-bold uppercase tracking-[0.16em]"
                      style={{ color: selectedAccent.value }}
                    >
                      Skills
                    </h4>
                    <p className="mt-1 text-[10px] leading-relaxed text-[#6B5B4B]">
                      Design systems · User research · Prototyping · Figma
                    </p>
                  </section>
                </div>
              </div>
            </article>
          </div>
        ) : null}
      </div>
    </section>
  );
};

export default ThemeCustomizer;
