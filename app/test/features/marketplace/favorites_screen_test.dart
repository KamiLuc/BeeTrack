import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:mocktail/mocktail.dart';

import 'package:app/core/api/api_client.dart';
import 'package:app/core/storage/token_storage.dart';
import 'package:app/features/auth/bloc/auth_bloc.dart';
import 'package:app/features/auth/data/auth_repository.dart';
import 'package:app/features/marketplace/view/favorites_screen.dart';
import 'package:app/features/marketplace/view/listing_detail_screen.dart';
import 'package:app/l10n/app_localizations.dart';

class _MockAuthRepository extends Mock implements AuthRepository {}

/// Records requests and serves configurable responses for the endpoints
/// FavoritesScreen and its actions call.
class _RecordingAdapter implements HttpClientAdapter {
  final List<RequestOptions> requests = [];
  List<Map<String, dynamic>> favorites = [];

  bool failList = false;
  bool failRemove = false;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    requests.add(options);

    if (options.path.endsWith('/favorites') && options.method == 'GET') {
      if (failList) {
        throw DioException(requestOptions: options, message: 'list failed');
      }
      return _json({'items': favorites});
    }

    if (RegExp(r'/listings/\d+/favorite$').hasMatch(options.path) &&
        options.method == 'DELETE') {
      if (failRemove) {
        throw DioException(requestOptions: options, message: 'remove failed');
      }
      final id = int.parse(
        RegExp(
          r'/listings/(\d+)/favorite$',
        ).firstMatch(options.path)!.group(1)!,
      );
      favorites = favorites.where((l) => l['id'] != id).toList();
      return _json({});
    }

    if (RegExp(r'/listings/\d+/favorite$').hasMatch(options.path) &&
        options.method == 'POST') {
      return _json({});
    }

    return _json({});
  }

  ResponseBody _json(Object? data) => ResponseBody.fromString(
    jsonEncode(data),
    200,
    headers: {
      Headers.contentTypeHeader: [Headers.jsonContentType],
    },
  );

  @override
  void close({bool force = false}) {}
}

/// The [TokenStorage] backing the current test's [ApiClient], also handed to
/// [_wrap] via a [RepositoryProvider] so `context.read<TokenStorage>()`
/// resolves inside a pushed ListingDetailScreen, just like in main.dart.
late TokenStorage _tokenStorage;

Future<(ApiClient, _RecordingAdapter)> _fakeApiClient() async {
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();
  _tokenStorage = TokenStorage(prefs);
  final apiClient = ApiClient(storage: _tokenStorage, baseUrl: 'http://test');
  final adapter = _RecordingAdapter();
  apiClient.dio.httpClientAdapter = adapter;
  return (apiClient, adapter);
}

Widget _wrap(ApiClient apiClient) => RepositoryProvider<ApiClient>.value(
  value: apiClient,
  child: RepositoryProvider<TokenStorage>.value(
    value: _tokenStorage,
    child: BlocProvider<AuthBloc>(
      create: (_) => AuthBloc(auth: _MockAuthRepository()),
      child: MaterialApp(
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        locale: const Locale('en'),
        home: const FavoritesScreen(),
      ),
    ),
  ),
);

Map<String, dynamic> _listingJson({
  int id = 1,
  String title = 'Wildflower Honey',
}) => {
  'id': id,
  'user_id': 1,
  'title': title,
  'description': 'Fresh honey.',
  'category': 'HONEY',
  'price': 20.0,
  'quantity': '5 jars',
  'address': 'Krakow',
  'contact_phone': '123456789',
  'contact_email': 'seller@example.com',
  'is_hidden': false,
  'created_at': DateTime(2026, 1, 1).toIso8601String(),
  'updated_at': DateTime(2026, 1, 1).toIso8601String(),
  'images': [],
};

void main() {
  late AppLocalizations l10n;
  setUpAll(() async {
    l10n = await AppLocalizations.delegate.load(const Locale('en'));
  });

  group('FavoritesScreen', () {
    testWidgets('shows the empty state when the user has no favorites', (
      tester,
    ) async {
      final (apiClient, _) = await _fakeApiClient();

      await tester.pumpWidget(_wrap(apiClient));
      await tester.pumpAndSettle();

      expect(find.text(l10n.favoritesEmpty), findsOneWidget);
    });

    testWidgets(
      'shows a retry button on load failure, and retrying reloads the list',
      (tester) async {
        final (apiClient, adapter) = await _fakeApiClient();
        adapter.failList = true;

        await tester.pumpWidget(_wrap(apiClient));
        await tester.pumpAndSettle();

        expect(find.text(l10n.generalError), findsOneWidget);

        adapter.failList = false;
        adapter.favorites = [_listingJson()];
        await tester.tap(
          find.widgetWithText(ElevatedButton, l10n.generalRetry),
        );
        await tester.pumpAndSettle();

        expect(find.text('Wildflower Honey'), findsOneWidget);
      },
    );

    testWidgets('renders favorited listings', (tester) async {
      final (apiClient, adapter) = await _fakeApiClient();
      adapter.favorites = [_listingJson()];

      await tester.pumpWidget(_wrap(apiClient));
      await tester.pumpAndSettle();

      expect(find.text('Wildflower Honey'), findsOneWidget);
      expect(find.byIcon(Icons.favorite), findsOneWidget);
    });

    testWidgets('tapping a card opens ListingDetailScreen', (tester) async {
      final (apiClient, adapter) = await _fakeApiClient();
      adapter.favorites = [_listingJson()];

      await tester.pumpWidget(_wrap(apiClient));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Wildflower Honey'));
      await tester.pumpAndSettle();

      expect(find.byType(ListingDetailScreen), findsOneWidget);
    });

    testWidgets(
      'tapping the heart icon unfavorites without removing the card, and '
      'tapping again undoes it',
      (tester) async {
        final (apiClient, adapter) = await _fakeApiClient();
        adapter.favorites = [_listingJson()];

        await tester.pumpWidget(_wrap(apiClient));
        await tester.pumpAndSettle();

        expect(find.text('Wildflower Honey'), findsOneWidget);

        await tester.tap(find.byIcon(Icons.favorite));
        await tester.pumpAndSettle();

        // Card stays visible with the outline heart instead of disappearing,
        // so an accidental tap doesn't silently drop the listing.
        expect(find.text('Wildflower Honey'), findsOneWidget);
        expect(find.byIcon(Icons.favorite), findsNothing);
        expect(find.byIcon(Icons.favorite_border), findsOneWidget);
        expect(
          adapter.requests.any(
            (r) =>
                r.path.contains('/listings/1/favorite') &&
                r.method == 'DELETE',
          ),
          isTrue,
        );

        // Tapping again re-favorites it.
        await tester.tap(find.byIcon(Icons.favorite_border));
        await tester.pumpAndSettle();

        expect(find.byIcon(Icons.favorite), findsOneWidget);
        expect(
          adapter.requests.any(
            (r) =>
                r.path.contains('/listings/1/favorite') &&
                r.method == 'POST',
          ),
          isTrue,
        );
      },
    );

    testWidgets(
      'remove failure shows an error snackbar and reverts the heart',
      (tester) async {
        final (apiClient, adapter) = await _fakeApiClient();
        adapter.favorites = [_listingJson()];
        adapter.failRemove = true;

        await tester.pumpWidget(_wrap(apiClient));
        await tester.pumpAndSettle();

        await tester.tap(find.byIcon(Icons.favorite));
        await tester.pumpAndSettle();

        expect(find.text('Wildflower Honey'), findsOneWidget);
        expect(find.byIcon(Icons.favorite), findsOneWidget);
        expect(find.text(l10n.generalError), findsOneWidget);
      },
    );
  });
}
