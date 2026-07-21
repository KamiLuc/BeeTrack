import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:app/features/honey_batch/data/honey_batch_certification_model.dart';
import 'package:app/features/honey_batch/view/certification_status_badge.dart';
import 'package:app/l10n/app_localizations.dart';

Widget _wrap(CertificationStatus? status) => MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      locale: const Locale('en'),
      home: Scaffold(
        body: CertificationStatusBadge(status: status),
      ),
    );

void main() {
  late AppLocalizations l10n;
  setUpAll(() async {
    l10n = await AppLocalizations.delegate.load(const Locale('en'));
  });

  group('CertificationStatusBadge', () {
    testWidgets('shows "not certified" label when status is null', (
      tester,
    ) async {
      await tester.pumpWidget(_wrap(null));
      expect(find.text(l10n.honeyBatchNotCertified), findsOneWidget);
    });

    final expectedLabels = {
      CertificationStatus.queued: (l) => l.honeyBatchInProgress,
      CertificationStatus.submitting: (l) => l.honeyBatchInProgress,
      CertificationStatus.submitted: (l) => l.honeyBatchInProgress,
      CertificationStatus.pendingConfirmation: (l) => l.honeyBatchInProgress,
      CertificationStatus.confirmed: (l) => l.honeyBatchStatusConfirmed,
      CertificationStatus.failed: (l) => l.honeyBatchStatusFailed,
      CertificationStatus.reverted: (l) => l.honeyBatchStatusReverted,
    };

    for (final entry in expectedLabels.entries) {
      testWidgets('shows label for ${entry.key.name}', (tester) async {
        await tester.pumpWidget(_wrap(entry.key));
        expect(find.text(entry.value(l10n)), findsOneWidget);
      });
    }
  });
}
