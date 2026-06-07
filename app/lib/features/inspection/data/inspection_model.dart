class Inspection {
  final int id;
  final int hiveId;
  final DateTime inspectedAt;
  final String queenSeen;
  final String broodPattern;
  final String aggressiveness;
  final int? framesBrood;
  final int? framesHoney;
  final int? framesPollen;
  final int? framesAddedDrawn;
  final int? framesAddedFoundation;
  final int? framesAddedHoney;
  final int? queenCellsCount;
  final bool queenAdded;
  final String notes;
  final int photoCount;

  const Inspection({
    required this.id,
    required this.hiveId,
    required this.inspectedAt,
    required this.queenSeen,
    required this.broodPattern,
    required this.aggressiveness,
    this.framesBrood,
    this.framesHoney,
    this.framesPollen,
    this.framesAddedDrawn,
    this.framesAddedFoundation,
    this.framesAddedHoney,
    this.queenCellsCount,
    required this.queenAdded,
    required this.notes,
    this.photoCount = 0,
  });

  factory Inspection.fromJson(Map<String, dynamic> json) {
    return Inspection(
      id: json['id'] as int,
      hiveId: json['hive_id'] as int,
      inspectedAt: DateTime.parse(json['inspected_at'] as String),
      queenSeen: json['queen_status'] as String? ?? '',
      broodPattern: json['brood_pattern'] as String? ?? '',
      aggressiveness: json['aggressiveness'] as String? ?? '',
      framesBrood: json['frames_brood'] as int?,
      framesHoney: json['frames_honey'] as int?,
      framesPollen: json['frames_pollen'] as int?,
      framesAddedDrawn: json['frames_added_drawn'] as int?,
      framesAddedFoundation: json['frames_added_foundation'] as int?,
      framesAddedHoney: json['frames_added_honey'] as int?,
      queenCellsCount: json['queen_cells_count'] as int?,
      queenAdded: json['queen_added'] as bool? ?? false,
      notes: json['notes'] as String? ?? '',
      photoCount: json['photo_count'] as int? ?? 0,
    );
  }
}

const broodPatternValues = ['none', 'poor', 'good', 'excellent'];
const aggressivenessValues = ['calm', 'mild', 'aggressive', 'very_aggressive'];
