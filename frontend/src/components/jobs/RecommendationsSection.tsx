import React from 'react';
import { AlertCircle, RefreshCw, Sparkles } from 'lucide-react';
import { useRecommendations } from '../../hooks/useRecommendations';
import { RecommendationCard } from './RecommendationCard';

export const RecommendationsSection: React.FC = () => {
  const { recommendations, loading, error, retry } = useRecommendations();
  return (
    <section className="mt-8" aria-labelledby="recommendations-heading">
      <div className="flex items-end justify-between gap-4">
        <div><p className="text-xs font-semibold uppercase tracking-[0.16em] text-[var(--accent-gold)]">Picked for you</p><h2 id="recommendations-heading" className="mt-1 font-serif text-2xl font-bold text-[var(--text-heading)]">Recommended jobs</h2><p className="mt-1 text-sm text-[var(--text-muted)]">Personalized using your profile, experience, and activity.</p></div>
      </div>
      {loading ? <div className="mt-4 grid gap-4 sm:grid-cols-2 xl:grid-cols-3" aria-label="Loading recommendations">{[1, 2, 3].map((item) => <div key={item} className="h-72 animate-pulse rounded-lg border border-[var(--border-hairline)] bg-[var(--bg-card)]" />)}</div> : error ? <div role="alert" className="mt-4 flex flex-col gap-3 rounded-md border border-[var(--status-rejected)] bg-[var(--status-rejected)]/10 p-4 text-sm text-[var(--text-heading)] sm:flex-row sm:items-center sm:justify-between"><span className="flex items-center gap-2"><AlertCircle className="shrink-0 text-[var(--status-rejected)]" size={18} />{error}</span><button type="button" onClick={() => void retry()} className="inline-flex items-center justify-center gap-2 rounded-md border border-[var(--text-button-fill)] px-3 py-2 font-semibold text-[var(--text-button-fill)]"><RefreshCw size={15} />Try again</button></div> : recommendations.length === 0 ? <div className="mt-4 rounded-lg border border-[var(--border-hairline)] bg-[var(--bg-secondary)] px-6 py-10 text-center"><Sparkles className="mx-auto text-[var(--accent-gold)]" size={24} /><h3 className="mt-3 font-serif text-xl font-semibold text-[var(--text-heading)]">Recommendations are on the way</h3><p className="mx-auto mt-2 max-w-md text-sm leading-6 text-[var(--text-muted)]">Complete your profile and upload a resume to help SkillMatch find roles tailored to you.</p></div> : <div className="mt-4 grid gap-4 sm:grid-cols-2 xl:grid-cols-3">{recommendations.map((recommendation) => <RecommendationCard key={recommendation.id} recommendation={recommendation} />)}</div>}
    </section>
  );
};
