import 'package:flutter_test/flutter_test.dart';
import 'package:app/features/honey_batch/data/honey_batch_certification_model.dart';
import 'package:app/features/honey_batch/data/honey_batch_model.dart';
import 'package:app/features/honey_batch/data/processing_method.dart';

void main() {
  Map<String, dynamic> baseJson({Object? certification}) => {
        'id': 1,
        'verification_token': 'tok-123',
        'gathering_date': '2024-05-01T00:00:00Z',
        'amount_grams': 2500,
        'processing_method': 'filtered',
        'honey_type': 'Acacia',
        'pdf_file_hash': 'hash-abc',
        'created_at': '2024-05-02T08:00:00Z',
        'updated_at': '2024-05-03T09:00:00Z',
        if (certification != null) 'certification': certification,
      };

  group('HoneyBatchModel.fromJson', () {
    test('parses fields with certification absent', () {
      final m = HoneyBatchModel.fromJson(baseJson());
      expect(m.id, 1);
      expect(m.verificationToken, 'tok-123');
      expect(
        m.gatheringDate.isAtSameMomentAs(DateTime.parse('2024-05-01T00:00:00Z')),
        isTrue,
      );
      expect(m.amountGrams, 2500);
      expect(m.processingMethod, ProcessingMethod.filtered);
      expect(m.honeyType, 'Acacia');
      expect(m.pdfFileHash, 'hash-abc');
      expect(
        m.createdAt.isAtSameMomentAs(DateTime.parse('2024-05-02T08:00:00Z')),
        isTrue,
      );
      expect(
        m.updatedAt.isAtSameMomentAs(DateTime.parse('2024-05-03T09:00:00Z')),
        isTrue,
      );
      expect(m.certification, isNull);
    });

    test('parses fields with certification null', () {
      final json = baseJson()..['certification'] = null;
      final m = HoneyBatchModel.fromJson(json);
      expect(m.certification, isNull);
    });

    test('parses nested certification when present', () {
      final m = HoneyBatchModel.fromJson(
        baseJson(certification: {'status': 'submitted'}),
      );
      expect(m.certification, isNotNull);
      expect(m.certification!.status, CertificationStatus.submitted);
    });

    test('amountKg converts grams to kilograms', () {
      final m = HoneyBatchModel.fromJson(baseJson());
      expect(m.amountKg, 2.5);
    });
  });
}
