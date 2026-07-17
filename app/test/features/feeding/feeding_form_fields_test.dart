import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:app/features/feeding/view/feeding_form_fields.dart';
import 'package:app/l10n/app_localizations.dart';

Widget _wrap({
  required GlobalKey<FormState> formKey,
  required DateTime fedAt,
  required TextEditingController feedTypeController,
  required TextEditingController amountController,
  required TextEditingController notesController,
  List<String> feedTypeOptions = const [],
  VoidCallback? onDateTap,
}) =>
    MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      locale: const Locale('en'),
      home: Scaffold(
        body: SingleChildScrollView(
          child: FeedingFormFields(
            formKey: formKey,
            fedAt: fedAt,
            feedTypeController: feedTypeController,
            amountController: amountController,
            notesController: notesController,
            feedTypeOptions: feedTypeOptions,
            onDateTap: onDateTap ?? () {},
          ),
        ),
      ),
    );

void main() {
  final date = DateTime(2026, 6, 8);

  group('FeedingFormFields', () {
    testWidgets('renders date, feed type, amount, and notes fields',
        (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        fedAt: date,
        feedTypeController: TextEditingController(),
        amountController: TextEditingController(text: '1L'),
        notesController: TextEditingController(),
      ));

      expect(find.text('Feeding date'), findsOneWidget);
      expect(find.text('Feed'), findsOneWidget);
      expect(find.text('Amount'), findsOneWidget);
      expect(find.text('Note'), findsOneWidget);
    });

    testWidgets('displays formatted fedAt date with time', (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        fedAt: date,
        feedTypeController: TextEditingController(),
        amountController: TextEditingController(text: '1L'),
        notesController: TextEditingController(),
      ));

      expect(find.textContaining('Jun 8, 2026'), findsOneWidget);
    });

    testWidgets('calls onDateTap when date field is tapped', (tester) async {
      var tapped = false;
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        fedAt: date,
        feedTypeController: TextEditingController(),
        amountController: TextEditingController(text: '1L'),
        notesController: TextEditingController(),
        onDateTap: () => tapped = true,
      ));

      await tester.tap(find.textContaining('Jun 8, 2026'));
      expect(tapped, isTrue);
    });

    testWidgets('validate shows error when feed type is empty', (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        fedAt: date,
        feedTypeController: TextEditingController(),
        amountController: TextEditingController(text: '1L'),
        notesController: TextEditingController(),
      ));

      formKey.currentState!.validate();
      await tester.pump();

      expect(find.text('Feed type is required'), findsOneWidget);
    });

    testWidgets('validate shows error when amount is empty', (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        fedAt: date,
        feedTypeController: TextEditingController(text: 'Sugar syrup'),
        amountController: TextEditingController(),
        notesController: TextEditingController(),
      ));

      formKey.currentState!.validate();
      await tester.pump();

      expect(find.text('Amount is required'), findsOneWidget);
    });

    testWidgets('validate passes when feed type and amount are filled',
        (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        fedAt: date,
        feedTypeController: TextEditingController(text: 'Sugar syrup'),
        amountController: TextEditingController(text: '1L'),
        notesController: TextEditingController(),
      ));

      expect(formKey.currentState!.validate(), isTrue);
    });

    testWidgets('validate shows error when feed type is over 50 characters',
        (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        fedAt: date,
        feedTypeController: TextEditingController(text: 'a' * 51),
        amountController: TextEditingController(text: '1L'),
        notesController: TextEditingController(),
      ));

      expect(formKey.currentState!.validate(), isFalse);
      await tester.pump();
      expect(
        find.text('Feed must be at most 50 characters'),
        findsOneWidget,
      );
    });

    testWidgets('validate shows error when amount is over 20 characters',
        (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        fedAt: date,
        feedTypeController: TextEditingController(text: 'Sugar syrup'),
        amountController: TextEditingController(text: '1' * 21),
        notesController: TextEditingController(),
      ));

      expect(formKey.currentState!.validate(), isFalse);
      await tester.pump();
      expect(
        find.text('Amount must be at most 20 characters'),
        findsOneWidget,
      );
    });

    testWidgets('validate shows error when notes are over 5000 characters',
        (tester) async {
      final formKey = GlobalKey<FormState>();
      await tester.pumpWidget(_wrap(
        formKey: formKey,
        fedAt: date,
        feedTypeController: TextEditingController(text: 'Sugar syrup'),
        amountController: TextEditingController(text: '1L'),
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
        fedAt: date,
        feedTypeController: TextEditingController(text: 'Pollen patty'),
        amountController: TextEditingController(text: '2'),
        notesController: TextEditingController(text: 'placed above frames'),
      ));

      expect(find.text('2'), findsOneWidget);
      expect(find.text('placed above frames'), findsOneWidget);
    });
  });
}
