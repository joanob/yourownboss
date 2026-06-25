export function formatMoney(amount: number): string {
  return amount.toLocaleString('es-ES') + ' €';
}

export function formatSeconds(totalSeconds: number): string {
  if (totalSeconds <= 0) return '0s';
  const h = Math.floor(totalSeconds / 3600);
  const m = Math.floor((totalSeconds % 3600) / 60);
  const s = totalSeconds % 60;
  if (h > 0) return `${h}h ${m > 0 ? `${m}m` : ''}`.trim();
  if (m > 0) return `${m}m ${s > 0 ? `${s}s` : ''}`.trim();
  return `${s}s`;
}

export function formatCountdown(ms: number): string {
  if (ms <= 0) return '¡Listo!';
  return formatSeconds(Math.ceil(ms / 1000));
}

export function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString('es-ES', {
    dateStyle: 'medium',
    timeStyle: 'short',
  });
}
