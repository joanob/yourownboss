/**
 * Formats money from int64 (thousandths) to string.
 * @param money - Amount in thousandths (e.g., 50000000 = 50,000.000)
 * @param displayDecimals - Whether to display decimals (default: true). If false, decimals are hidden and only the integer portion is shown.
 * @returns Formatted string (e.g., "50,000.000" or "50,000")
 */
export function formatMoney(money: number, displayDecimals = true): string {
  const amount = money / 1000;
  const hasNonZeroDecimals = Math.abs(money % 1000) !== 0;

  // If caller requested to hide decimals and there are no non-zero decimals,
  // show only the integer portion. Otherwise show three decimal places.
  if (!displayDecimals && !hasNonZeroDecimals) {
    return Math.trunc(amount).toLocaleString('en-US', {
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    });
  }

  return amount.toLocaleString('en-US', {
    minimumFractionDigits: 3,
    maximumFractionDigits: 3,
  });
}

/**
 * Parses formatted money string to int64 (thousandths)
 * @param moneyStr - Formatted string (e.g., "50,000.000")
 * @returns Amount in thousandths (e.g., 50000000)
 */
export function parseMoney(moneyStr: string): number {
  const cleaned = moneyStr.replace(/,/g, '');
  const amount = parseFloat(cleaned);
  return Math.round(amount * 1000);
}
