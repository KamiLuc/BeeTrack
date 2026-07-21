enum CertificationRequestStatus {
  pending('pending'),
  approved('approved'),
  rejected('rejected');

  final String value;

  const CertificationRequestStatus(this.value);

  factory CertificationRequestStatus.fromJson(String value) {
    return CertificationRequestStatus.values.firstWhere((s) => s.value == value);
  }

  String toJson() => value;
}

class HoneyBatchCertificationRequestModel {
  final CertificationRequestStatus status;
  final String? rejectionReason;
  final DateTime? createdAt;

  const HoneyBatchCertificationRequestModel({
    required this.status,
    this.rejectionReason,
    this.createdAt,
  });

  factory HoneyBatchCertificationRequestModel.fromJson(Map<String, dynamic> json) {
    return HoneyBatchCertificationRequestModel(
      status: CertificationRequestStatus.fromJson(json['status'] as String),
      rejectionReason: json['rejection_reason'] as String?,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'] as String).toLocal()
          : null,
    );
  }
}
