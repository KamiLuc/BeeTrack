enum CertificationStatus {
  queued('queued'),
  submitting('submitting'),
  submitted('submitted'),
  pendingConfirmation('pending_confirmation'),
  confirmed('confirmed'),
  failed('failed'),
  reverted('reverted');

  final String value;

  const CertificationStatus(this.value);

  factory CertificationStatus.fromJson(String value) {
    return CertificationStatus.values.firstWhere((s) => s.value == value);
  }

  String toJson() => value;

  bool get isTerminal => switch (this) {
        CertificationStatus.confirmed ||
        CertificationStatus.failed ||
        CertificationStatus.reverted =>
          true,
        _ => false,
      };

  bool get isLive => switch (this) {
        CertificationStatus.submitted ||
        CertificationStatus.pendingConfirmation ||
        CertificationStatus.confirmed =>
          true,
        _ => false,
      };
}

class HoneyBatchCertificationModel {
  final CertificationStatus status;
  final String? transactionHash;
  final int? blockNumber;
  final int? gasUsed;
  final DateTime? confirmationTimestamp;
  final DateTime? createdAt;

  const HoneyBatchCertificationModel({
    required this.status,
    this.transactionHash,
    this.blockNumber,
    this.gasUsed,
    this.confirmationTimestamp,
    this.createdAt,
  });

  factory HoneyBatchCertificationModel.fromJson(Map<String, dynamic> json) {
    return HoneyBatchCertificationModel(
      status: CertificationStatus.fromJson(json['status'] as String),
      transactionHash: json['transaction_hash'] as String?,
      blockNumber: json['block_number'] as int?,
      gasUsed: json['gas_used'] as int?,
      confirmationTimestamp: json['confirmation_timestamp'] != null
          ? DateTime.parse(json['confirmation_timestamp'] as String).toLocal()
          : null,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'] as String).toLocal()
          : null,
    );
  }
}
