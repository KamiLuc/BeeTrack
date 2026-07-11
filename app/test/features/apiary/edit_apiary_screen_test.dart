import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:app/core/api/api_client.dart';
import 'package:app/core/storage/token_storage.dart';
import 'package:app/features/apiary/data/apiary_model.dart';
import 'package:app/features/apiary/view/edit_apiary_screen.dart';
import 'package:app/l10n/app_localizations.dart';

class _FakeEmptyListAdapter implements HttpClientAdapter {
  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    return ResponseBody.fromString(
      jsonEncode(<Map<String, dynamic>>[]),
      200,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}

Future<ApiClient> _fakeApiClient() async {
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();
  final apiClient =
      ApiClient(storage: TokenStorage(prefs), baseUrl: 'http://test');
  apiClient.dio.httpClientAdapter = _FakeEmptyListAdapter();
  return apiClient;
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

const _apiary = Apiary(
  id: 1,
  name: 'Meadow',
  lat: null,
  lng: null,
  gridRows: 3,
  gridCols: 3,
  hiveCount: 0,
  userRole: 'owner',
);

void main() {
  group('EditApiaryScreen', () {
    testWidgets('shows apiary name required error, not the field label, '
        'when name is cleared', (tester) async {
      final apiClient = await _fakeApiClient();

      await tester.pumpWidget(
        _wrap(apiClient, const EditApiaryScreen(apiary: _apiary)),
      );
      await tester.pumpAndSettle();

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

      await tester.pumpWidget(
        _wrap(apiClient, const EditApiaryScreen(apiary: _apiary)),
      );
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(TextFormField).first, '   ');
      tester.state<FormState>(find.byType(Form)).validate();
      await tester.pump();

      expect(find.text('Apiary name cannot be empty'), findsOneWidget);
    });

    testWidgets(
        'does not show apiary name required error for a non-empty name',
        (tester) async {
      final apiClient = await _fakeApiClient();

      await tester.pumpWidget(
        _wrap(apiClient, const EditApiaryScreen(apiary: _apiary)),
      );
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(TextFormField).first, 'New Name');
      final isValid = tester.state<FormState>(find.byType(Form)).validate();
      await tester.pump();

      expect(isValid, isTrue);
      expect(find.text('Apiary name cannot be empty'), findsNothing);
    });

    testWidgets(
        'shows apiary name required error live on user interaction, '
        'without calling validate() explicitly', (tester) async {
      final apiClient = await _fakeApiClient();

      await tester.pumpWidget(
        _wrap(apiClient, const EditApiaryScreen(apiary: _apiary)),
      );
      await tester.pumpAndSettle();

      expect(find.text('Apiary name cannot be empty'), findsNothing);

      await tester.enterText(find.byType(TextFormField).first, '');
      await tester.pump();

      expect(find.text('Apiary name cannot be empty'), findsOneWidget);
    });
  });
}
