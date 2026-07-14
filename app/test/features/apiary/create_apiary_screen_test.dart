import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:app/core/api/api_client.dart';
import 'package:app/core/storage/token_storage.dart';
import 'package:app/features/apiary/view/create_apiary_screen.dart';
import 'package:app/l10n/app_localizations.dart';

Future<ApiClient> _fakeApiClient() async {
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();
  return ApiClient(storage: TokenStorage(prefs), baseUrl: 'http://test');
}

Widget _wrap(ApiClient apiClient, Widget child) => MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      locale: const Locale('en'),
      home: RepositoryProvider<ApiClient>.value(
        value: apiClient,
        child: child,
      ),
    );

void main() {
  group('CreateApiaryScreen', () {
    testWidgets('shows apiary name required error, not the field label, '
        'when name is cleared', (tester) async {
      final apiClient = await _fakeApiClient();

      await tester.pumpWidget(_wrap(apiClient, const CreateApiaryScreen()));

      await tester.enterText(find.byType(TextFormField).first, '');
      tester.state<FormState>(find.byType(Form)).validate();
      await tester.pump();

      expect(find.text('Apiary name cannot be empty'), findsOneWidget);
      expect(find.text('Apiary'), findsOneWidget);
    });

    testWidgets(
        'shows apiary name required error when name is only whitespace',
        (tester) async {
      final apiClient = await _fakeApiClient();

      await tester.pumpWidget(_wrap(apiClient, const CreateApiaryScreen()));

      await tester.enterText(find.byType(TextFormField).first, '   ');
      tester.state<FormState>(find.byType(Form)).validate();
      await tester.pump();

      expect(find.text('Apiary name cannot be empty'), findsOneWidget);
    });

    testWidgets(
        'does not show apiary name required error for a non-empty name',
        (tester) async {
      final apiClient = await _fakeApiClient();

      await tester.pumpWidget(_wrap(apiClient, const CreateApiaryScreen()));

      await tester.enterText(find.byType(TextFormField).first, 'My Apiary');
      final isValid = tester.state<FormState>(find.byType(Form)).validate();
      await tester.pump();

      expect(isValid, isTrue);
      expect(find.text('Apiary name cannot be empty'), findsNothing);
    });

    testWidgets(
        'shows apiary name required error live on user interaction, '
        'without calling validate() explicitly', (tester) async {
      final apiClient = await _fakeApiClient();

      await tester.pumpWidget(_wrap(apiClient, const CreateApiaryScreen()));

      expect(find.text('Apiary name cannot be empty'), findsNothing);

      await tester.enterText(find.byType(TextFormField).first, 'a');
      await tester.pump();
      await tester.enterText(find.byType(TextFormField).first, '');
      await tester.pump();

      expect(find.text('Apiary name cannot be empty'), findsOneWidget);
    });

    testWidgets('truncates apiary name input at 50 characters',
        (tester) async {
      final apiClient = await _fakeApiClient();

      await tester.pumpWidget(_wrap(apiClient, const CreateApiaryScreen()));

      await tester.enterText(find.byType(TextFormField).first, 'a' * 60);
      await tester.pump();

      final field = tester.widget<TextFormField>(
        find.byType(TextFormField).first,
      );
      expect(field.controller!.text.length, 50);
    });
  });
}
