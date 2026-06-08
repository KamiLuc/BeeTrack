class Harvest {
  final int id;
  final int hiveId;
  final String? harvestedByName;
  final DateTime harvestedAt;
  final int frames;
  final int halfFrames;
  final double kilograms;
  final String notes;

  const Harvest({
    required this.id,
    required this.hiveId,
    this.harvestedByName,
    required this.harvestedAt,
    required this.frames,
    required this.halfFrames,
    required this.kilograms,
    required this.notes,
  });

  factory Harvest.fromJson(Map<String, dynamic> json) {
    return Harvest(
      id: json['id'] as int,
      hiveId: json['hive_id'] as int,
      harvestedByName: json['harvested_by_name'] as String?,
      harvestedAt: DateTime.parse(json['harvested_at'] as String).toLocal(),
      frames: json['frames'] as int? ?? 0,
      halfFrames: json['half_frames'] as int? ?? 0,
      kilograms: (json['kilograms'] as num?)?.toDouble() ?? 0.0,
      notes: json['notes'] as String? ?? '',
    );
  }
}
