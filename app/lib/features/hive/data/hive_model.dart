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
  final DateTime createdAt;

  const HiveDisease({
    required this.id,
    required this.disease,
    required this.createdAt,
  });

  factory HiveDisease.fromJson(Map<String, dynamic> json) {
    return HiveDisease(
      id: json['id'] as int,
      disease: json['disease'] as String,
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }
}

class Hive {
  final int id;
  final int apiaryId;
  final String name;
  final String type;
  final bool active;
  final bool queenNeedsReplacement;
  final bool readyForHarvest;
  final bool needsFood;
  final bool boxNeedsAdding;
  final int gridRow;
  final int gridCol;
  final List<HiveDisease> diseases;
  final DateTime? lastInspectedAt;
  final DateTime? readyForHarvestSince;
  final DateTime? queenNeedsReplacementSince;
  final DateTime? needsFoodSince;
  final DateTime? boxNeedsAddingSince;

  const Hive({
    required this.id,
    required this.apiaryId,
    required this.name,
    required this.type,
    required this.active,
    required this.queenNeedsReplacement,
    required this.readyForHarvest,
    required this.needsFood,
    required this.boxNeedsAdding,
    required this.gridRow,
    required this.gridCol,
    this.diseases = const [],
    this.lastInspectedAt,
    this.readyForHarvestSince,
    this.queenNeedsReplacementSince,
    this.needsFoodSince,
    this.boxNeedsAddingSince,
  });

  factory Hive.fromJson(Map<String, dynamic> json) {
    final diseasesList = (json['diseases'] as List<dynamic>? ?? [])
        .map((e) => HiveDisease.fromJson(e as Map<String, dynamic>))
        .toList();
    final lastInspectedRaw = json['last_inspected_at'];
    final readyForHarvestSinceRaw = json['ready_for_harvest_since'];
    final queenNeedsReplacementSinceRaw = json['queen_needs_replacement_since'];
    final needsFoodSinceRaw = json['needs_food_since'];
    final boxNeedsAddingSinceRaw = json['box_needs_adding_since'];
    return Hive(
      id: json['id'] as int,
      apiaryId: json['apiary_id'] as int,
      name: json['name'] as String,
      type: json['type'] as String,
      active: json['active'] as bool,
      queenNeedsReplacement: json['queen_needs_replacement'] as bool? ?? false,
      readyForHarvest: json['ready_for_harvest'] as bool? ?? false,
      needsFood: json['needs_food'] as bool? ?? false,
      boxNeedsAdding: json['box_needs_adding'] as bool? ?? false,
      gridRow: json['grid_row'] as int,
      gridCol: json['grid_col'] as int,
      diseases: diseasesList,
      lastInspectedAt: lastInspectedRaw != null
          ? DateTime.parse(lastInspectedRaw as String)
          : null,
      readyForHarvestSince: readyForHarvestSinceRaw != null
          ? DateTime.parse(readyForHarvestSinceRaw as String)
          : null,
      queenNeedsReplacementSince: queenNeedsReplacementSinceRaw != null
          ? DateTime.parse(queenNeedsReplacementSinceRaw as String)
          : null,
      needsFoodSince: needsFoodSinceRaw != null
          ? DateTime.parse(needsFoodSinceRaw as String)
          : null,
      boxNeedsAddingSince: boxNeedsAddingSinceRaw != null
          ? DateTime.parse(boxNeedsAddingSinceRaw as String)
          : null,
    );
  }

  Hive copyWith({
    bool? queenNeedsReplacement,
    bool? needsFood,
    bool? boxNeedsAdding,
    List<HiveDisease>? diseases,
    DateTime? lastInspectedAt,
  }) {
    return Hive(
      id: id,
      apiaryId: apiaryId,
      name: name,
      type: type,
      active: active,
      queenNeedsReplacement:
          queenNeedsReplacement ?? this.queenNeedsReplacement,
      readyForHarvest: readyForHarvest,
      needsFood: needsFood ?? this.needsFood,
      boxNeedsAdding: boxNeedsAdding ?? this.boxNeedsAdding,
      gridRow: gridRow,
      gridCol: gridCol,
      diseases: diseases ?? this.diseases,
      lastInspectedAt: lastInspectedAt ?? this.lastInspectedAt,
      readyForHarvestSince: readyForHarvestSince,
      queenNeedsReplacementSince: queenNeedsReplacementSince,
      needsFoodSince: needsFoodSince,
      boxNeedsAddingSince: boxNeedsAddingSince,
    );
  }
}
