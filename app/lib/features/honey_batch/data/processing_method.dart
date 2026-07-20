import '../../../l10n/app_localizations.dart';

enum ProcessingMethod {
  raw('raw'),
  filtered('filtered'),
  pasteurized('pasteurized');

  final String value;

  const ProcessingMethod(this.value);

  factory ProcessingMethod.fromJson(String value) {
    return ProcessingMethod.values.firstWhere(
      (m) => m.value == value,
      orElse: () => ProcessingMethod.raw,
    );
  }

  String toJson() => value;
}

String processingMethodLabel(AppLocalizations l10n, ProcessingMethod method) {
  return switch (method) {
    ProcessingMethod.raw => l10n.honeyBatchMethodRaw,
    ProcessingMethod.filtered => l10n.honeyBatchMethodFiltered,
    ProcessingMethod.pasteurized => l10n.honeyBatchMethodPasteurized,
  };
}
