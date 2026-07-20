import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:app/features/honey_batch/data/honey_batch_certification_model.dart';
import 'package:app/features/honey_batch/data/honey_batch_model.dart';
import 'package:app/features/honey_batch/data/processing_method.dart';
import 'package:app/features/honey_batch/view/honey_batches_home_screen.dart';
import 'package:app/l10n/app_localizations.dart';

HoneyBatchModel _batch({
  ProcessingMethod processingMethod = ProcessingMethod.raw,
  String honeyType = 'Wildflower',
  int amountGrams = 1500,
  HoneyBatchCertificationModel? certification,
}) =>
    HoneyBatchModel(
      id: 1,
      verificationToken: 'token-1',
      gatheringDate: DateTime(2025, 6, 1),
      amountGrams: amountGrams,
      processingMethod: processingMethod,
      honeyType: honeyType,
      pdfFileHash: 'hash-1',
      certification: certification,
      createdAt: DateTime(2025, 6, 1),
      updatedAt: DateTime(2025, 6, 1),
    );

Widget _wrap(HoneyBatchModel batch) => MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      locale: const Locale('en'),
      home: Scaffold(
        body: HoneyBatchCard(batch: batch, onTap: () {}),
      ),
    );

void main() {
  late AppLocalizations l10n;
  setUpAll(() async {
    l10n = await AppLocalizations.delegate.load(const Locale('en'));
  });

  group('HoneyBatchCard', () {
    testWidgets('shows honey type, processing method and amount in kg', (
      tester,
    ) async {
      await tester.pumpWidget(
        _wrap(_batch(
          honeyType: 'Wildflower',
          processingMethod: ProcessingMethod.filtered,
          amountGrams: 2500,
        )),
      );

      expect(
        find.text('Wildflower · ${processingMethodLabel(l10n, ProcessingMethod.filtered)}'),
        findsOneWidget,
      );
      expect(find.text('2.5 kg'), findsOneWidget);
    });

    testWidgets('shows "not certified" badge when certification is null', (
      tester,
    ) async {
      await tester.pumpWidget(_wrap(_batch(certification: null)));
      expect(find.text(l10n.honeyBatchNotCertified), findsOneWidget);
    });

    testWidgets('shows the certification status badge when certified', (
      tester,
    ) async {
      await tester.pumpWidget(
        _wrap(_batch(
          certification: const HoneyBatchCertificationModel(
            status: CertificationStatus.confirmed,
          ),
        )),
      );
      expect(find.text(l10n.honeyBatchStatusConfirmed), findsOneWidget);
    });

    testWidgets('invokes onTap when the card is tapped', (tester) async {
      var tapped = false;
      await tester.pumpWidget(
        MaterialApp(
          localizationsDelegates: AppLocalizations.localizationsDelegates,
          supportedLocales: AppLocalizations.supportedLocales,
          locale: const Locale('en'),
          home: Scaffold(
            body: HoneyBatchCard(
              batch: _batch(),
              onTap: () => tapped = true,
            ),
          ),
        ),
      );

      await tester.tap(find.byType(InkWell));
      await tester.pump();

      expect(tapped, isTrue);
    });
  });
}
