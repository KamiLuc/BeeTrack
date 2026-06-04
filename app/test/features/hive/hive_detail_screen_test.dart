import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:app/features/hive/data/hive_model.dart';
import 'package:app/features/hive/view/hive_detail_screen.dart';
import 'package:app/l10n/app_localizations.dart';

const _active = Hive(
  id: 1,
  apiaryId: 1,
  name: 'Alpha',
  type: 'langstroth',
  active: true,
  gridRow: 0,
  gridCol: 0,
);

const _inactive = Hive(
  id: 2,
  apiaryId: 1,
  name: 'Beta',
  type: 'dadant',
  active: false,
  gridRow: 0,
  gridCol: 1,
);

Widget _wrap(Widget child) => MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      locale: const Locale('en'),
      home: child,
    );

void main() {
  group('HiveDetailScreen', () {
    testWidgets('shows hive name in AppBar', (tester) async {
      await tester.pumpWidget(_wrap(
        const HiveDetailScreen(hive: _active, apiaryId: 1),
      ));
      expect(find.text('Alpha'), findsOneWidget);
    });

    testWidgets('shows mapped type label', (tester) async {
      await tester.pumpWidget(_wrap(
        const HiveDetailScreen(hive: _active, apiaryId: 1),
      ));
      expect(find.text('Langstroth'), findsOneWidget);
    });

    testWidgets('shows Dadant type label', (tester) async {
      await tester.pumpWidget(_wrap(
        const HiveDetailScreen(hive: _inactive, apiaryId: 1),
      ));
      expect(find.text('Dadant'), findsOneWidget);
    });

    testWidgets('shows Active status for active hive', (tester) async {
      await tester.pumpWidget(_wrap(
        const HiveDetailScreen(hive: _active, apiaryId: 1),
      ));
      expect(find.text('Active'), findsOneWidget);
      expect(find.text('Inactive'), findsNothing);
    });

    testWidgets('shows Inactive status for inactive hive', (tester) async {
      await tester.pumpWidget(_wrap(
        const HiveDetailScreen(hive: _inactive, apiaryId: 1),
      ));
      expect(find.text('Inactive'), findsOneWidget);
      expect(find.text('Active'), findsNothing);
    });

    testWidgets('shows all three section headings', (tester) async {
      await tester.pumpWidget(_wrap(
        const HiveDetailScreen(hive: _active, apiaryId: 1),
      ));
      expect(find.text('Inspections', skipOffstage: false), findsOneWidget);
      expect(find.text('Treatments', skipOffstage: false), findsOneWidget);
      expect(find.text('Harvests', skipOffstage: false), findsOneWidget);
    });

    testWidgets('shows empty state text for each section', (tester) async {
      await tester.pumpWidget(_wrap(
        const HiveDetailScreen(hive: _active, apiaryId: 1),
      ));
      expect(
        find.text('No inspections yet', skipOffstage: false),
        findsOneWidget,
      );
      expect(
        find.text('No active treatments', skipOffstage: false),
        findsOneWidget,
      );
      expect(
        find.text('No harvests yet', skipOffstage: false),
        findsOneWidget,
      );
    });
  });
}
