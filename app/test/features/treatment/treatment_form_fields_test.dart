import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:app/features/treatment/view/treatment_form_fields.dart';
import 'package:app/l10n/app_localizations.dart';

Widget _wrap({
  required GlobalKey<FormState> formKey,
  required DateTime treatedAt,
  required TextEditingController medicineController,
  required TextEditingController doseController,
  required TextEditingController notesController,
  List<String> medicineOptions = const [],
  VoidCallback? onDateTap,
}) =>
    MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      locale: const Locale('en'),
      home: Scaffold(
        body: SingleChildScrollView(
          child: TreatmentFormFields(
            formKey: formKey,
            treatedAt: treatedAt,
            medicineController: medicineController,
            doseController: doseController,
            notesController: notesController,
            medicineOptions: medicineOptions,
            onDateTap: onDateTap ?? () {},
          ),
        ),
      ),
    );

void main() {
  final date = DateTime(2026, 6, 8);

  group('TreatmentFormFields', () {
    testWidgets('renders date, medicine, dose, and notes fields', (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        treatedAt: date,
        medicineController: TextEditingController(),
        doseController: TextEditingController(text: '1'),
        notesController: TextEditingController(),
      ));

      expect(find.text('Treatment date'), findsOneWidget);
      expect(find.text('Medicine'), findsOneWidget);
      expect(find.text('Dose'), findsOneWidget);
      expect(find.text('Note'), findsOneWidget);
    });

    testWidgets('displays formatted treatedAt date with time', (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        treatedAt: date,
        medicineController: TextEditingController(),
        doseController: TextEditingController(text: '1'),
        notesController: TextEditingController(),
      ));

      expect(find.textContaining('Jun 8, 2026'), findsOneWidget);
    });

    testWidgets('calls onDateTap when date field is tapped', (tester) async {
      var tapped = false;
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        treatedAt: date,
        medicineController: TextEditingController(),
        doseController: TextEditingController(text: '1'),
        notesController: TextEditingController(),
        onDateTap: () => tapped = true,
      ));

      await tester.tap(find.textContaining('Jun 8, 2026'));
      expect(tapped, isTrue);
    });

    testWidgets('validate shows error when medicine is empty', (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        treatedAt: date,
        medicineController: TextEditingController(),
        doseController: TextEditingController(text: '1'),
        notesController: TextEditingController(),
      ));

      formKey.currentState!.validate();
      await tester.pump();

      expect(find.text('Medicine name is required'), findsOneWidget);
    });

    testWidgets('validate shows error when dose is empty', (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        treatedAt: date,
        medicineController: TextEditingController(text: 'Apivar'),
        doseController: TextEditingController(),
        notesController: TextEditingController(),
      ));

      formKey.currentState!.validate();
      await tester.pump();

      expect(find.text('Dose is required'), findsOneWidget);
    });

    testWidgets('validate passes when medicine and dose are filled', (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        treatedAt: date,
        medicineController: TextEditingController(text: 'Apivar'),
        doseController: TextEditingController(text: '2'),
        notesController: TextEditingController(),
      ));

      expect(formKey.currentState!.validate(), isTrue);
    });

    testWidgets('validate shows error when medicine is over 50 characters',
        (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        treatedAt: date,
        medicineController: TextEditingController(text: 'a' * 51),
        doseController: TextEditingController(text: '2'),
        notesController: TextEditingController(),
      ));

      expect(formKey.currentState!.validate(), isFalse);
      await tester.pump();
      expect(
        find.text('Medicine must be at most 50 characters'),
        findsOneWidget,
      );
    });

    testWidgets('validate shows error when dose is over 20 characters',
        (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        treatedAt: date,
        medicineController: TextEditingController(text: 'Apivar'),
        doseController: TextEditingController(text: '1' * 21),
        notesController: TextEditingController(),
      ));

      expect(formKey.currentState!.validate(), isFalse);
      await tester.pump();
      expect(
        find.text('Dose must be at most 20 characters'),
        findsOneWidget,
      );
    });

    testWidgets('validate shows error when notes are over 5000 characters',
        (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        treatedAt: date,
        medicineController: TextEditingController(text: 'Apivar'),
        doseController: TextEditingController(text: '2'),
        notesController: TextEditingController(text: 'a' * 5001),
      ));

      expect(formKey.currentState!.validate(), isFalse);
      await tester.pump();
      expect(
        find.text('Note must be at most 5000 characters'),
        findsOneWidget,
      );
    });

    testWidgets('pre-fills fields from controllers', (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        treatedAt: date,
        medicineController: TextEditingController(text: 'Apiwarol'),
        doseController: TextEditingController(text: '3'),
        notesController: TextEditingController(text: 'applied evenly'),
      ));

      expect(find.text('3'), findsOneWidget);
      expect(find.text('applied evenly'), findsOneWidget);
    });
  });
}
