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
  queenless: false,
  readyForHarvest: false,
  needsFood: false,
  gridRow: 0,
  gridCol: 0,
);

const _inactive = Hive(
  id: 2,
  apiaryId: 1,
  name: 'Beta',
  type: 'dadant',
  active: false,
  queenless: false,
  readyForHarvest: false,
  needsFood: false,
  gridRow: 0,
  gridCol: 1,
);

const _queenless = Hive(
  id: 4,
  apiaryId: 1,
  name: 'Delta',
  type: 'langstroth',
  active: true,
  queenless: true,
  readyForHarvest: false,
  needsFood: false,
  gridRow: 1,
  gridCol: 0,
);

const _readyForHarvest = Hive(
  id: 5,
  apiaryId: 1,
  name: 'Epsilon',
  type: 'langstroth',
  active: true,
  queenless: false,
  readyForHarvest: true,
  needsFood: false,
  gridRow: 1,
  gridCol: 1,
);

const _needsFood = Hive(
  id: 7,
  apiaryId: 1,
  name: 'Eta',
  type: 'langstroth',
  active: true,
  queenless: false,
  readyForHarvest: false,
  needsFood: true,
  gridRow: 1,
  gridCol: 3,
);

const _inactiveWithStatuses = Hive(
  id: 8,
  apiaryId: 1,
  name: 'Theta',
  type: 'langstroth',
  active: false,
  queenless: true,
  readyForHarvest: true,
  needsFood: true,
  gridRow: 1,
  gridCol: 4,
);

final _withDiseases = Hive(
  id: 6,
  apiaryId: 1,
  name: 'Zeta',
  type: 'langstroth',
  active: true,
  queenless: false,
  readyForHarvest: false,
  needsFood: false,
  gridRow: 1,
  gridCol: 2,
  diseases: const [HiveDisease(id: 1, disease: 'varroa')],
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

    testWidgets('shows Add inspection button when no inspections', (tester) async {
      await tester.pumpWidget(_wrap(
        const HiveDetailScreen(hive: _active, apiaryId: 1),
      ));
      expect(find.text('Add inspection', skipOffstage: false), findsOneWidget);
    });

    testWidgets('does not show View all button when no inspections', (tester) async {
      await tester.pumpWidget(_wrap(
        const HiveDetailScreen(hive: _active, apiaryId: 1),
      ));
      expect(find.text('View all', skipOffstage: false), findsNothing);
    });

    testWidgets('shows Queenless chip only when queenless is true', (tester) async {
      await tester.pumpWidget(_wrap(
        const HiveDetailScreen(hive: _queenless, apiaryId: 1),
      ));
      expect(find.text('Queenless'), findsOneWidget);
    });

    testWidgets('hides Queenless chip when queenless is false', (tester) async {
      await tester.pumpWidget(_wrap(
        const HiveDetailScreen(hive: _active, apiaryId: 1),
      ));
      expect(find.text('Queenless'), findsNothing);
    });

    testWidgets('shows Ready for harvest chip only when readyForHarvest is true', (tester) async {
      await tester.pumpWidget(_wrap(
        const HiveDetailScreen(hive: _readyForHarvest, apiaryId: 1),
      ));
      expect(find.text('Ready for harvest'), findsOneWidget);
    });

    testWidgets('hides Ready for harvest chip when readyForHarvest is false', (tester) async {
      await tester.pumpWidget(_wrap(
        const HiveDetailScreen(hive: _active, apiaryId: 1),
      ));
      expect(find.text('Ready for harvest'), findsNothing);
    });

    testWidgets('shows Needs food chip only when needsFood is true', (tester) async {
      await tester.pumpWidget(_wrap(
        const HiveDetailScreen(hive: _needsFood, apiaryId: 1),
      ));
      expect(find.text('Needs food'), findsOneWidget);
    });

    testWidgets('hides Needs food chip when needsFood is false', (tester) async {
      await tester.pumpWidget(_wrap(
        const HiveDetailScreen(hive: _active, apiaryId: 1),
      ));
      expect(find.text('Needs food'), findsNothing);
    });

    testWidgets('shows Diseases section with disease chip when diseases present', (tester) async {
      await tester.pumpWidget(_wrap(
        HiveDetailScreen(hive: _withDiseases, apiaryId: 1),
      ));
      expect(find.text('Diseases', skipOffstage: false), findsOneWidget);
      expect(find.text('Varroa', skipOffstage: false), findsOneWidget);
    });

    testWidgets('hides Diseases section when no diseases', (tester) async {
      await tester.pumpWidget(_wrap(
        const HiveDetailScreen(hive: _active, apiaryId: 1),
      ));
      expect(find.text('Diseases', skipOffstage: false), findsNothing);
    });

    testWidgets('hides status chips for inactive hive even when flags are true',
        (tester) async {
      await tester.pumpWidget(_wrap(
        const HiveDetailScreen(hive: _inactiveWithStatuses, apiaryId: 1),
      ));
      expect(find.text('Queenless'), findsNothing);
      expect(find.text('Ready for harvest'), findsNothing);
      expect(find.text('Needs food'), findsNothing);
    });

    testWidgets('hides Add inspection button for inactive hive', (tester) async {
      await tester.pumpWidget(_wrap(
        const HiveDetailScreen(hive: _inactive, apiaryId: 1),
      ));
      expect(find.text('Add inspection', skipOffstage: false), findsNothing);
    });

    testWidgets('hides Log treatment/feeding/harvest buttons for inactive hive',
        (tester) async {
      await tester.pumpWidget(_wrap(
        const HiveDetailScreen(hive: _inactive, apiaryId: 1),
      ));
      expect(find.text('Log treatment', skipOffstage: false), findsNothing);
      expect(find.text('Log feeding', skipOffstage: false), findsNothing);
      expect(find.text('Log harvest', skipOffstage: false), findsNothing);
    });
  });
}
