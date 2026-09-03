import React, { useEffect, useState } from 'react';
import { AlertCircle, CheckCircle2, Loader2, Plus, Save, X } from 'lucide-react';
import { AppShell } from '../components/AppShell';
import { useAuth } from '../hooks/useAuth';
import { useProfile } from '../hooks/useProfile';

export const Profile: React.FC = () => {
  const { user } = useAuth();
  const { profile, loading, saving, error, saveError, load, save, clearSaveError } = useProfile();

  const [bio, setBio] = useState('');
  const [skills, setSkills] = useState<string[]>([]);
  const [experience, setExperience] = useState<string[]>([]);
  const [resumeUrl, setResumeUrl] = useState('');
  const [skillDraft, setSkillDraft] = useState('');
  const [experienceDraft, setExperienceDraft] = useState('');
  const [creating, setCreating] = useState(false);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    if (profile) {
      setBio(profile.bio);
      setSkills(profile.skills);
      setExperience(profile.experience);
      setResumeUrl(profile.resumeUrl);
    }
  }, [profile]);

  const initials = (user?.fullName || 'You').split(' ').filter(Boolean).map((part) => part[0]).join('').slice(0, 2).toUpperCase();

  const addSkill = () => {
    const value = skillDraft.trim();
    if (!value) return;
    if (!skills.some((skill) => skill.toLowerCase() === value.toLowerCase())) setSkills([...skills, value]);
    setSkillDraft('');
  };

  const removeSkill = (skill: string) => setSkills(skills.filter((item) => item !== skill));

  const addExperience = () => {
    const value = experienceDraft.trim();
    if (!value) return;
    setExperience([...experience, value]);
    setExperienceDraft('');
  };

  const removeExperience = (index: number) => setExperience(experience.filter((_, i) => i !== index));

  const updateExperience = (index: number, value: string) => setExperience(experience.map((item, i) => (i === index ? value : item)));

  const startCreating = () => {
    setBio('');
    setSkills([]);
    setExperience([]);
    setResumeUrl('');
    setCreating(true);
  };

  const handleSave = async () => {
    setSaved(false);
    clearSaveError();
    try {
      await save({ bio, skills, experience, resumeUrl });
      setSaved(true);
      setCreating(false);
    } catch {
      /* error surfaced via saveError */
    }
  };

  const showForm = creating || profile !== null;

  return (
    <AppShell>
      <h1 className="font-serif text-4xl font-bold text-[var(--text-heading)]">Your Master Profile</h1>
      <p className="mt-2 text-[var(--text-body)]">This master profile powers every tailored version we generate.</p>
      <div className="mt-6 rounded-lg border border-[var(--accent-gold)] bg-[var(--bg-card)] p-4 text-sm text-[var(--text-insight)]">Changes here update every future tailored CV automatically. Keep your core experience comprehensive.</div>

      {loading ? (
        <div className="mt-8 flex items-center gap-2 text-sm text-[var(--text-muted)]"><Loader2 className="animate-spin" size={18} />Loading your profile…</div>
      ) : error ? (
        <div role="alert" className="mt-8 flex flex-col gap-3 rounded-md border border-[var(--status-rejected)] bg-[var(--status-rejected)]/10 p-4 text-sm sm:flex-row sm:items-center sm:justify-between"><span className="flex items-center gap-2"><AlertCircle size={18} />{error}</span><button type="button" onClick={() => void load()} className="inline-flex items-center gap-2 rounded border border-[var(--text-button-fill)] px-3 py-2 text-sm font-semibold"><Loader2 size={15} />Try again</button></div>
      ) : !showForm ? (
        <div className="mt-8 rounded-xl border border-[var(--border-hairline)] bg-[var(--bg-secondary)] p-12 text-center">
          <h2 className="font-serif text-2xl font-bold text-[var(--text-heading)]">No profile yet</h2>
          <p className="mx-auto mt-2 max-w-md text-sm text-[var(--text-muted)]">Tell us about your skills and experience to unlock personalized job recommendations.</p>
          <button type="button" onClick={startCreating} className="mt-6 inline-flex items-center gap-2 rounded bg-[var(--btn-primary-bg)] px-4 py-2 text-sm font-semibold text-[var(--btn-primary-text)]"><Plus size={16} />Create your profile</button>
        </div>
      ) : (
        <div className="mt-7 space-y-6">
          <section className="rounded-xl border border-[var(--border-hairline)] bg-[var(--bg-secondary)] p-6">
            <div className="flex gap-4">
              <div className="grid h-14 w-14 place-items-center rounded-full bg-[var(--text-button-fill)] font-serif text-xl text-[var(--btn-primary-text)]">{initials}</div>
              <div>
                <h2 className="font-serif text-2xl font-bold text-[var(--text-heading)]">{user?.fullName || 'Your profile'}</h2>
                <p className="text-sm text-[var(--text-muted)]">{user?.email || ''}</p>
              </div>
            </div>
            <label className="mt-5 block text-sm">
              <span className="mb-1 block font-semibold text-[var(--text-heading)]">Bio</span>
              <textarea value={bio} onChange={(event) => setBio(event.target.value)} rows={4} placeholder="A short summary of who you are and what you do." className="w-full rounded-lg border border-[var(--border-hairline)] bg-[var(--bg-input)] p-3 text-sm leading-6 outline-none focus:border-[var(--accent-gold)]" />
            </label>
          </section>

          <section className="rounded-xl border border-[var(--border-hairline)] bg-[var(--bg-secondary)] p-6">
            <h2 className="font-serif text-2xl font-bold text-[var(--text-heading)]">Work History</h2>
            <p className="mt-1 text-xs font-bold tracking-[.2em] text-[var(--text-muted)]">EXPERIENCE</p>
            <div className="mt-4 space-y-3">
              {experience.map((entry, index) => (
                <div key={index} className="flex items-start gap-2">
                  <textarea value={entry} onChange={(event) => updateExperience(index, event.target.value)} rows={2} className="min-w-0 flex-1 rounded-lg border border-[var(--border-hairline)] bg-[var(--bg-input)] p-3 text-sm leading-6 outline-none focus:border-[var(--accent-gold)]" />
                  <button type="button" onClick={() => removeExperience(index)} aria-label="Remove experience" className="mt-2 shrink-0 rounded p-1 text-[var(--text-muted)] hover:text-[var(--status-rejected)]"><X size={16} /></button>
                </div>
              ))}
              {experience.length === 0 && <p className="text-sm text-[var(--text-muted)]">No experience added yet.</p>}
            </div>
            <div className="mt-3 flex gap-2">
              <input value={experienceDraft} onChange={(event) => setExperienceDraft(event.target.value)} onKeyDown={(event) => { if (event.key === 'Enter') { event.preventDefault(); addExperience(); } }} placeholder="e.g. Senior Product Designer at Acme Corp — 2020 to present" className="h-11 min-w-0 flex-1 rounded-lg border border-[var(--border-hairline)] bg-[var(--bg-input)] px-3 text-sm outline-none focus:border-[var(--accent-gold)]" />
              <button type="button" onClick={addExperience} className="inline-flex items-center gap-1 rounded border border-dashed border-[var(--border-dashed-gold)] px-4 text-sm text-[var(--text-button-fill)]"><Plus size={16} />Add</button>
            </div>
          </section>

          <section className="rounded-xl border border-[var(--border-hairline)] bg-[var(--bg-secondary)] p-6">
            <h2 className="font-serif text-2xl font-bold text-[var(--text-heading)]">Skills</h2>
            <p className="mt-1 text-xs font-bold tracking-[.2em] text-[var(--text-muted)]">WHAT YOU BRING</p>
            <div className="mt-4 flex flex-wrap gap-2">
              {skills.map((skill) => (
                <span className="flex items-center gap-2 rounded-full bg-[var(--bg-chip)] px-3 py-2 text-sm" key={skill}>{skill}<button type="button" onClick={() => removeSkill(skill)} aria-label={`Remove ${skill}`} className="text-[var(--text-muted)] hover:text-[var(--status-rejected)]"><X size={14} /></button></span>
              ))}
              {skills.length === 0 && <span className="text-sm text-[var(--text-muted)]">No skills added yet.</span>}
            </div>
            <div className="mt-3 flex gap-2">
              <input value={skillDraft} onChange={(event) => setSkillDraft(event.target.value)} onKeyDown={(event) => { if (event.key === 'Enter') { event.preventDefault(); addSkill(); } }} placeholder="e.g. Figma" className="h-11 min-w-0 flex-1 rounded-lg border border-[var(--border-hairline)] bg-[var(--bg-input)] px-3 text-sm outline-none focus:border-[var(--accent-gold)]" />
              <button type="button" onClick={addSkill} className="inline-flex items-center gap-1 rounded border border-dashed border-[var(--border-dashed-gold)] px-4 text-sm text-[var(--text-button-fill)]"><Plus size={16} />Add</button>
            </div>
          </section>

          <section className="rounded-xl border border-[var(--border-hairline)] bg-[var(--bg-secondary)] p-6">
            <h2 className="font-serif text-2xl font-bold text-[var(--text-heading)]">Resume</h2>
            <label className="mt-4 block text-sm">
              <span className="mb-1 block font-semibold text-[var(--text-heading)]">Resume link</span>
              <input value={resumeUrl} onChange={(event) => setResumeUrl(event.target.value)} placeholder="https://example.com/your-resume.pdf" className="h-11 w-full rounded-lg border border-[var(--border-hairline)] bg-[var(--bg-input)] px-3 text-sm outline-none focus:border-[var(--accent-gold)]" />
            </label>
          </section>

          {saveError && <div role="alert" className="flex items-center gap-2 rounded border border-[var(--status-rejected)] bg-[var(--status-rejected)]/10 p-3 text-sm"><AlertCircle size={16} />{saveError}</div>}
          {saved && <div role="status" className="flex items-center gap-2 rounded border border-[var(--status-offer)] bg-[var(--status-offer)]/10 p-3 text-sm"><CheckCircle2 size={16} />Profile saved.</div>}

          <div className="flex items-center gap-3">
            <button type="button" onClick={() => void handleSave()} disabled={saving} className="inline-flex items-center gap-2 rounded bg-[var(--btn-primary-bg)] px-5 py-2 text-sm font-semibold text-[var(--btn-primary-text)] disabled:opacity-50">{saving ? <Loader2 className="animate-spin" size={16} /> : <Save size={16} />}{saving ? 'Saving…' : 'Save profile'}</button>
            {!saving && saved && <span className="text-sm text-[var(--status-offer)]">All changes saved.</span>}
          </div>
        </div>
      )}
    </AppShell>
  );
};
