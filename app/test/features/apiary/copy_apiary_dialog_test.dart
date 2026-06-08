import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:app/features/apiary/view/apiaries_screen.dart';
import 'package:app/l10n/app_localizations.dart';

Widget _wrap(Widget child) => MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      locale: const Locale('en'),
      home: child,
    );

Widget _dialogHost({String defaultName = 'My Apiary (copy)'}) => Builder(
      builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return Scaffold(
          body: ElevatedButton(
            onPressed: () => showDialog<String>(
              context: context,
              builder: (_) => CopyApiaryDialog(defaultName: defaultName, l10n: l10n),
            ),
            child: const Text('Open'),
          ),
        );
      },
    );

Future<void> _openDialog(WidgetTester tester, {String defaultName = 'My Apiary (copy)'}) async {
  await tester.pumpWidget(_wrap(_dialogHost(defaultName: defaultName)));
  await tester.tap(find.text('Open'));
  await tester.pumpAndSettle();
}

void main() {
  group('CopyApiaryDialog', () {
    testWidgets('shows dialog title', (tester) async {
      await _openDialog(tester);
      expect(find.text('Copy apiary'), findsOneWidget);
    });

    testWidgets('pre-fills text field with default name', (tester) async {
      await _openDialog(tester, defaultName: 'Meadow (copy)');
      expect(find.widgetWithText(TextField, 'Meadow (copy)'), findsOneWidget);
    });

    testWidgets('shows New name label on text field', (tester) async {
      await _openDialog(tester);
      expect(find.text('New name'), findsOneWidget);
    });

    testWidgets('Cancel closes dialog without result', (tester) async {
      await _openDialog(tester);
      await tester.tap(find.text('Cancel'));
      await tester.pumpAndSettle();
      expect(find.text('Copy apiary'), findsNothing);
    });

    testWidgets('Confirm closes dialog with default name', (tester) async {
      String? result;
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return Scaffold(
          body: ElevatedButton(
            onPressed: () async {
              result = await showDialog<String>(
                context: context,
                builder: (_) =>
                    CopyApiaryDialog(defaultName: 'Forest (copy)', l10n: l10n),
              );
            },
            child: const Text('Open'),
          ),
        );
      })));
      await tester.tap(find.text('Open'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Confirm'));
      await tester.pumpAndSettle();
      expect(result, 'Forest (copy)');
    });

    testWidgets('Confirm returns edited name when text is changed', (tester) async {
      String? result;
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return Scaffold(
          body: ElevatedButton(
            onPressed: () async {
              result = await showDialog<String>(
                context: context,
                builder: (_) =>
                    CopyApiaryDialog(defaultName: 'Old name', l10n: l10n),
              );
            },
            child: const Text('Open'),
          ),
        );
      })));
      await tester.tap(find.text('Open'));
      await tester.pumpAndSettle();
      await tester.enterText(find.byType(TextField), 'New name');
      await tester.tap(find.text('Confirm'));
      await tester.pumpAndSettle();
      expect(result, 'New name');
    });

    testWidgets('Confirm does nothing when field is cleared', (tester) async {
      await _openDialog(tester);
      await tester.enterText(find.byType(TextField), '   ');
      await tester.tap(find.text('Confirm'));
      await tester.pumpAndSettle();
      expect(find.text('Copy apiary'), findsOneWidget);
    });

    testWidgets('submitting via keyboard confirms with current text', (tester) async {
      String? result;
      await tester.pumpWidget(_wrap(Builder(builder: (context) {
        final l10n = AppLocalizations.of(context)!;
        return Scaffold(
          body: ElevatedButton(
            onPressed: () async {
              result = await showDialog<String>(
                context: context,
                builder: (_) =>
                    CopyApiaryDialog(defaultName: 'Hill apiary (copy)', l10n: l10n),
              );
            },
            child: const Text('Open'),
          ),
        );
      })));
      await tester.tap(find.text('Open'));
      await tester.pumpAndSettle();
      await tester.testTextInput.receiveAction(TextInputAction.done);
      await tester.pumpAndSettle();
      expect(result, 'Hill apiary (copy)');
    });
  });
}
