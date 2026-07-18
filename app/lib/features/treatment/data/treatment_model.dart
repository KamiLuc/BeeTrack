class Treatment {
  final int id;
  final int hiveId;
  final DateTime treatedAt;
  final String medicineName;
  final String dose;
  final String notes;
  final String? treatedByName;

  const Treatment({
    required this.id,
    required this.hiveId,
    required this.treatedAt,
    required this.medicineName,
    required this.dose,
    required this.notes,
    this.treatedByName,
  });

  factory Treatment.fromJson(Map<String, dynamic> json) {
    return Treatment(
      id: json['id'] as int,
      hiveId: json['hive_id'] as int,
      treatedAt: DateTime.parse(json['treated_at'] as String).toLocal(),
      medicineName: json['medicine_name'] as String? ?? '',
      dose: json['dose'] as String? ?? '1',
      notes: json['notes'] as String? ?? '',
      treatedByName: json['treated_by_name'] as String?,
    );
  }
}
