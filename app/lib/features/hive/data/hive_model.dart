const diseaseValues = [
  'varroa',
  'nosema',
  'dwv',
  'american_foulbrood',
  'chalkbrood',
  'european_foulbrood',
  'laying_workers',
];

class HiveDisease {
  final int id;
  final String disease;

  const HiveDisease({required this.id, required this.disease});

  factory HiveDisease.fromJson(Map<String, dynamic> json) {
    return HiveDisease(
      id: json['id'] as int,
      disease: json['disease'] as String,
    );
  }
}

class Hive {
  final int id;
  final int apiaryId;
  final String name;
  final String type;
  final bool active;
  final bool queenless;
  final bool readyForHarvest;
  final bool needsFood;
  final int gridRow;
  final int gridCol;
  final List<HiveDisease> diseases;
  final DateTime? lastInspectedAt;

  const Hive({
    required this.id,
    required this.apiaryId,
    required this.name,
    required this.type,
    required this.active,
    required this.queenless,
    required this.readyForHarvest,
    required this.needsFood,
    required this.gridRow,
    required this.gridCol,
    this.diseases = const [],
    this.lastInspectedAt,
  });

  factory Hive.fromJson(Map<String, dynamic> json) {
    final diseasesList = (json['diseases'] as List<dynamic>? ?? [])
        .map((e) => HiveDisease.fromJson(e as Map<String, dynamic>))
        .toList();
    final lastInspectedRaw = json['last_inspected_at'];
    return Hive(
      id: json['id'] as int,
      apiaryId: json['apiary_id'] as int,
      name: json['name'] as String,
      type: json['type'] as String,
      active: json['active'] as bool,
      queenless: json['queenless'] as bool? ?? false,
      readyForHarvest: json['ready_for_harvest'] as bool? ?? false,
      needsFood: json['needs_food'] as bool? ?? false,
      gridRow: json['grid_row'] as int,
      gridCol: json['grid_col'] as int,
      diseases: diseasesList,
      lastInspectedAt: lastInspectedRaw != null
          ? DateTime.parse(lastInspectedRaw as String)
          : null,
    );
  }

  Hive copyWith({
    bool? queenless,
    bool? needsFood,
    List<HiveDisease>? diseases,
    DateTime? lastInspectedAt,
  }) {
    return Hive(
      id: id,
      apiaryId: apiaryId,
      name: name,
      type: type,
      active: active,
      queenless: queenless ?? this.queenless,
      readyForHarvest: readyForHarvest,
      needsFood: needsFood ?? this.needsFood,
      gridRow: gridRow,
      gridCol: gridCol,
      diseases: diseases ?? this.diseases,
      lastInspectedAt: lastInspectedAt ?? this.lastInspectedAt,
    );
  }
}
