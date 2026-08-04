import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:app/features/harvest/data/harvest_model.dart';
import 'package:app/features/harvest/view/harvest_summary.dart';
import 'package:app/l10n/app_localizations.dart';

Widget _wrap(Widget child) => MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      locale: const Locale('en'),
      home: Scaffold(body: child),
    );

Harvest _harvest({
  int frames = 5,
  int halfFrames = 0,
  double kilograms = 12.5,
  String notes = '',
  String? harvestedByName,
}) =>
    Harvest(
      id: 1,
      hiveId: 1,
      harvestedByName: harvestedByName,
      harvestedAt: DateTime(2025, 6, 1, 10, 30),
      frames: frames,
      halfFrames: halfFrames,
      kilograms: kilograms,
      notes: notes,
    );

void main() {
  group('HarvestSummary', () {
    testWidgets('shows frame count without half frames', (tester) async {
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return HarvestSummary(harvest: _harvest(frames: 5), l10n: l10n);
      })));
      expect(find.text('5 frames'), findsOneWidget);
    });

    testWidgets('shows frame count plus half frames when present',
        (tester) async {
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return HarvestSummary(
          harvest: _harvest(frames: 5, halfFrames: 2),
          l10n: l10n,
        );
      })));
      expect(find.text('5 frames + 2 half frames'), findsOneWidget);
    });

    testWidgets('shows kilograms with two decimal places', (tester) async {
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return HarvestSummary(
          harvest: _harvest(kilograms: 12.5),
          l10n: l10n,
        );
      })));
      expect(find.text('12.50 kg'), findsOneWidget);
    });

    testWidgets('shows note when notes are present', (tester) async {
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return HarvestSummary(
          harvest: _harvest(notes: 'Great yield'),
          l10n: l10n,
        );
      })));
      expect(find.text('Note'), findsOneWidget);
      expect(find.text('Great yield'), findsOneWidget);
    });

    testWidgets('omits note when notes are empty', (tester) async {
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return HarvestSummary(harvest: _harvest(), l10n: l10n);
      })));
      expect(find.text('Note'), findsNothing);
    });

    testWidgets('shows harvested-by name when different from current user',
        (tester) async {
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return HarvestSummary(
          harvest: _harvest(harvestedByName: 'Alice'),
          l10n: l10n,
          currentUserName: 'Bob',
        );
      })));
      expect(find.text('By Alice'), findsOneWidget);
    });

    testWidgets('omits harvested-by name when it matches current user',
        (tester) async {
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return HarvestSummary(
          harvest: _harvest(harvestedByName: 'Alice'),
          l10n: l10n,
          currentUserName: 'Alice',
        );
      })));
      expect(find.textContaining('By '), findsNothing);
    });

    testWidgets('omits harvested-by name when empty', (tester) async {
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return HarvestSummary(
          harvest: _harvest(harvestedByName: ''),
          l10n: l10n,
        );
      })));
      expect(find.textContaining('By '), findsNothing);
    });

    testWidgets('shows date and time when showDate is true', (tester) async {
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return HarvestSummary(harvest: _harvest(), l10n: l10n);
      })));
      expect(find.textContaining('Jun 1, 2025'), findsOneWidget);
      expect(find.textContaining('10:30'), findsOneWidget);
    });

    testWidgets('does not show date when showDate is false', (tester) async {
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return HarvestSummary(
          harvest: _harvest(),
          l10n: l10n,
          showDate: false,
        );
      })));
      expect(find.textContaining('2025'), findsNothing);
    });
  });
}
