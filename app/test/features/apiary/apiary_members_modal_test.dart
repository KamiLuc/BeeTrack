import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:app/core/api/api_client.dart';
import 'package:app/core/storage/token_storage.dart';
import 'package:app/features/apiary/view/apiary_members_modal.dart';
import 'package:app/l10n/app_localizations.dart';

Future<ApiClient> _fakeApiClient() async {
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();
  return ApiClient(storage: TokenStorage(prefs), baseUrl: 'http://test');
}

Widget _wrap(ApiClient apiClient) => MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      locale: const Locale('en'),
      home: Scaffold(
        body: ApiaryMembersModal(apiaryId: 1, apiClient: apiClient),
      ),
    );

void main() {
  group('ApiaryMembersModal invite email', () {
    testWidgets('truncates email input at 150 characters', (tester) async {
      final apiClient = await _fakeApiClient();
      await tester.pumpWidget(_wrap(apiClient));
      await tester.pump();

      await tester.enterText(
        find.byType(TextField),
        '${'a' * 150}@example.com',
      );
      await tester.pump();

      final field = tester.widget<TextField>(find.byType(TextField));
      expect(field.controller!.text.length, 150);
    });
  });
}
