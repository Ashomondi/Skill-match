import React from 'react';

export const MatchRing: React.FC<{ value: number; size?: number }> = ({ value, size = 68 }) => {
  const clamped = Math.max(0, Math.min(100, Math.round(Number.isFinite(value) ? value : 0)));

  return (
    <div
      className="relative grid shrink-0 place-items-center rounded-full bg-[conic-gradient(var(--text-heading)_calc(var(--match)*1%),var(--border-hairline)_0)]"
      style={{ width: size, height: size, '--match': clamped } as React.CSSProperties}
    >
      <div
        className="grid place-items-center rounded-full bg-[var(--bg-secondary)]"
        style={{ width: size - 8, height: size - 8 }}
      >
        <b className="text-sm text-[var(--text-heading)]">{clamped}%</b>
        <span className="-mt-1 text-[8px] uppercase text-[var(--text-muted)]">match</span>
      </div>
    </div>
  );
};
