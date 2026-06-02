class Hive {
  final int id;
  final int apiaryId;
  final String name;
  final String type;
  final bool active;
  final int gridRow;
  final int gridCol;

  const Hive({
    required this.id,
    required this.apiaryId,
    required this.name,
    required this.type,
    required this.active,
    required this.gridRow,
    required this.gridCol,
  });

  factory Hive.fromJson(Map<String, dynamic> json) {
    return Hive(
      id: json['id'] as int,
      apiaryId: json['apiary_id'] as int,
      name: json['name'] as String,
      type: json['type'] as String,
      active: json['active'] as bool,
      gridRow: json['grid_row'] as int,
      gridCol: json['grid_col'] as int,
    );
  }
}
