import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:app/core/api/api_client.dart';
import 'package:app/core/storage/token_storage.dart';
import 'package:app/features/apiary/view/apiaries_screen.dart';
import 'package:app/features/apiary/view/create_apiary_screen.dart';
import 'package:app/l10n/app_localizations.dart';

/// Returns an empty apiaries list from `/api/v1/apiaries` and a benign
/// response for everything else, so ApiariesScreen loads and settles.
class _EmptyApiariesHttpClientAdapter implements HttpClientAdapter {
  @override
  void close({bool force = false}) {}

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    if (options.path.contains('/apiaries') && options.method == 'GET') {
      return ResponseBody.fromString(
        jsonEncode([]),
        200,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );
    }
    return ResponseBody.fromString(
      jsonEncode({}),
      200,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }
}

/// Returns a single apiary (with optional GPS coords) from
/// `/api/v1/apiaries` and a benign response for everything else.
class _NonEmptyApiariesHttpClientAdapter implements HttpClientAdapter {
  _NonEmptyApiariesHttpClientAdapter({this.withGps = false});

  final bool withGps;

  @override
  void close({bool force = false}) {}

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    if (options.path.contains('/apiaries') && options.method == 'GET') {
      return ResponseBody.fromString(
        jsonEncode([
          {
            'id': 1,
            'name': 'Backyard',
            'lat': withGps ? 50.0 : null,
            'lng': withGps ? 19.0 : null,
            'grid_rows': 2,
            'grid_cols': 2,
            'hive_count': 3,
            'user_role': 'owner',
          },
        ]),
        200,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );
    }
    return ResponseBody.fromString(
      jsonEncode({}),
      200,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }
}

Future<ApiClient> _fakeApiClient() async {
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();
  final tokenStorage = TokenStorage(prefs);
  final apiClient = ApiClient(storage: tokenStorage, baseUrl: 'http://test');
  apiClient.dio.httpClientAdapter = _EmptyApiariesHttpClientAdapter();
  return apiClient;
}

Future<ApiClient> _fakeApiClientWithApiary({required bool withGps}) async {
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();
  final tokenStorage = TokenStorage(prefs);
  final apiClient = ApiClient(storage: tokenStorage, baseUrl: 'http://test');
  apiClient.dio.httpClientAdapter =
      _NonEmptyApiariesHttpClientAdapter(withGps: withGps);
  return apiClient;
}

Widget _wrap(Widget child, {required ApiClient apiClient}) =>
    RepositoryProvider<ApiClient>.value(
      value: apiClient,
      child: MaterialApp(
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        locale: const Locale('en'),
        home: child,
      ),
    );

void main() {
  group('ApiariesScreen empty state', () {
    testWidgets('shows "no apiaries" text and an always-present banner', (
      tester,
    ) async {
      final apiClient = await _fakeApiClient();

      await tester.pumpWidget(
        _wrap(
          ApiariesScreen(onSelectSection: (_) {}),
          apiClient: apiClient,
        ),
      );
      await tester.pumpAndSettle();

      expect(find.text('You have no apiaries yet'), findsOneWidget);

      final addButton = tester.widget<IconButton>(
        find.widgetWithIcon(IconButton, Icons.add),
      );
      expect(addButton.onPressed, isNotNull);

      await tester.tap(find.byIcon(Icons.add));
      await tester.pumpAndSettle();

      expect(find.byType(CreateApiaryScreen), findsOneWidget);
      expect(find.byIcon(Icons.map_outlined), findsNothing);
    });
  });

  group('ApiariesScreen banner map button', () {
    testWidgets('is present but disabled when apiaries lack GPS', (
      tester,
    ) async {
      final apiClient = await _fakeApiClientWithApiary(withGps: false);

      await tester.pumpWidget(
        _wrap(
          ApiariesScreen(onSelectSection: (_) {}),
          apiClient: apiClient,
        ),
      );
      await tester.pumpAndSettle();

      final mapButton = tester.widget<IconButton>(
        find.widgetWithIcon(IconButton, Icons.map_outlined),
      );
      expect(mapButton.onPressed, isNull);
    });

    testWidgets('is present and enabled when an apiary has GPS', (
      tester,
    ) async {
      final apiClient = await _fakeApiClientWithApiary(withGps: true);

      await tester.pumpWidget(
        _wrap(
          ApiariesScreen(onSelectSection: (_) {}),
          apiClient: apiClient,
        ),
      );
      await tester.pumpAndSettle();

      final mapButton = tester.widget<IconButton>(
        find.widgetWithIcon(IconButton, Icons.map_outlined),
      );
      expect(mapButton.onPressed, isNotNull);
    });
  });
}
