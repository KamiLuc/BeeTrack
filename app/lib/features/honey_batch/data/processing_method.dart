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

  String get label => switch (this) {
        ProcessingMethod.raw => 'Raw',
        ProcessingMethod.filtered => 'Filtered',
        ProcessingMethod.pasteurized => 'Pasteurized',
      };
}
