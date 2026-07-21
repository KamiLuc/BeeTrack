// Mirrors app/lib/features/honey_batch/data/honey_batch_model.dart's
// amountKg getter + toStringAsFixed(1) display (e.g. "1.5 kg").
export function formatKg(grams: number): string {
  return `${(grams / 1000).toFixed(1)} kg`;
}
