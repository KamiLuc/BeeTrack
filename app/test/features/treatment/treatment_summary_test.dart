import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:app/features/treatment/data/treatment_model.dart';
import 'package:app/features/treatment/view/treatment_summary.dart';
import 'package:app/l10n/app_localizations.dart';

Widget _wrap(Widget child) => MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      locale: const Locale('en'),
      home: Scaffold(body: child),
    );

Treatment _treatment({
  String medicineName = 'Apivar',
  String dose = '2',
  String notes = '',
  String? treatedByName,
}) =>
    Treatment(
      id: 1,
      hiveId: 1,
      treatedAt: DateTime(2025, 6, 1, 10, 30),
      medicineName: medicineName,
      dose: dose,
      notes: notes,
      treatedByName: treatedByName,
    );

void main() {
  group('TreatmentSummary', () {
    testWidgets('shows medicine name and numeric dose as a count',
        (tester) async {
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return TreatmentSummary(
          treatment: _treatment(medicineName: 'Apivar', dose: '2'),
          l10n: l10n,
        );
      })));
      expect(find.text('Apivar · 2 doses'), findsOneWidget);
    });

    testWidgets('shows singular dose for a dose of 1', (tester) async {
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return TreatmentSummary(
          treatment: _treatment(medicineName: 'Apivar', dose: '1'),
          l10n: l10n,
        );
      })));
      expect(find.text('Apivar · 1 dose'), findsOneWidget);
    });

    testWidgets('shows raw dose text when it is not numeric', (tester) async {
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return TreatmentSummary(
          treatment: _treatment(medicineName: 'Oxalic acid', dose: '5ml'),
          l10n: l10n,
        );
      })));
      expect(find.text('Oxalic acid · 5ml'), findsOneWidget);
    });

    testWidgets('shows note when notes are present', (tester) async {
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return TreatmentSummary(
          treatment: _treatment(notes: 'No side effects'),
          l10n: l10n,
        );
      })));
      expect(find.text('Note'), findsOneWidget);
      expect(find.text('No side effects'), findsOneWidget);
    });

    testWidgets('omits note when notes are empty', (tester) async {
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return TreatmentSummary(treatment: _treatment(), l10n: l10n);
      })));
      expect(find.text('Note'), findsNothing);
    });

    testWidgets('shows treated-by name when different from current user',
        (tester) async {
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return TreatmentSummary(
          treatment: _treatment(treatedByName: 'Alice'),
          l10n: l10n,
          currentUserName: 'Bob',
        );
      })));
      expect(find.text('By Alice'), findsOneWidget);
    });

    testWidgets('omits treated-by name when it matches current user',
        (tester) async {
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return TreatmentSummary(
          treatment: _treatment(treatedByName: 'Alice'),
          l10n: l10n,
          currentUserName: 'Alice',
        );
      })));
      expect(find.textContaining('By '), findsNothing);
    });

    testWidgets('shows date and time when showDate is true', (tester) async {
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return TreatmentSummary(treatment: _treatment(), l10n: l10n);
      })));
      expect(find.textContaining('Jun 1, 2025'), findsOneWidget);
      expect(find.textContaining('10:30'), findsOneWidget);
    });

    testWidgets('does not show date when showDate is false', (tester) async {
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return TreatmentSummary(
          treatment: _treatment(),
          l10n: l10n,
          showDate: false,
        );
      })));
      expect(find.textContaining('2025'), findsNothing);
    });
  });
}
