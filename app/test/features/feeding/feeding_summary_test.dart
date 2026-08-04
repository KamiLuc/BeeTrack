import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:app/features/feeding/data/feeding_model.dart';
import 'package:app/features/feeding/view/feeding_summary.dart';
import 'package:app/l10n/app_localizations.dart';

Widget _wrap(Widget child) => MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      locale: const Locale('en'),
      home: Scaffold(body: child),
    );

Feeding _feeding({
  String feedType = 'Syrup',
  String amount = '1L',
  String notes = '',
  String? fedByName,
}) =>
    Feeding(
      id: 1,
      hiveId: 1,
      fedAt: DateTime(2025, 6, 1, 10, 30),
      feedType: feedType,
      amount: amount,
      notes: notes,
      fedByName: fedByName,
    );

void main() {
  group('FeedingSummary', () {
    testWidgets('shows feed type and amount', (tester) async {
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return FeedingSummary(feeding: _feeding(), l10n: l10n);
      })));
      expect(find.text('Syrup · 1L'), findsOneWidget);
    });

    testWidgets('shows note when notes are present', (tester) async {
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return FeedingSummary(
          feeding: _feeding(notes: 'Ate it all'),
          l10n: l10n,
        );
      })));
      expect(find.text('Note'), findsOneWidget);
      expect(find.text('Ate it all'), findsOneWidget);
    });

    testWidgets('omits note when notes are empty', (tester) async {
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return FeedingSummary(feeding: _feeding(), l10n: l10n);
      })));
      expect(find.text('Note'), findsNothing);
    });

    testWidgets('shows fed-by name when different from current user',
        (tester) async {
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return FeedingSummary(
          feeding: _feeding(fedByName: 'Alice'),
          l10n: l10n,
          currentUserName: 'Bob',
        );
      })));
      expect(find.text('By Alice'), findsOneWidget);
    });

    testWidgets('omits fed-by name when it matches current user',
        (tester) async {
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return FeedingSummary(
          feeding: _feeding(fedByName: 'Alice'),
          l10n: l10n,
          currentUserName: 'Alice',
        );
      })));
      expect(find.textContaining('By '), findsNothing);
    });

    testWidgets('shows date and time when showDate is true', (tester) async {
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return FeedingSummary(feeding: _feeding(), l10n: l10n);
      })));
      expect(find.textContaining('Jun 1, 2025'), findsOneWidget);
      expect(find.textContaining('10:30'), findsOneWidget);
    });

    testWidgets('does not show date when showDate is false', (tester) async {
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return FeedingSummary(
          feeding: _feeding(),
          l10n: l10n,
          showDate: false,
        );
      })));
      expect(find.textContaining('2025'), findsNothing);
    });
  });
}
