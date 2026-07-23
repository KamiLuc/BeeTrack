import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:app/core/api/api_client.dart';
import 'package:app/core/storage/token_storage.dart';
import 'package:app/features/hive/data/hive_model.dart';
import 'package:app/features/hive/view/edit_hive_screen.dart';
import 'package:app/l10n/app_localizations.dart';

class _FakeHiveListAdapter implements HttpClientAdapter {
  _FakeHiveListAdapter(this.hivesJson);

  final List<Map<String, dynamic>> hivesJson;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    return ResponseBody.fromString(
      jsonEncode(hivesJson),
      200,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}

Future<ApiClient> _fakeApiClient(List<Map<String, dynamic>> hivesJson) async {
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();
  final apiClient = ApiClient(storage: TokenStorage(prefs), baseUrl: 'http://test');
  apiClient.dio.httpClientAdapter = _FakeHiveListAdapter(hivesJson);
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

const _hive = Hive(
  id: 1,
  apiaryId: 1,
  name: 'Alpha',
  type: 'langstroth',
  active: true,
  queenless: false,
  readyForHarvest: false,
  needsFood: false,
  gridRow: 0,
  gridCol: 0,
);

const _duplicateNameError =
    'A hive with this name already exists in this apiary';

void main() {
  group('EditHiveScreen', () {
    testWidgets(
        'shows duplicate name error for case-insensitive match against '
        'another hive in the apiary', (tester) async {
      final apiClient = await _fakeApiClient([
        {
          'id': 1,
          'apiary_id': 1,
          'name': 'Alpha',
          'type': 'langstroth',
          'active': true,
          'grid_row': 0,
          'grid_col': 0,
        },
        {
          'id': 2,
          'apiary_id': 1,
          'name': 'Beta',
          'type': 'langstroth',
          'active': true,
          'grid_row': 0,
          'grid_col': 1,
        },
      ]);

      await tester.pumpWidget(_wrap(
        apiClient,
        const EditHiveScreen(apiaryId: 1, hive: _hive),
      ));
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(TextFormField).first, 'BETA');
      await tester.pump();

      expect(find.text(_duplicateNameError), findsOneWidget);
      expect(find.text('Edit hive'), findsOneWidget);
    });

    testWidgets('excludes the hive being edited from the duplicate check',
        (tester) async {
      final apiClient = await _fakeApiClient([
        {
          'id': 1,
          'apiary_id': 1,
          'name': 'Alpha',
          'type': 'langstroth',
          'active': true,
          'grid_row': 0,
          'grid_col': 0,
        },
      ]);

      await tester.pumpWidget(_wrap(
        apiClient,
        const EditHiveScreen(apiaryId: 1, hive: _hive),
      ));
      await tester.pumpAndSettle();

      // Re-entering the hive's own current name should not be flagged as a
      // duplicate, since the hive is excluded from the fetched name set.
      await tester.enterText(find.byType(TextFormField).first, 'Alpha');
      await tester.pump();

      expect(find.text(_duplicateNameError), findsNothing);
    });

    testWidgets('does not flag a unique name as a duplicate', (tester) async {
      final apiClient = await _fakeApiClient([
        {
          'id': 1,
          'apiary_id': 1,
          'name': 'Alpha',
          'type': 'langstroth',
          'active': true,
          'grid_row': 0,
          'grid_col': 0,
        },
        {
          'id': 2,
          'apiary_id': 1,
          'name': 'Beta',
          'type': 'langstroth',
          'active': true,
          'grid_row': 0,
          'grid_col': 1,
        },
      ]);

      await tester.pumpWidget(_wrap(
        apiClient,
        const EditHiveScreen(apiaryId: 1, hive: _hive),
      ));
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(TextFormField).first, 'Gamma');
      await tester.pump();

      expect(find.text(_duplicateNameError), findsNothing);
    });
  });
}
