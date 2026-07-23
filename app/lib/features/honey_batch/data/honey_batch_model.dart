import 'honey_batch_certification_model.dart';
import 'honey_batch_certification_request_model.dart';
import 'processing_method.dart';

class HoneyBatchModel {
  final int id;
  final String verificationToken;
  final String verificationUrl;
  final DateTime gatheringDate;
  final int amountGrams;
  final ProcessingMethod processingMethod;
  final String honeyType;
  final String pdfFilename;
  final String pdfFileHash;
  final DateTime createdAt;
  final DateTime updatedAt;
  final HoneyBatchCertificationModel? certification;
  final HoneyBatchCertificationRequestModel? certificationRequest;

  const HoneyBatchModel({
    required this.id,
    required this.verificationToken,
    required this.verificationUrl,
    required this.gatheringDate,
    required this.amountGrams,
    required this.processingMethod,
    required this.honeyType,
    required this.pdfFilename,
    required this.pdfFileHash,
    required this.createdAt,
    required this.updatedAt,
    this.certification,
    this.certificationRequest,
  });

  double get amountKg => amountGrams / 1000;

  factory HoneyBatchModel.fromJson(Map<String, dynamic> json) {
    return HoneyBatchModel(
      id: json['id'] as int,
      verificationToken: json['verification_token'] as String,
      verificationUrl: json['verification_url'] as String,
      gatheringDate: DateTime.parse(json['gathering_date'] as String).toLocal(),
      amountGrams: json['amount_grams'] as int,
      processingMethod:
          ProcessingMethod.fromJson(json['processing_method'] as String),
      honeyType: json['honey_type'] as String,
      pdfFilename: json['pdf_filename'] as String? ?? '',
      pdfFileHash: json['pdf_file_hash'] as String,
      createdAt: DateTime.parse(json['created_at'] as String).toLocal(),
      updatedAt: DateTime.parse(json['updated_at'] as String).toLocal(),
      certification: json['certification'] != null
          ? HoneyBatchCertificationModel.fromJson(
              json['certification'] as Map<String, dynamic>)
          : null,
      certificationRequest: json['certification_request'] != null
          ? HoneyBatchCertificationRequestModel.fromJson(
              json['certification_request'] as Map<String, dynamic>)
          : null,
    );
  }
}
