class InspectionImage {
  final int id;
  final int inspectionId;
  final String mimeType;
  final DateTime createdAt;

  const InspectionImage({
    required this.id,
    required this.inspectionId,
    required this.mimeType,
    required this.createdAt,
  });

  factory InspectionImage.fromJson(Map<String, dynamic> json) {
    return InspectionImage(
      id: json['id'] as int,
      inspectionId: json['inspection_id'] as int,
      mimeType: json['mime_type'] as String,
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }
}
