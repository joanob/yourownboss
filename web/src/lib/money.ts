/**
 * Formats money from int64 (thousandths) to string with 3 decimals
 * @param money - Amount in thousandths (e.g., 50000000 = 50,000.000)
 * @returns Formatted string (e.g., "50,000.000")
 */
export function formatMoney(money: number): string {
  const amount = money / 1000;
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
