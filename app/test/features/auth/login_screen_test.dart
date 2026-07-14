import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:app/features/auth/bloc/auth_bloc.dart';
import 'package:app/features/auth/data/auth_repository.dart';
import 'package:app/features/auth/view/login_screen.dart';
import 'package:app/l10n/app_localizations.dart';

class _MockAuthRepository extends Mock implements AuthRepository {}

Widget _wrap(AuthBloc authBloc) => MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      locale: const Locale('en'),
      home: BlocProvider<AuthBloc>.value(
        value: authBloc,
        child: const LoginScreen(),
      ),
    );

void main() {
  group('LoginScreen validation', () {
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
  });
}
