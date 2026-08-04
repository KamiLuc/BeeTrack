import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:app/features/hive/data/hive_model.dart';
import 'package:app/features/hive/view/update_hive_status_form_screen.dart';
import 'package:app/l10n/app_localizations.dart';

Widget _wrap(Widget child) => MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      locale: const Locale('en'),
      home: child,
    );

final _hive = Hive(
  id: 1,
  apiaryId: 1,
  name: 'Alpha',
  type: 'langstroth',
  active: true,
  queenNeedsReplacement: false,
  readyForHarvest: false,
  needsFood: true,
  boxNeedsAdding: false,
  gridRow: 0,
  gridCol: 0,
  diseases: [
    HiveDisease(id: 1, disease: 'varroa', createdAt: DateTime(2026, 1, 1)),
  ],
);

bool _switchValue(WidgetTester tester, String label) {
  final row = tester.widget<Row>(
    find.ancestor(of: find.text(label), matching: find.byType(Row)).first,
  );
  return row.children.whereType<Switch>().single.value;
}

void main() {
  group('UpdateHiveStatusFormScreen', () {
    testWidgets(
        'pre-fills toggles from toolArguments when present, falling back to '
        "the hive's current values otherwise", (tester) async {
      await tester.pumpWidget(_wrap(UpdateHiveStatusFormScreen(
        hive: _hive,
        toolArguments: const {'ready_for_harvest': true},
        onSaveProposed: (_) async {},
      )));
      await tester.pumpAndSettle();

      // ready_for_harvest: explicitly proposed as true (overrides hive's false).
      expect(_switchValue(tester, 'Ready for harvest'), isTrue);
      // needs_food: not in toolArguments, falls back to hive.needsFood (true).
      expect(_switchValue(tester, 'Needs food'), isTrue);
      // queen_needs_replacement / box_needs_adding: not proposed, fall back
      // to the hive's current false values.
      expect(_switchValue(tester, 'Queen needs replacement'), isFalse);
      expect(_switchValue(tester, 'Needs a box added'), isFalse);

      // diseases: not proposed, falls back to the hive's current disease set.
      final varroaChip =
          tester.widget<FilterChip>(find.widgetWithText(FilterChip, 'Varroa'));
      expect(varroaChip.selected, isTrue);
    });

    testWidgets('saves all 4 flags plus the current disease set on submit',
        (tester) async {
      Map<String, dynamic>? saved;
      await tester.pumpWidget(_wrap(UpdateHiveStatusFormScreen(
        hive: _hive,
        toolArguments: const {'ready_for_harvest': true},
        onSaveProposed: (args) async {
          saved = args;
        },
      )));
      await tester.pumpAndSettle();

      // Deselect the pre-filled varroa disease before saving.
      await tester.tap(find.widgetWithText(FilterChip, 'Varroa'));
      await tester.pumpAndSettle();

      await tester.tap(find.byIcon(Icons.check));
      await tester.pumpAndSettle();

      expect(saved, isNotNull);
      expect(saved!['ready_for_harvest'], isTrue);
      expect(saved!['needs_food'], isTrue);
      expect(saved!['queen_needs_replacement'], isFalse);
      expect(saved!['box_needs_adding'], isFalse);
      expect(saved!['diseases'], isEmpty);
    });

    testWidgets('a failed save shows an error and leaves the form open',
        (tester) async {
      await tester.pumpWidget(_wrap(UpdateHiveStatusFormScreen(
        hive: _hive,
        toolArguments: const {},
        onSaveProposed: (_) async => throw Exception('boom'),
      )));
      await tester.pumpAndSettle();

      await tester.tap(find.byIcon(Icons.check));
      await tester.pumpAndSettle();

      expect(find.byType(SnackBar), findsOneWidget);
      expect(find.text('Update hive status'), findsOneWidget);
    });
  });
}
