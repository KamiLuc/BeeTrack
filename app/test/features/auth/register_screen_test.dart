import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:app/features/auth/bloc/auth_bloc.dart';
import 'package:app/features/auth/data/auth_repository.dart';
import 'package:app/features/auth/view/register_screen.dart';
import 'package:app/l10n/app_localizations.dart';

class _MockAuthRepository extends Mock implements AuthRepository {}

Widget _wrap(AuthBloc authBloc) => MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      locale: const Locale('en'),
      home: BlocProvider<AuthBloc>.value(
        value: authBloc,
        child: const RegisterScreen(),
      ),
    );

void main() {
  group('RegisterScreen validation', () {
    testWidgets('truncates email input at 150 characters', (tester) async {
      final authBloc = AuthBloc(auth: _MockAuthRepository());
      await tester.pumpWidget(_wrap(authBloc));

      await tester.enterText(
        find.widgetWithText(TextFormField, 'Email'),
        '${'a' * 150}@example.com',
      );
      await tester.pump();

      final field = tester.widget<TextFormField>(
        find.widgetWithText(TextFormField, 'Email'),
      );
      expect(field.controller!.text.length, 150);
    });

    testWidgets('truncates password input at 72 characters', (tester) async {
      final authBloc = AuthBloc(auth: _MockAuthRepository());
      await tester.pumpWidget(_wrap(authBloc));

      await tester.enterText(
        find.widgetWithText(TextFormField, 'Password'),
        'a' * 80,
      );
      await tester.pump();

      final field = tester.widget<TextFormField>(
        find.widgetWithText(TextFormField, 'Password'),
      );
      expect(field.controller!.text.length, 72);
    });

    testWidgets('shows mismatch error and does not submit when passwords differ',
        (tester) async {
      final mockRepo = _MockAuthRepository();
      final authBloc = AuthBloc(auth: mockRepo);
      await tester.pumpWidget(_wrap(authBloc));

      await tester.enterText(
        find.widgetWithText(TextFormField, 'Email'),
        'user@example.com',
      );
      await tester.enterText(
        find.widgetWithText(TextFormField, 'Password'),
        'password1',
      );
      await tester.enterText(
        find.widgetWithText(TextFormField, 'Confirm password'),
        'password2',
      );
      await tester.tap(find.widgetWithText(ElevatedButton, 'Register'));
      await tester.pump();

      expect(find.text('Passwords do not match'), findsOneWidget);
      verifyNever(
        () => mockRepo.register(
          email: any(named: 'email'),
          lang: any(named: 'lang'),
          name: any(named: 'name'),
          password: any(named: 'password'),
        ),
      );
    });

    testWidgets('does not show mismatch error when passwords match',
        (tester) async {
      final mockRepo = _MockAuthRepository();
      when(
        () => mockRepo.register(
          email: any(named: 'email'),
          lang: any(named: 'lang'),
          name: any(named: 'name'),
          password: any(named: 'password'),
        ),
      ).thenAnswer((_) async {});
      final authBloc = AuthBloc(auth: mockRepo);
      await tester.pumpWidget(_wrap(authBloc));

      await tester.enterText(
        find.widgetWithText(TextFormField, 'Email'),
        'user@example.com',
      );
      await tester.enterText(
        find.widgetWithText(TextFormField, 'Password'),
        'password1',
      );
      await tester.enterText(
        find.widgetWithText(TextFormField, 'Confirm password'),
        'password1',
      );
      await tester.tap(find.widgetWithText(ElevatedButton, 'Register'));
      await tester.pump();

      expect(find.text('Passwords do not match'), findsNothing);
    });
  });
}
