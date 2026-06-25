import { useState, useEffect } from 'react';

interface CountdownResult {
  remaining: number; // ms remaining
  isDone: boolean;
}

export function useCountdown(endsAt: string | null): CountdownResult {
  const [remaining, setRemaining] = useState<number>(() => {
    if (!endsAt) return 0;
    return Math.max(0, new Date(endsAt).getTime() - Date.now());
  });

  useEffect(() => {
    if (!endsAt) {
      setRemaining(0);
      return;
    }

    const update = () => {
      const diff = new Date(endsAt).getTime() - Date.now();
      setRemaining(Math.max(0, diff));
    };

    update();
    const id = setInterval(update, 1000);
    return () => clearInterval(id);
  }, [endsAt]);

  return { remaining, isDone: remaining === 0 && endsAt !== null };
}
