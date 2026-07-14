import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:app/features/hive/view/hive_form_widgets.dart';
import 'package:app/l10n/app_localizations.dart';

Widget _wrap({
  required GlobalKey<FormState> formKey,
  required TextEditingController controller,
  Set<String> existingNames = const {},
}) =>
    MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      locale: const Locale('en'),
      home: Scaffold(
        body: Form(
          key: formKey,
          child: HiveNameField(
            controller: controller,
            existingNames: existingNames,
          ),
        ),
      ),
    );

Widget _wrapTypeDropdown({
  required GlobalKey<FormState> formKey,
  required String value,
}) =>
    MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      locale: const Locale('en'),
      home: Scaffold(
        body: Form(
          key: formKey,
          child: HiveTypeDropdown(
            value: value,
            onChanged: (_) {},
          ),
        ),
      ),
    );

void main() {
  group('HiveNameField', () {
    testWidgets('validate shows Required when name is empty', (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        controller: TextEditingController(),
      ));

      formKey.currentState!.validate();
      await tester.pump();

      expect(find.text('Required'), findsOneWidget);
    });

    testWidgets('validate shows Required when name is only whitespace',
        (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        controller: TextEditingController(text: '   '),
      ));

      formKey.currentState!.validate();
      await tester.pump();

      expect(find.text('Required'), findsOneWidget);
    });

    testWidgets('validate passes for unique name', (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        controller: TextEditingController(text: 'Alpha'),
        existingNames: const {'beta', 'gamma'},
      ));

      expect(formKey.currentState!.validate(), isTrue);
    });

    testWidgets('validate shows duplicate name error for exact match',
        (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        controller: TextEditingController(text: 'Beta'),
        existingNames: const {'beta'},
      ));

      formKey.currentState!.validate();
      await tester.pump();

      expect(
        find.text('A hive with this name already exists in this apiary'),
        findsOneWidget,
      );
    });

    testWidgets('validate is case-insensitive for duplicate detection',
        (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        controller: TextEditingController(text: 'BETA'),
        existingNames: const {'beta'},
      ));

      expect(formKey.currentState!.validate(), isFalse);
      await tester.pump();
      expect(
        find.text('A hive with this name already exists in this apiary'),
        findsOneWidget,
      );
    });

    testWidgets('validate trims whitespace before duplicate comparison',
        (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        controller: TextEditingController(text: '  Beta  '),
        existingNames: const {'beta'},
      ));

      expect(formKey.currentState!.validate(), isFalse);
    });

    testWidgets('empty check takes priority over duplicate check',
        (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        controller: TextEditingController(),
        existingNames: const {''},
      ));

      formKey.currentState!.validate();
      await tester.pump();

      expect(find.text('Required'), findsOneWidget);
      expect(
        find.text('A hive with this name already exists in this apiary'),
        findsNothing,
      );
    });

    testWidgets('validate passes when existingNames is empty',
        (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        controller: TextEditingController(text: 'Anything'),
      ));

      expect(formKey.currentState!.validate(), isTrue);
    });

    testWidgets('validate shows too-long error for a name over 50 characters',
        (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        controller: TextEditingController(text: 'a' * 51),
      ));

      expect(formKey.currentState!.validate(), isFalse);
      await tester.pump();
      expect(
        find.text('Hive name must be at most 50 characters'),
        findsOneWidget,
      );
    });

    testWidgets('validate passes for a name at exactly 50 characters',
        (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        controller: TextEditingController(text: 'a' * 50),
      ));

      expect(formKey.currentState!.validate(), isTrue);
    });
  });

  group('HiveTypeDropdown', () {
    testWidgets('validate shows too-long error for a type over 50 characters',
        (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrapTypeDropdown(
        formKey: formKey,
        value: 'a' * 51,
      ));

      expect(formKey.currentState!.validate(), isFalse);
      await tester.pump();
      expect(
        find.text('Hive type must be at most 50 characters'),
        findsOneWidget,
      );
    });

    testWidgets('validate passes for the preset type value', (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrapTypeDropdown(
        formKey: formKey,
        value: 'dadant',
      ));

      expect(formKey.currentState!.validate(), isTrue);
    });
  });
}
