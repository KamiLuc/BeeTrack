import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:app/features/harvest/data/harvest_model.dart';
import 'package:app/features/harvest/view/harvest_history_screen.dart';
import 'package:app/l10n/app_localizations.dart';

Harvest _harvest({required int frames, required int halfFrames}) => Harvest(
      id: 1,
      hiveId: 1,
      harvestedAt: DateTime(2026, 6, 8),
      frames: frames,
      halfFrames: halfFrames,
      kilograms: 12.5,
      notes: '',
    );

Widget _wrap(Harvest harvest) => MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      locale: const Locale('en'),
      home: Scaffold(
        body: HarvestCard(
          harvest: harvest,
          onEdit: () {},
          onDelete: () {},
        ),
      ),
    );

void main() {
  group('HarvestCard frame count text', () {
    testWidgets('omits half frames when halfFrames is 0', (tester) async {
      await tester.pumpWidget(_wrap(_harvest(frames: 3, halfFrames: 0)));
      expect(find.text('3 frames'), findsOneWidget);
    });

    testWidgets('shows singular frame text for count of 1', (tester) async {
      await tester.pumpWidget(_wrap(_harvest(frames: 1, halfFrames: 0)));
      expect(find.text('1 frame'), findsOneWidget);
    });

    testWidgets('combines frames and half frames when halfFrames > 0',
        (tester) async {
      await tester.pumpWidget(_wrap(_harvest(frames: 3, halfFrames: 2)));
      expect(find.text('3 frames + 2 half frames'), findsOneWidget);
    });

    testWidgets('shows singular half frame text for count of 1',
        (tester) async {
      await tester.pumpWidget(_wrap(_harvest(frames: 2, halfFrames: 1)));
      expect(find.text('2 frames + 1 half frame'), findsOneWidget);
    });
  });
}
