class Feeding {
  final int id;
  final int hiveId;
  final DateTime fedAt;
  final String feedType;
  final String amount;
  final String notes;
  final String? fedByName;

  const Feeding({
    required this.id,
    required this.hiveId,
    required this.fedAt,
    required this.feedType,
    required this.amount,
    required this.notes,
    this.fedByName,
  });

  factory Feeding.fromJson(Map<String, dynamic> json) {
    return Feeding(
      id: json['id'] as int,
      hiveId: json['hive_id'] as int,
      fedAt: DateTime.parse(json['fed_at'] as String).toLocal(),
      feedType: json['feed_type'] as String? ?? '',
      amount: json['amount'] as String? ?? '',
      notes: json['notes'] as String? ?? '',
      fedByName: json['fed_by_name'] as String?,
    );
  }
}
