import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:app/features/harvest/view/harvest_form_fields.dart';
import 'package:app/l10n/app_localizations.dart';

Widget _wrap({
  required GlobalKey<FormState> formKey,
  required DateTime harvestedAt,
  required TextEditingController framesController,
  required TextEditingController halfFramesController,
  required TextEditingController kilogramsController,
  TextEditingController? notesController,
  VoidCallback? onDateTap,
}) =>
    MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      locale: const Locale('en'),
      home: Scaffold(
        body: SingleChildScrollView(
          child: HarvestFormFields(
            formKey: formKey,
            harvestedAt: harvestedAt,
            framesController: framesController,
            halfFramesController: halfFramesController,
            kilogramsController: kilogramsController,
            notesController: notesController ?? TextEditingController(),
            onDateTap: onDateTap ?? () {},
          ),
        ),
      ),
    );

void main() {
  final date = DateTime(2026, 6, 8);

  group('HarvestFormFields', () {
    testWidgets('renders date, frames, half frames, kilograms, and note fields',
        (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        harvestedAt: date,
        framesController: TextEditingController(text: '0'),
        halfFramesController: TextEditingController(text: '0'),
        kilogramsController: TextEditingController(),
      ));

      expect(find.text('Harvest date'), findsOneWidget);
      expect(find.text('Frames'), findsOneWidget);
      expect(find.text('Half frames'), findsOneWidget);
      expect(find.text('Kilograms (kg)'), findsOneWidget);
      expect(find.text('Note'), findsOneWidget);
    });

    testWidgets('displays formatted harvestedAt date', (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        harvestedAt: date,
        framesController: TextEditingController(text: '0'),
        halfFramesController: TextEditingController(text: '0'),
        kilogramsController: TextEditingController(),
      ));

      expect(find.textContaining('Jun 8, 2026'), findsOneWidget);
    });

    testWidgets('calls onDateTap when date field is tapped', (tester) async {
      var tapped = false;
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        harvestedAt: date,
        framesController: TextEditingController(text: '0'),
        halfFramesController: TextEditingController(text: '0'),
        kilogramsController: TextEditingController(),
        onDateTap: () => tapped = true,
      ));

      await tester.tap(find.textContaining('Jun 8, 2026'));
      expect(tapped, isTrue);
    });

    testWidgets('validate shows error when both frames and half frames are 0',
        (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        harvestedAt: date,
        framesController: TextEditingController(text: '0'),
        halfFramesController: TextEditingController(text: '0'),
        kilogramsController: TextEditingController(text: '5.00'),
      ));

      formKey.currentState!.validate();
      await tester.pump();

      expect(find.text('At least one frame is required'), findsOneWidget);
    });

    testWidgets('validate passes when only half frames > 0', (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        harvestedAt: date,
        framesController: TextEditingController(text: '0'),
        halfFramesController: TextEditingController(text: '2'),
        kilogramsController: TextEditingController(text: '5.00'),
      ));

      expect(formKey.currentState!.validate(), isTrue);
    });

    testWidgets('validate shows error when kilograms is empty', (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        harvestedAt: date,
        framesController: TextEditingController(text: '5'),
        halfFramesController: TextEditingController(text: '2'),
        kilogramsController: TextEditingController(),
      ));

      formKey.currentState!.validate();
      await tester.pump();

      expect(find.text('Kilograms is required'), findsOneWidget);
    });

    testWidgets('validate passes when kilograms is filled', (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        harvestedAt: date,
        framesController: TextEditingController(text: '5'),
        halfFramesController: TextEditingController(text: '2'),
        kilogramsController: TextEditingController(text: '12.50'),
      ));

      expect(formKey.currentState!.validate(), isTrue);
    });

    testWidgets('truncates frames input at 2 characters', (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        harvestedAt: date,
        framesController: TextEditingController(),
        halfFramesController: TextEditingController(text: '0'),
        kilogramsController: TextEditingController(text: '5.00'),
      ));

      await tester.enterText(find.widgetWithText(TextFormField, 'Frames'), '1' * 10);
      await tester.pump();

      final field = tester.widget<TextFormField>(
        find.widgetWithText(TextFormField, 'Frames'),
      );
      expect(field.controller!.text.length, 2);
    });

    testWidgets('truncates half frames input at 2 characters',
        (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        harvestedAt: date,
        framesController: TextEditingController(text: '0'),
        halfFramesController: TextEditingController(),
        kilogramsController: TextEditingController(text: '5.00'),
      ));

      await tester.enterText(
        find.widgetWithText(TextFormField, 'Half frames'),
        '1' * 10,
      );
      await tester.pump();

      final field = tester.widget<TextFormField>(
        find.widgetWithText(TextFormField, 'Half frames'),
      );
      expect(field.controller!.text.length, 2);
    });

    testWidgets(
        'validate shows error when kilograms text is over 50 characters',
        (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        harvestedAt: date,
        framesController: TextEditingController(text: '5'),
        halfFramesController: TextEditingController(text: '2'),
        kilogramsController: TextEditingController(text: '1' * 51),
      ));

      expect(formKey.currentState!.validate(), isFalse);
      await tester.pump();
      expect(
        find.text('Kilograms (kg) must be at most 50 characters'),
        findsOneWidget,
      );
    });

    testWidgets('validate shows error when kilograms is over 1000',
        (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        harvestedAt: date,
        framesController: TextEditingController(text: '5'),
        halfFramesController: TextEditingController(text: '2'),
        kilogramsController: TextEditingController(text: '1000.01'),
      ));

      expect(formKey.currentState!.validate(), isFalse);
      await tester.pump();
      expect(
        find.text('Kilograms (kg) must be at most 1000'),
        findsOneWidget,
      );
    });

    testWidgets('validate passes when kilograms is exactly 1000',
        (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        harvestedAt: date,
        framesController: TextEditingController(text: '5'),
        halfFramesController: TextEditingController(text: '2'),
        kilogramsController: TextEditingController(text: '1000'),
      ));

      expect(formKey.currentState!.validate(), isTrue);
    });

    testWidgets('validate shows error when notes are over 5000 characters',
        (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        harvestedAt: date,
        framesController: TextEditingController(text: '5'),
        halfFramesController: TextEditingController(text: '2'),
        kilogramsController: TextEditingController(text: '12.50'),
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
        harvestedAt: date,
        framesController: TextEditingController(text: '8'),
        halfFramesController: TextEditingController(text: '3'),
        kilogramsController: TextEditingController(text: '24.75'),
      ));

      expect(find.text('8'), findsOneWidget);
      expect(find.text('3'), findsOneWidget);
      expect(find.text('24.75'), findsOneWidget);
    });
  });
}
