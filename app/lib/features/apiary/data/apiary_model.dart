class Apiary {
  final int id;
  final String name;
  final double? lat;
  final double? lng;
  final int gridRows;
  final int gridCols;
  final int hiveCount;
  final String userRole;
  final DateTime? lastInspectedAt;

  const Apiary({
    required this.id,
    required this.name,
    required this.lat,
    required this.lng,
    required this.gridRows,
    required this.gridCols,
    required this.hiveCount,
    required this.userRole,
    this.lastInspectedAt,
  });

  factory Apiary.fromJson(Map<String, dynamic> json) {
    return Apiary(
      id: json['id'] as int,
      name: json['name'] as String,
      lat: (json['lat'] as num?)?.toDouble(),
      lng: (json['lng'] as num?)?.toDouble(),
      gridRows: json['grid_rows'] as int,
      gridCols: json['grid_cols'] as int,
      hiveCount: json['hive_count'] as int? ?? 0,
      userRole: json['user_role'] as String,
      lastInspectedAt: json['last_inspected_at'] != null
          ? DateTime.parse(json['last_inspected_at'] as String)
          : null,
    );
  }
}
