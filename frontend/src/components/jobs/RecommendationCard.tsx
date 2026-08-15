import React from 'react';
import { ArrowUpRight, MapPin, Sparkles } from 'lucide-react';
import { Link } from 'react-router-dom';
import { Recommendation } from '../../services/recommendations';

export const RecommendationCard: React.FC<{ recommendation: Recommendation }> = ({ recommendation }) => (
  <article className="flex h-full flex-col rounded-lg border border-[var(--border-hairline)] bg-[var(--bg-secondary)] p-5">
    <div className="flex items-start justify-between gap-3">
      <span className="grid h-10 w-10 place-items-center rounded-full bg-[var(--bg-card)] font-serif font-bold text-[var(--text-heading)]">{recommendation.company.charAt(0).toUpperCase()}</span>
      <span className="inline-flex items-center gap-1 rounded-full bg-[var(--bg-chip)] px-3 py-1 text-xs font-semibold text-[var(--text-heading)]"><Sparkles size={12} />{recommendation.relevanceScore ? `${recommendation.relevanceScore}% match` : recommendation.relevanceLabel || 'Recommended'}</span>
    </div>
    <h3 className="mt-4 text-lg font-semibold text-[var(--text-heading)]">{recommendation.title}</h3>
    <p className="mt-1 text-sm text-[var(--text-body)]">{recommendation.company}</p>
    <p className="mt-2 flex items-center gap-1.5 text-xs text-[var(--text-muted)]"><MapPin size={13} />{recommendation.location}</p>
    {recommendation.reasons.length > 0 && <div className="mt-4 border-t border-[var(--border-hairline)] pt-4"><p className="text-xs font-semibold uppercase tracking-wide text-[var(--accent-gold)]">Why it fits</p><ul className="mt-2 space-y-1.5 text-sm leading-5 text-[var(--text-insight)]">{recommendation.reasons.slice(0, 2).map((reason) => <li key={reason} className="flex gap-2"><span aria-hidden="true">•</span><span>{reason}</span></li>)}</ul></div>}
    <Link to={`/discover/${recommendation.id}`} className="mt-auto inline-flex items-center justify-center gap-2 pt-5 text-sm font-semibold text-[var(--text-button-fill)]">View recommendation <ArrowUpRight size={16} /></Link>
  </article>
);
