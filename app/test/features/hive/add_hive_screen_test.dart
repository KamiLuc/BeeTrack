import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:app/features/hive/view/add_hive_screen.dart';
import 'package:app/l10n/app_localizations.dart';

Widget _wrap(Widget child) => MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      locale: const Locale('en'),
      home: child,
    );

const _duplicateNameError =
    'A hive with this name already exists in this apiary';

void main() {
  group('AddHiveScreen', () {
    testWidgets('shows duplicate name error for case-insensitive match '
        'and blocks submission', (tester) async {
      await tester.pumpWidget(_wrap(AddHiveScreen(
        apiaryId: 1,
        gridRow: 0,
        gridCol: 0,
        defaultName: 'New hive',
        existingNames: const {'alpha'},
      )));

      await tester.enterText(find.byType(TextFormField).first, 'ALPHA');
      await tester.pump();
      await tester.tap(find.text('Save'));
      await tester.pump();

      expect(find.text(_duplicateNameError), findsOneWidget);
      expect(find.text('Add hive'), findsOneWidget);
    });

    testWidgets('does not show duplicate name error for a unique name',
        (tester) async {
      await tester.pumpWidget(_wrap(AddHiveScreen(
        apiaryId: 1,
        gridRow: 0,
        gridCol: 0,
        defaultName: 'New hive',
        existingNames: const {'alpha'},
      )));

      await tester.enterText(find.byType(TextFormField).first, 'Beta');
      await tester.pump();

      expect(find.text(_duplicateNameError), findsNothing);
    });

    testWidgets('shows Required error when name is cleared', (tester) async {
      await tester.pumpWidget(_wrap(const AddHiveScreen(
        apiaryId: 1,
        gridRow: 0,
        gridCol: 0,
        defaultName: 'New hive',
      )));

      await tester.enterText(find.byType(TextFormField).first, '');
      await tester.pump();

      expect(find.text('Required'), findsOneWidget);
    });

    testWidgets('defaults to no existing names when not provided',
        (tester) async {
      await tester.pumpWidget(_wrap(const AddHiveScreen(
        apiaryId: 1,
        gridRow: 0,
        gridCol: 0,
        defaultName: 'New hive',
      )));

      await tester.enterText(find.byType(TextFormField).first, 'Anything');
      await tester.pump();

      expect(find.text(_duplicateNameError), findsNothing);
    });
  });
}
