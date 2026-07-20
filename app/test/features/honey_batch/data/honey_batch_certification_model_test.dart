import 'package:flutter_test/flutter_test.dart';
import 'package:app/features/honey_batch/data/honey_batch_certification_model.dart';

void main() {
  group('CertificationStatus', () {
    test('fromJson/toJson round-trip for all values', () {
      for (final status in CertificationStatus.values) {
        expect(CertificationStatus.fromJson(status.toJson()), status);
      }
    });

    test('isTerminal is true only for confirmed/failed/reverted', () {
      const terminal = {
        CertificationStatus.confirmed,
        CertificationStatus.failed,
        CertificationStatus.reverted,
      };
      for (final status in CertificationStatus.values) {
        expect(status.isTerminal, terminal.contains(status), reason: status.name);
      }
    });

    test('isLive is true only for submitted/pendingConfirmation/confirmed', () {
      const live = {
        CertificationStatus.submitted,
        CertificationStatus.pendingConfirmation,
        CertificationStatus.confirmed,
      };
      for (final status in CertificationStatus.values) {
        expect(status.isLive, live.contains(status), reason: status.name);
      }
    });
  });

  group('HoneyBatchCertificationModel.fromJson', () {
    test('parses all fields when present', () {
      final json = {
        'status': 'confirmed',
        'transaction_hash': '0xabc',
        'block_number': 123,
        'gas_used': 21000,
        'confirmation_timestamp': '2024-05-01T10:00:00Z',
        'created_at': '2024-04-30T09:00:00Z',
      };
      final m = HoneyBatchCertificationModel.fromJson(json);
      expect(m.status, CertificationStatus.confirmed);
      expect(m.transactionHash, '0xabc');
      expect(m.blockNumber, 123);
      expect(m.gasUsed, 21000);
      expect(
        m.confirmationTimestamp!
            .isAtSameMomentAs(DateTime.parse('2024-05-01T10:00:00Z')),
        isTrue,
      );
      expect(
        m.createdAt!.isAtSameMomentAs(DateTime.parse('2024-04-30T09:00:00Z')),
        isTrue,
      );
    });

    test('parses null optional fields', () {
      final json = {'status': 'queued'};
      final m = HoneyBatchCertificationModel.fromJson(json);
      expect(m.status, CertificationStatus.queued);
      expect(m.transactionHash, isNull);
      expect(m.blockNumber, isNull);
      expect(m.gasUsed, isNull);
      expect(m.confirmationTimestamp, isNull);
      expect(m.createdAt, isNull);
    });
  });
}
