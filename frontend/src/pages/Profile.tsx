import React, { useEffect, useState } from 'react';
import { Check, ChevronDown, Edit3, Loader2, Plus, Save, Trash2, X } from 'lucide-react';
import { AppShell } from '../components/AppShell';
import { userService, UserProfile, WorkExperience, Education } from '../services/user';

export const Profile: React.FC = () => {
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [savedSuccess, setSavedSuccess] = useState(false);
  const [isEditingInfo, setIsEditingInfo] = useState(false);

  const [openWorkIndex, setOpenWorkIndex] = useState<number>(0);
  const [newSkillInput, setNewSkillInput] = useState('');

  const [profile, setProfile] = useState<UserProfile>({
    fullName: '',
    email: '',
    phone: '',
    location: '',
    summary: '',
    skills: [],
    workHistory: [],
    education: [],
  });

  useEffect(() => {
    let mounted = true;
    userService.getProfile().then((p) => {
      if (mounted) {
        setProfile(p);
        setLoading(false);
      }
    });
    return () => {
      mounted = false;
    };
  }, []);

  const handleSave = async () => {
    setSaving(true);
    try {
      await userService.updateProfile(profile);
      setSavedSuccess(true);
      setIsEditingInfo(false);
      setTimeout(() => setSavedSuccess(false), 3000);
    } catch (err) {
      console.error('Failed to save profile:', err);
    } finally {
      setSaving(false);
    }
  };

  const handleAddSkill = (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    const trimmed = newSkillInput.trim();
    if (trimmed && !profile.skills.includes(trimmed)) {
      setProfile((prev) => ({ ...prev, skills: [...prev.skills, trimmed] }));
      setNewSkillInput('');
    }
  };

  const handleRemoveSkill = (skill: string) => {
    setProfile((prev) => ({
      ...prev,
      skills: prev.skills.filter((s) => s !== skill),
    }));
  };

  const handleAddWork = () => {
    const newWork: WorkExperience = {
      id: `work-${Date.now()}`,
      title: 'New Role Title',
      company: 'Company Name',
      duration: 'Year – Present',
      bullets: ['Describe your core achievements in this role.'],
    };
    setProfile((prev) => ({
      ...prev,
      workHistory: [newWork, ...prev.workHistory],
    }));
    setOpenWorkIndex(0);
  };

  const handleRemoveWork = (id: string) => {
    setProfile((prev) => ({
      ...prev,
      workHistory: prev.workHistory.filter((w) => w.id !== id),
    }));
  };

  const handleAddBullet = (workId: string) => {
    setProfile((prev) => ({
      ...prev,
      workHistory: prev.workHistory.map((w) =>
        w.id === workId ? { ...w, bullets: [...w.bullets, 'New achievement bullet point.'] } : w
      ),
    }));
  };

  const handleUpdateBullet = (workId: string, bIndex: number, text: string) => {
    setProfile((prev) => ({
      ...prev,
      workHistory: prev.workHistory.map((w) => {
        if (w.id !== workId) return w;
        const bullets = [...w.bullets];
        bullets[bIndex] = text;
        return { ...w, bullets };
      }),
    }));
  };

  const handleRemoveBullet = (workId: string, bIndex: number) => {
    setProfile((prev) => ({
      ...prev,
      workHistory: prev.workHistory.map((w) => {
        if (w.id !== workId) return w;
        return { ...w, bullets: w.bullets.filter((_, idx) => idx !== bIndex) };
      }),
    }));
  };

  const handleAddEducation = () => {
    const newEdu: Education = {
      id: `edu-${Date.now()}`,
      degree: 'B.S. in Computer Science',
      institution: 'University Name',
      years: '2016 – 2020',
    };
    setProfile((prev) => ({
      ...prev,
      education: [...prev.education, newEdu],
    }));
  };

  const handleRemoveEducation = (id: string) => {
    setProfile((prev) => ({
      ...prev,
      education: prev.education.filter((e) => e.id !== id),
    }));
  };

  const initials = profile.fullName
    ? profile.fullName
        .split(' ')
        .map((p) => p[0])
        .slice(0, 2)
        .join('')
        .toUpperCase()
    : 'SM';

  if (loading) {
    return (
      <AppShell>
        <div className="flex min-h-[300px] items-center justify-center gap-2 text-sm text-[var(--text-muted)]">
          <Loader2 className="animate-spin" size={20} />
          Loading your master profile...
        </div>
      </AppShell>
    );
  }

  return (
    <AppShell>
      <div className="space-y-8">
        {/* Header & Save Action */}
        <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div>
            <h1 className="font-serif text-3xl font-bold text-[var(--text-heading)] md:text-4xl">
              Your Master Profile
            </h1>
            <p className="mt-1 text-sm text-[var(--text-body)]">
              This master profile powers your AI conversational agent and personalized CV tailoring.
            </p>
          </div>

          <div className="flex items-center gap-3">
            <button
              onClick={handleSave}
              disabled={saving}
              className="inline-flex items-center gap-2 rounded-lg bg-[var(--btn-primary-bg)] px-4 py-2 text-sm font-medium text-[var(--btn-primary-text)] transition hover:opacity-90 disabled:opacity-50"
            >
              {saving ? (
                <>
                  <Loader2 size={16} className="animate-spin" />
                  Saving...
                </>
              ) : savedSuccess ? (
                <>
                  <Check size={16} />
                  Changes Saved!
                </>
              ) : (
                <>
                  <Save size={16} />
                  Save Profile
                </>
              )}
            </button>
          </div>
        </div>

        {savedSuccess && (
          <div className="flex items-center gap-2 rounded-lg border border-emerald-500/30 bg-emerald-500/10 px-4 py-3 text-sm text-emerald-400">
            <Check size={16} />
            Profile successfully saved and synchronized.
          </div>
        )}

        {/* Notice Banner */}
        <div className="rounded-lg border border-[var(--accent-gold)] bg-[var(--bg-card)] p-4 text-sm text-[var(--text-insight)]">
          Changes made here update every future tailored CV and AI recommendation automatically.
        </div>

        {/* Personal & Summary Card */}
        <section className="rounded-xl border border-[var(--border-hairline)] bg-[var(--bg-secondary)] p-6 shadow-sm">
          <div className="flex flex-col gap-5 md:flex-row md:items-start md:justify-between">
            <div className="flex gap-4">
              <div className="grid h-16 w-16 flex-shrink-0 place-items-center rounded-full bg-[var(--text-button-fill)] font-serif text-xl font-bold text-[var(--btn-primary-text)]">
                {initials}
              </div>

              <div className="flex-1 space-y-1">
                {isEditingInfo ? (
                  <div className="space-y-3">
                    <input
                      type="text"
                      value={profile.fullName}
                      onChange={(e) => setProfile({ ...profile, fullName: e.target.value })}
                      placeholder="Full Name"
                      className="w-full rounded border border-[var(--border-hairline)] bg-[var(--bg-card)] px-3 py-1.5 text-sm text-[var(--text-heading)] focus:outline-none focus:ring-1 focus:ring-[var(--accent-gold)]"
                    />
                    <div className="grid gap-2 sm:grid-cols-3">
                      <input
                        type="email"
                        value={profile.email}
                        onChange={(e) => setProfile({ ...profile, email: e.target.value })}
                        placeholder="Email"
                        className="rounded border border-[var(--border-hairline)] bg-[var(--bg-card)] px-3 py-1.5 text-xs text-[var(--text-heading)]"
                      />
                      <input
                        type="text"
                        value={profile.phone}
                        onChange={(e) => setProfile({ ...profile, phone: e.target.value })}
                        placeholder="Phone"
                        className="rounded border border-[var(--border-hairline)] bg-[var(--bg-card)] px-3 py-1.5 text-xs text-[var(--text-heading)]"
                      />
                      <input
                        type="text"
                        value={profile.location}
                        onChange={(e) => setProfile({ ...profile, location: e.target.value })}
                        placeholder="Location"
                        className="rounded border border-[var(--border-hairline)] bg-[var(--bg-card)] px-3 py-1.5 text-xs text-[var(--text-heading)]"
                      />
                    </div>
                  </div>
                ) : (
                  <>
                    <h2 className="font-serif text-2xl font-bold text-[var(--text-heading)]">
                      {profile.fullName || 'Your Name'}
                    </h2>
                    <p className="text-sm text-[var(--text-muted)]">
                      {[profile.email, profile.phone, profile.location].filter(Boolean).join(' • ')}
                    </p>
                  </>
                )}
              </div>
            </div>

            <button
              onClick={() => setIsEditingInfo(!isEditingInfo)}
              className="inline-flex items-center gap-1 text-xs font-medium text-[var(--text-button-fill)] hover:underline"
            >
              <Edit3 size={14} />
              {isEditingInfo ? 'Done editing info' : 'Edit details'}
            </button>
          </div>

          <div className="mt-6 border-t border-[var(--border-hairline)] pt-5">
            <h3 className="text-xs font-bold uppercase tracking-wider text-[var(--text-muted)]">
              Professional Summary
            </h3>
            {isEditingInfo ? (
              <textarea
                value={profile.summary}
                onChange={(e) => setProfile({ ...profile, summary: e.target.value })}
                rows={4}
                className="mt-2.5 w-full rounded-lg border border-[var(--border-hairline)] bg-[var(--bg-card)] p-3 text-sm leading-relaxed text-[var(--text-heading)] focus:outline-none focus:ring-1 focus:ring-[var(--accent-gold)]"
                placeholder="Write a concise overview of your background, specialties, and career highlights..."
              />
            ) : (
              <p className="mt-2.5 text-sm leading-relaxed text-[var(--text-body)]">
                {profile.summary || 'Add a professional summary to help personalize your job matches.'}
              </p>
            )}
          </div>
        </section>

        {/* Work History Section */}
        <section className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="font-serif text-2xl font-bold text-[var(--text-heading)]">
              Work History
            </h2>
            <button
              onClick={handleAddWork}
              className="inline-flex items-center gap-1.5 rounded-lg border border-dashed border-[var(--border-dashed-gold)] px-3 py-1.5 text-xs font-medium text-[var(--text-button-fill)] hover:bg-[var(--bg-card)]"
            >
              <Plus size={14} />
              Add role
            </button>
          </div>

          <div className="space-y-3">
            {profile.workHistory.map((work, idx) => (
              <article
                key={work.id}
                className="rounded-xl border border-[var(--border-hairline)] bg-[var(--bg-secondary)] shadow-sm"
              >
                <div className="flex items-center justify-between p-5">
                  <button
                    onClick={() => setOpenWorkIndex(openWorkIndex === idx ? -1 : idx)}
                    className="flex flex-1 items-center justify-between text-left"
                  >
                    <div>
                      <h3 className="font-semibold text-[var(--text-heading)]">{work.title}</h3>
                      <p className="mt-0.5 text-xs text-[var(--text-muted)]">
                        {work.company} • {work.duration}
                      </p>
                    </div>
                    <ChevronDown
                      className={`transition-transform duration-200 ${
                        openWorkIndex === idx ? 'rotate-180' : ''
                      }`}
                      size={18}
                    />
                  </button>
                  <button
                    onClick={() => handleRemoveWork(work.id)}
                    className="ml-4 text-[var(--text-muted)] hover:text-red-400"
                    title="Delete role"
                  >
                    <Trash2 size={16} />
                  </button>
                </div>

                {openWorkIndex === idx && (
                  <div className="border-t border-[var(--border-hairline)] bg-[var(--bg-card)]/40 p-5 space-y-4">
                    <div className="grid gap-3 sm:grid-cols-3">
                      <div>
                        <label className="text-xs font-medium text-[var(--text-muted)]">Role Title</label>
                        <input
                          type="text"
                          value={work.title}
                          onChange={(e) => {
                            const updated = [...profile.workHistory];
                            updated[idx].title = e.target.value;
                            setProfile({ ...profile, workHistory: updated });
                          }}
                          className="mt-1 w-full rounded border border-[var(--border-hairline)] bg-[var(--bg-card)] px-3 py-1.5 text-xs text-[var(--text-heading)]"
                        />
                      </div>
                      <div>
                        <label className="text-xs font-medium text-[var(--text-muted)]">Company</label>
                        <input
                          type="text"
                          value={work.company}
                          onChange={(e) => {
                            const updated = [...profile.workHistory];
                            updated[idx].company = e.target.value;
                            setProfile({ ...profile, workHistory: updated });
                          }}
                          className="mt-1 w-full rounded border border-[var(--border-hairline)] bg-[var(--bg-card)] px-3 py-1.5 text-xs text-[var(--text-heading)]"
                        />
                      </div>
                      <div>
                        <label className="text-xs font-medium text-[var(--text-muted)]">Duration</label>
                        <input
                          type="text"
                          value={work.duration}
                          onChange={(e) => {
                            const updated = [...profile.workHistory];
                            updated[idx].duration = e.target.value;
                            setProfile({ ...profile, workHistory: updated });
                          }}
                          className="mt-1 w-full rounded border border-[var(--border-hairline)] bg-[var(--bg-card)] px-3 py-1.5 text-xs text-[var(--text-heading)]"
                        />
                      </div>
                    </div>

                    <div>
                      <label className="text-xs font-medium text-[var(--text-muted)]">
                        Key Responsibilities & Achievements
                      </label>
                      <div className="mt-2 space-y-2">
                        {work.bullets.map((bullet, bIdx) => (
                          <div key={bIdx} className="flex items-center gap-2">
                            <input
                              type="text"
                              value={bullet}
                              onChange={(e) => handleUpdateBullet(work.id, bIdx, e.target.value)}
                              className="flex-1 rounded border border-[var(--border-hairline)] bg-[var(--bg-card)] px-3 py-1.5 text-xs text-[var(--text-heading)]"
                            />
                            <button
                              onClick={() => handleRemoveBullet(work.id, bIdx)}
                              className="text-[var(--text-muted)] hover:text-red-400"
                            >
                              <X size={14} />
                            </button>
                          </div>
                        ))}
                      </div>
                      <button
                        onClick={() => handleAddBullet(work.id)}
                        className="mt-2 text-xs text-[var(--text-button-fill)] hover:underline"
                      >
                        + Add bullet point
                      </button>
                    </div>
                  </div>
                )}
              </article>
            ))}
          </div>
        </section>

        {/* Skills Section */}
        <section className="space-y-4">
          <h2 className="font-serif text-2xl font-bold text-[var(--text-heading)]">
            Core Competencies & Skills
          </h2>

          <div className="rounded-xl border border-[var(--border-hairline)] bg-[var(--bg-secondary)] p-5 space-y-4">
            <form onSubmit={handleAddSkill} className="flex gap-2">
              <input
                type="text"
                value={newSkillInput}
                onChange={(e) => setNewSkillInput(e.target.value)}
                placeholder="Add a new skill (e.g. Go, PostgreSQL, Docker)..."
                className="flex-1 rounded-lg border border-[var(--border-hairline)] bg-[var(--bg-card)] px-3 py-2 text-xs text-[var(--text-heading)] focus:outline-none focus:ring-1 focus:ring-[var(--accent-gold)]"
              />
              <button
                type="submit"
                className="rounded-lg bg-[var(--btn-primary-bg)] px-4 py-2 text-xs font-medium text-[var(--btn-primary-text)] hover:opacity-90"
              >
                Add Skill
              </button>
            </form>

            <div className="flex flex-wrap gap-2 pt-2">
              {profile.skills.map((skill) => (
                <span
                  key={skill}
                  className="inline-flex items-center gap-1.5 rounded-full bg-[var(--bg-chip)] px-3 py-1.5 text-xs font-medium text-[var(--text-heading)]"
                >
                  {skill}
                  <button
                    onClick={() => handleRemoveSkill(skill)}
                    className="text-[var(--text-muted)] hover:text-red-400"
                    title={`Remove ${skill}`}
                  >
                    <X size={13} />
                  </button>
                </span>
              ))}
            </div>
          </div>
        </section>

        {/* Education Section */}
        <section className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="font-serif text-2xl font-bold text-[var(--text-heading)]">
              Education & Certifications
            </h2>
            <button
              onClick={handleAddEducation}
              className="inline-flex items-center gap-1.5 rounded-lg border border-dashed border-[var(--border-dashed-gold)] px-3 py-1.5 text-xs font-medium text-[var(--text-button-fill)] hover:bg-[var(--bg-card)]"
            >
              <Plus size={14} />
              Add education
            </button>
          </div>

          <div className="space-y-3">
            {profile.education.map((edu, idx) => (
              <div
                key={edu.id}
                className="flex items-center justify-between rounded-xl border border-[var(--border-hairline)] bg-[var(--bg-secondary)] p-5 shadow-sm"
              >
                <div className="space-y-1">
                  <input
                    type="text"
                    value={edu.degree}
                    onChange={(e) => {
                      const updated = [...profile.education];
                      updated[idx].degree = e.target.value;
                      setProfile({ ...profile, education: updated });
                    }}
                    className="rounded border border-[var(--border-hairline)] bg-[var(--bg-card)] px-2.5 py-1 text-sm font-semibold text-[var(--text-heading)]"
                  />
                  <div className="flex gap-2">
                    <input
                      type="text"
                      value={edu.institution}
                      onChange={(e) => {
                        const updated = [...profile.education];
                        updated[idx].institution = e.target.value;
                        setProfile({ ...profile, education: updated });
                      }}
                      className="rounded border border-[var(--border-hairline)] bg-[var(--bg-card)] px-2.5 py-1 text-xs text-[var(--text-muted)]"
                    />
                    <input
                      type="text"
                      value={edu.years}
                      onChange={(e) => {
                        const updated = [...profile.education];
                        updated[idx].years = e.target.value;
                        setProfile({ ...profile, education: updated });
                      }}
                      className="rounded border border-[var(--border-hairline)] bg-[var(--bg-card)] px-2.5 py-1 text-xs text-[var(--text-muted)] w-28"
                    />
                  </div>
                </div>

                <button
                  onClick={() => handleRemoveEducation(edu.id)}
                  className="text-[var(--text-muted)] hover:text-red-400"
                  title="Remove entry"
                >
                  <Trash2 size={16} />
                </button>
              </div>
            ))}
          </div>
        </section>
      </div>
    </AppShell>
  );
};
