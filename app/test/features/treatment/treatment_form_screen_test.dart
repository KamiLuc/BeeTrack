import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:app/core/api/api_client.dart';
import 'package:app/core/storage/token_storage.dart';
import 'package:app/features/hive/data/hive_model.dart';
import 'package:app/features/treatment/view/treatment_form_screen.dart';
import 'package:app/l10n/app_localizations.dart';

class _EmptyListAdapter implements HttpClientAdapter {
  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    return ResponseBody.fromString(
      '[]',
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
  apiClient.dio.httpClientAdapter = _EmptyListAdapter();
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
  id: 10,
  apiaryId: 1,
  name: 'Alpha',
  type: 'langstroth',
  active: true,
  queenNeedsReplacement: false,
  readyForHarvest: false,
  needsFood: false,
  boxNeedsAdding: false,
  gridRow: 0,
  gridCol: 0,
);

void main() {
  group('TreatmentFormScreen onSaveProposed', () {
    testWidgets(
        'calls onSaveProposed with form state and pops true, without hitting the repository',
        (tester) async {
      final apiClient = await _fakeApiClient();
      Map<String, dynamic>? received;

      await tester.pumpWidget(_wrap(
        apiClient,
        Navigator(
          onGenerateRoute: (settings) => MaterialPageRoute(
            builder: (_) => TreatmentFormScreen(
              apiaryId: 1,
              hive: _hive,
              onSaveProposed: (args) async {
                received = args;
              },
            ),
          ),
        ),
      ));
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(TextFormField).at(0), 'Apivar');
      await tester.enterText(find.byType(TextFormField).at(1), '2');
      await tester.enterText(find.byType(TextFormField).at(2), 'applied evenly');
      await tester.pump();

      await tester.tap(find.byIcon(Icons.check));
      await tester.pumpAndSettle();

      expect(received, isNotNull);
      expect(received!['medicine_name'], 'Apivar');
      expect(received!['dose'], '2');
      expect(received!['notes'], 'applied evenly');
      expect(find.byType(TreatmentFormScreen), findsNothing);
    });

    testWidgets(
        'shows an error and resets loading state when onSaveProposed throws',
        (tester) async {
      final apiClient = await _fakeApiClient();

      await tester.pumpWidget(_wrap(
        apiClient,
        TreatmentFormScreen(
          apiaryId: 1,
          hive: _hive,
          onSaveProposed: (args) async {
            throw Exception('boom');
          },
        ),
      ));
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(TextFormField).at(0), 'Apivar');
      await tester.enterText(find.byType(TextFormField).at(1), '2');
      await tester.pump();

      await tester.tap(find.byIcon(Icons.check));
      await tester.pumpAndSettle();

      expect(find.byType(TreatmentFormScreen), findsOneWidget);
      expect(find.byType(SnackBar), findsOneWidget);
      expect(find.byIcon(Icons.check), findsOneWidget);
    });
  });
}
