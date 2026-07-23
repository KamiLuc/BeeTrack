import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:app/core/theme/app_layout.dart';
import 'package:app/core/widgets/delete_dialog.dart';
import 'package:app/l10n/app_localizations.dart';

void main() {
  late AppLocalizations l10n;
  setUpAll(() async {
    l10n = await AppLocalizations.delegate.load(const Locale('en'));
  });

  Widget wrap({required bool withPuzzle}) => MaterialApp(
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    supportedLocales: AppLocalizations.supportedLocales,
    locale: const Locale('en'),
    home: Scaffold(
      body: Builder(
        builder: (context) => ElevatedButton(
          onPressed: () => showDeleteDialog(
            context,
            title: 'Delete item',
            warning: 'Are you sure?',
            l10n: l10n,
            withPuzzle: withPuzzle,
          ),
          child: const Text('open'),
        ),
      ),
    ),
  );

  testWidgets(
    'simple confirm dialog content is capped to AppLayout.dialogWidth on a '
    'wide viewport',
    (tester) async {
      tester.view.physicalSize = const Size(1600, 900);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      await tester.pumpWidget(wrap(withPuzzle: false));
      await tester.tap(find.text('open'));
      await tester.pumpAndSettle();

      final context = tester.element(find.byType(AlertDialog));
      final expectedWidth = AppLayout.dialogWidth(context);

      final sizedBox = tester.widget<SizedBox>(
        find.ancestor(
          of: find.text('Are you sure?'),
          matching: find.byType(SizedBox),
        ),
      );

      expect(sizedBox.width, expectedWidth);
    },
  );

  testWidgets(
    'puzzle confirm dialog content is capped to AppLayout.dialogWidth on a '
    'wide viewport',
    (tester) async {
      tester.view.physicalSize = const Size(1600, 900);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      await tester.pumpWidget(wrap(withPuzzle: true));
      await tester.tap(find.text('open'));
      await tester.pumpAndSettle();

      final context = tester.element(find.byType(AlertDialog));
      final expectedWidth = AppLayout.dialogWidth(context);

      final sizedBox = tester.widget<SizedBox>(
        find.ancestor(
          of: find.text('Are you sure?'),
          matching: find.byType(SizedBox),
        ),
      );

      expect(sizedBox.width, expectedWidth);
    },
  );
}
