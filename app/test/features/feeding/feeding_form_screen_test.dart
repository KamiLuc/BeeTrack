import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:app/core/api/api_client.dart';
import 'package:app/core/storage/token_storage.dart';
import 'package:app/features/feeding/view/feeding_form_screen.dart';
import 'package:app/features/hive/data/hive_model.dart';
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
  group('FeedingFormScreen onSaveProposed', () {
    testWidgets(
        'calls onSaveProposed with form state and pops true, without hitting the repository',
        (tester) async {
      final apiClient = await _fakeApiClient();
      Map<String, dynamic>? received;

      await tester.pumpWidget(_wrap(
        apiClient,
        Navigator(
          onGenerateRoute: (settings) => MaterialPageRoute(
            builder: (_) => FeedingFormScreen(
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

      await tester.enterText(find.byType(TextFormField).at(0), 'Sugar syrup');
      await tester.enterText(find.byType(TextFormField).at(1), '1L');
      await tester.enterText(find.byType(TextFormField).at(2), 'topped up');
      await tester.pump();

      await tester.tap(find.byIcon(Icons.check));
      await tester.pumpAndSettle();

      expect(received, isNotNull);
      expect(received!['feed_type'], 'Sugar syrup');
      expect(received!['amount'], '1L');
      expect(received!['notes'], 'topped up');
      expect(find.byType(FeedingFormScreen), findsNothing);
    });

    testWidgets(
        'shows an error and resets loading state when onSaveProposed throws',
        (tester) async {
      final apiClient = await _fakeApiClient();

      await tester.pumpWidget(_wrap(
        apiClient,
        FeedingFormScreen(
          apiaryId: 1,
          hive: _hive,
          onSaveProposed: (args) async {
            throw Exception('boom');
          },
        ),
      ));
      await tester.pumpAndSettle();

      await tester.enterText(find.byType(TextFormField).at(0), 'Sugar syrup');
      await tester.enterText(find.byType(TextFormField).at(1), '1L');
      await tester.pump();

      await tester.tap(find.byIcon(Icons.check));
      await tester.pumpAndSettle();

      expect(find.byType(FeedingFormScreen), findsOneWidget);
      expect(find.byType(SnackBar), findsOneWidget);
      expect(find.byIcon(Icons.check), findsOneWidget);
    });
  });
}
