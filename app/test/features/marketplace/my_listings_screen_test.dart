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
import 'package:app/features/marketplace/view/create_listing_screen.dart';
import 'package:app/features/marketplace/view/listing_detail_screen.dart';
import 'package:app/features/marketplace/view/my_listings_screen.dart';
import 'package:app/l10n/app_localizations.dart';

class _MockAuthRepository extends Mock implements AuthRepository {}

/// Records requests and serves configurable responses for the endpoints
/// MyListingsScreen and its actions call.
class _RecordingAdapter implements HttpClientAdapter {
  final List<RequestOptions> requests = [];
  List<Map<String, dynamic>> listings = [];
  bool failSearch = false;
  bool failMutations = false;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    requests.add(options);

    if (options.path.endsWith('/listings') && options.method == 'GET') {
      if (failSearch) {
        throw DioException(requestOptions: options, message: 'search failed');
      }
      return _json({'items': listings, 'total': listings.length});
    }

    if (options.path.contains('/hide') && options.method == 'PATCH') {
      if (failMutations) {
        throw DioException(requestOptions: options, message: 'hide failed');
      }
      final id = int.parse(RegExp(r'/listings/(\d+)/hide').firstMatch(options.path)!.group(1)!);
      final hidden = (options.data as Map<String, dynamic>)['hidden'] as bool;
      final listing = listings.firstWhere((l) => l['id'] == id);
      return _json({...listing, 'is_hidden': hidden});
    }

    if (RegExp(r'/listings/\d+$').hasMatch(options.path) &&
        options.method == 'DELETE') {
      if (failMutations) {
        throw DioException(requestOptions: options, message: 'delete failed');
      }
      return _json({});
    }

    if (RegExp(r'/listings/\d+$').hasMatch(options.path) &&
        options.method == 'PATCH') {
      return _json(options.data as Map<String, dynamic>);
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
            home: const MyListingsScreen(),
          ),
        ),
      ),
    );

Map<String, dynamic> _listingJson({
  int id = 1,
  String title = 'Wildflower Honey',
  bool isHidden = false,
}) =>
    {
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
      'is_hidden': isHidden,
      'created_at': DateTime(2026, 1, 1).toIso8601String(),
      'updated_at': DateTime(2026, 1, 1).toIso8601String(),
      'images': [],
    };

void main() {
  late AppLocalizations l10n;
  setUpAll(() async {
    l10n = await AppLocalizations.delegate.load(const Locale('en'));
  });

  group('MyListingsScreen', () {
    testWidgets('shows the empty state when the user has no listings',
        (tester) async {
      final (apiClient, _) = await _fakeApiClient();

      await tester.pumpWidget(_wrap(apiClient));
      await tester.pumpAndSettle();

      expect(find.text(l10n.myListingsEmpty), findsOneWidget);
    });

    testWidgets(
        'shows a retry button on load failure, and retrying reloads the list',
        (tester) async {
      final (apiClient, adapter) = await _fakeApiClient();
      adapter.failSearch = true;

      await tester.pumpWidget(_wrap(apiClient));
      await tester.pumpAndSettle();

      expect(find.text(l10n.generalError), findsOneWidget);

      adapter.failSearch = false;
      adapter.listings = [_listingJson()];
      await tester.tap(find.widgetWithText(ElevatedButton, l10n.generalRetry));
      await tester.pumpAndSettle();

      expect(find.text('Wildflower Honey'), findsOneWidget);
    });

    testWidgets('renders a listing with the hidden badge when hidden',
        (tester) async {
      final (apiClient, adapter) = await _fakeApiClient();
      adapter.listings = [_listingJson(isHidden: true)];

      await tester.pumpWidget(_wrap(apiClient));
      await tester.pumpAndSettle();

      expect(find.text('Wildflower Honey'), findsOneWidget);
      expect(find.text(l10n.marketplaceHiddenBadge), findsOneWidget);
    });

    testWidgets('tapping a card does nothing — edit is only via the menu',
        (tester) async {
      final (apiClient, adapter) = await _fakeApiClient();
      adapter.listings = [_listingJson()];

      await tester.pumpWidget(_wrap(apiClient));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Wildflower Honey'));
      await tester.pumpAndSettle();

      expect(find.byType(ListingDetailScreen), findsNothing);
      expect(find.byType(CreateListingScreen), findsNothing);
      expect(find.byType(MyListingsScreen), findsOneWidget);
    });

    testWidgets(
        'edit action opens CreateListingScreen prefilled and reloads on '
        'successful edit', (tester) async {
      final (apiClient, adapter) = await _fakeApiClient();
      adapter.listings = [_listingJson()];

      await tester.pumpWidget(_wrap(apiClient));
      await tester.pumpAndSettle();

      await tester.tap(find.byIcon(Icons.more_vert));
      await tester.pumpAndSettle();
      await tester.tap(find.text(l10n.generalEdit));
      await tester.pumpAndSettle();

      expect(find.byType(CreateListingScreen), findsOneWidget);
      expect(find.text(l10n.marketplaceEditScreenTitle), findsOneWidget);
    });

    testWidgets('hide action toggles the hidden badge', (tester) async {
      final (apiClient, adapter) = await _fakeApiClient();
      adapter.listings = [_listingJson()];

      await tester.pumpWidget(_wrap(apiClient));
      await tester.pumpAndSettle();

      expect(find.text(l10n.marketplaceHiddenBadge), findsNothing);

      await tester.tap(find.byIcon(Icons.more_vert));
      await tester.pumpAndSettle();
      await tester.tap(find.text(l10n.marketplaceHideListing));
      await tester.pumpAndSettle();

      expect(find.text(l10n.marketplaceHiddenBadge), findsOneWidget);
      expect(
        adapter.requests.any(
          (r) => r.path.contains('/listings/1/hide') && r.method == 'PATCH',
        ),
        isTrue,
      );
    });

    testWidgets('delete action removes the card only after confirming',
        (tester) async {
      final (apiClient, adapter) = await _fakeApiClient();
      adapter.listings = [_listingJson()];

      await tester.pumpWidget(_wrap(apiClient));
      await tester.pumpAndSettle();

      await tester.tap(find.byIcon(Icons.more_vert));
      await tester.pumpAndSettle();
      await tester.tap(find.text(l10n.generalDelete));
      await tester.pumpAndSettle();

      expect(find.text(l10n.marketplaceDeleteConfirm), findsOneWidget);

      await tester.tap(find.text(l10n.generalCancel));
      await tester.pumpAndSettle();

      expect(find.text('Wildflower Honey'), findsOneWidget);

      await tester.tap(find.byIcon(Icons.more_vert));
      await tester.pumpAndSettle();
      await tester.tap(find.text(l10n.generalDelete));
      await tester.pumpAndSettle();
      await tester.tap(find.text(l10n.generalDelete).last);
      await tester.pumpAndSettle();

      expect(find.text('Wildflower Honey'), findsNothing);
      expect(find.text(l10n.myListingsEmpty), findsOneWidget);
    });
  });
}
