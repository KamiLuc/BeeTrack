import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:app/core/api/api_client.dart';
import 'package:app/core/storage/token_storage.dart';
import 'package:app/core/widgets/app_drawer.dart';
import 'package:app/features/auth/bloc/auth_bloc.dart';
import 'package:app/features/auth/data/auth_repository.dart';
import 'package:app/features/marketplace/view/create_listing_screen.dart';
import 'package:app/features/marketplace/view/favorites_screen.dart';
import 'package:app/features/marketplace/view/listing_detail_screen.dart';
import 'package:app/features/marketplace/view/marketplace_home_screen.dart';
import 'package:app/features/marketplace/view/my_listings_screen.dart';
import 'package:app/l10n/app_localizations.dart';

class _MockAuthRepository extends Mock implements AuthRepository {}

/// Builds a minimal (unsigned) JWT carrying the given `sub` claim, so
/// [TokenStorage.userId] can be exercised without a real auth flow.
String _jwtWithSub(int sub) {
  String segment(Object data) =>
      base64Url.encode(utf8.encode(jsonEncode(data))).replaceAll('=', '');
  return '${segment({'alg': 'none'})}.${segment({'sub': sub})}.signature';
}

/// Fails every request synchronously so tests never leave a pending network
/// timer behind when MarketplaceHomeScreen eagerly loads listings.
class _FailingHttpClientAdapter implements HttpClientAdapter {
  @override
  void close({bool force = false}) {}

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) {
    throw DioException(requestOptions: options, message: 'no network in tests');
  }
}

/// The [TokenStorage] backing the current test's [ApiClient], also handed to
/// [_wrap] via a [RepositoryProvider] so `context.read<TokenStorage>()`
/// resolves just like it does in the real widget tree (see main.dart).
late TokenStorage _tokenStorage;

Future<ApiClient> _fakeApiClient() async {
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();
  _tokenStorage = TokenStorage(prefs);
  final apiClient = ApiClient(storage: _tokenStorage, baseUrl: 'http://test');
  apiClient.dio.httpClientAdapter = _FailingHttpClientAdapter();
  return apiClient;
}

/// Returns a single listing for `/api/v1/listings` so tap-to-navigate can be
/// exercised without a real backend.
class _ListingsHttpClientAdapter implements HttpClientAdapter {
  _ListingsHttpClientAdapter({
    this.price = 20.0,
    this.total = 1,
    int? mineTotal,
    this.itemsPerPage = 1,
    this.favoritesCount = 0,
  }) : mineTotal = mineTotal ?? total;

  final double? price;
  final int total;
  final int mineTotal;
  final int itemsPerPage;
  final int favoritesCount;
  int listingsRequestCount = 0;
  final List<RequestOptions> listingsRequests = [];

  @override
  void close({bool force = false}) {}

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    final isMine = options.queryParameters['mine'] == true;
    if (options.path.contains('/listings') && options.method == 'GET') {
      listingsRequestCount++;
      listingsRequests.add(options);
    }
    if (options.path.contains('/listings')) {
      final offset = int.tryParse(
            '${options.queryParameters['offset'] ?? 0}',
          ) ??
          0;
      final pageTotal = isMine ? mineTotal : total;
      final count = isMine ? 1 : itemsPerPage;
      return ResponseBody.fromString(
        jsonEncode({
          'items': [
            for (var i = 0; i < count; i++)
              {
                'id': 7 + offset + i,
                'user_id': 1,
                'title': 'Wildflower Honey',
                'description': 'Fresh honey.',
                'category': 'honey',
                'price': price,
                'quantity': '5 jars',
                'address': 'Krakow',
                'contact_phone': '123456789',
                'contact_email': 'seller@example.com',
                'is_hidden': false,
                'created_at': DateTime(2026, 1, 1).toIso8601String(),
                'updated_at': DateTime(2026, 1, 1).toIso8601String(),
                'images': [],
              },
          ],
          'total': pageTotal,
        }),
        200,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );
    }
    if (options.path.contains('/favorites') && options.method == 'GET') {
      return ResponseBody.fromString(
        jsonEncode({
          'items': [
            for (var i = 0; i < favoritesCount; i++)
              {
                'id': 100 + i,
                'user_id': 2,
                'title': 'Favorited Listing',
                'description': 'Fresh honey.',
                'category': 'honey',
                'price': price,
                'quantity': '5 jars',
                'address': 'Krakow',
                'contact_phone': '123456789',
                'contact_email': 'seller@example.com',
                'is_hidden': false,
                'created_at': DateTime(2026, 1, 1).toIso8601String(),
                'updated_at': DateTime(2026, 1, 1).toIso8601String(),
                'images': [],
              },
          ],
          'total': favoritesCount,
        }),
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

Future<ApiClient> _fakeApiClientWithListings({
  _ListingsHttpClientAdapter? adapter,
}) async {
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();
  _tokenStorage = TokenStorage(prefs);
  final apiClient = ApiClient(storage: _tokenStorage, baseUrl: 'http://test');
  apiClient.dio.httpClientAdapter = adapter ?? _ListingsHttpClientAdapter();
  return apiClient;
}

Widget _wrap(
  Widget child, {
  required ApiClient apiClient,
  AuthBloc? authBloc,
}) => RepositoryProvider<ApiClient>.value(
  value: apiClient,
  child: RepositoryProvider<TokenStorage>.value(
    value: _tokenStorage,
    child: BlocProvider<AuthBloc>.value(
      value: authBloc ?? AuthBloc(auth: _MockAuthRepository()),
      child: MaterialApp(
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        locale: const Locale('en'),
        home: child,
      ),
    ),
  ),
);

void main() {
  group('MarketplaceHomeScreen', () {
    testWidgets(
      'unauthenticated: shows marketplace and drawer with login option',
      (tester) async {
        final apiClient = await _fakeApiClient();
        final authBloc = AuthBloc(auth: _MockAuthRepository());
        await tester.pumpWidget(
          _wrap(
            const MarketplaceHomeScreen(),
            apiClient: apiClient,
            authBloc: authBloc,
          ),
        );
        await tester.pumpAndSettle();

        expect(find.text('Marketplace'), findsOneWidget);

        await tester.tap(find.byTooltip('Open navigation menu'));
        await tester.pumpAndSettle();

        expect(find.widgetWithText(ListTile, 'Log in'), findsOneWidget);
        expect(find.widgetWithText(ListTile, 'Apiaries'), findsOneWidget);
      },
    );

    testWidgets(
      'unauthenticated: while browsing, Marketplace tile is selected and '
      'locked Apiaries tile is not, and tapping Apiaries calls onLogin',
      (tester) async {
        final apiClient = await _fakeApiClient();
        final authBloc = AuthBloc(auth: _MockAuthRepository());
        var loginTapped = false;
        await tester.pumpWidget(
          _wrap(
            MarketplaceHomeScreen(onLogin: () => loginTapped = true),
            apiClient: apiClient,
            authBloc: authBloc,
          ),
        );
        await tester.pumpAndSettle();

        await tester.tap(find.byTooltip('Open navigation menu'));
        await tester.pumpAndSettle();

        final apiariesTile = tester.widget<ListTile>(
          find.widgetWithText(ListTile, 'Apiaries'),
        );
        expect(apiariesTile.selected, isFalse);

        final marketplaceTile = tester.widget<ListTile>(
          find.widgetWithText(ListTile, 'Marketplace'),
        );
        expect(marketplaceTile.selected, isTrue);

        await tester.tap(find.widgetWithText(ListTile, 'Apiaries'));
        await tester.pumpAndSettle();

        expect(loginTapped, isTrue);
      },
    );

    testWidgets('unauthenticated: tapping Log in invokes onLogin callback', (
      tester,
    ) async {
      final apiClient = await _fakeApiClient();
      final authBloc = AuthBloc(auth: _MockAuthRepository());
      var loginTapped = false;
      await tester.pumpWidget(
        _wrap(
          MarketplaceHomeScreen(onLogin: () => loginTapped = true),
          apiClient: apiClient,
          authBloc: authBloc,
        ),
      );
      await tester.pumpAndSettle();

      await tester.tap(find.byTooltip('Open navigation menu'));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(ListTile, 'Log in'));
      await tester.pumpAndSettle();

      expect(loginTapped, isTrue);
    });

    testWidgets('authenticated: shows marketplace with apiaries option', (
      tester,
    ) async {
      final apiClient = await _fakeApiClient();
      final authBloc = AuthBloc(auth: _MockAuthRepository())
        ..emit(AuthAuthenticated());

      await tester.pumpWidget(
        _wrap(
          const MarketplaceHomeScreen(),
          apiClient: apiClient,
          authBloc: authBloc,
        ),
      );
      await tester.pumpAndSettle();

      expect(find.text('Marketplace'), findsOneWidget);

      await tester.tap(find.byTooltip('Open navigation menu'));
      await tester.pumpAndSettle();

      expect(find.widgetWithText(ListTile, 'Apiaries'), findsOneWidget);
      expect(find.widgetWithText(ListTile, 'Log out'), findsOneWidget);
    });

    testWidgets('re-selecting marketplace just closes drawer', (tester) async {
      final apiClient = await _fakeApiClient();
      final authBloc = AuthBloc(auth: _MockAuthRepository())
        ..emit(AuthAuthenticated());
      var selected = <AppSection>[];

      await tester.pumpWidget(
        _wrap(
          MarketplaceHomeScreen(onSelectSection: selected.add),
          apiClient: apiClient,
          authBloc: authBloc,
        ),
      );
      await tester.pumpAndSettle();

      await tester.tap(find.byTooltip('Open navigation menu'));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(ListTile, 'Marketplace'));
      await tester.pumpAndSettle();

      expect(find.text('Marketplace'), findsOneWidget);
      expect(selected, isEmpty);
    });

    testWidgets(
      'authenticated: selecting apiaries from drawer invokes onSelectSection',
      (tester) async {
        final apiClient = await _fakeApiClient();
        final authBloc = AuthBloc(auth: _MockAuthRepository())
          ..emit(AuthAuthenticated());
        var selected = <AppSection>[];

        await tester.pumpWidget(
          _wrap(
            MarketplaceHomeScreen(onSelectSection: selected.add),
            apiClient: apiClient,
            authBloc: authBloc,
          ),
        );
        await tester.pumpAndSettle();

        await tester.tap(find.byTooltip('Open navigation menu'));
        await tester.pumpAndSettle();

        await tester.tap(find.widgetWithText(ListTile, 'Apiaries'));
        await tester.pumpAndSettle();

        expect(selected, [AppSection.apiaries]);
      },
    );

    testWidgets('authenticated: tapping Log out dispatches LogoutRequested', (
      tester,
    ) async {
      final apiClient = await _fakeApiClient();
      final repo = _MockAuthRepository();
      when(() => repo.logout()).thenAnswer((_) async {});
      final authBloc = AuthBloc(auth: repo)..emit(AuthAuthenticated());

      await tester.pumpWidget(
        _wrap(
          const MarketplaceHomeScreen(),
          apiClient: apiClient,
          authBloc: authBloc,
        ),
      );
      await tester.pumpAndSettle();

      await tester.tap(find.byTooltip('Open navigation menu'));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(ListTile, 'Log out'));
      await tester.pumpAndSettle();

      expect(authBloc.state, isA<AuthUnauthenticated>());
    });

    testWidgets('shows retry option when listings fail to load', (
      tester,
    ) async {
      final apiClient = await _fakeApiClient();
      final authBloc = AuthBloc(auth: _MockAuthRepository());

      await tester.pumpWidget(
        _wrap(
          const MarketplaceHomeScreen(),
          apiClient: apiClient,
          authBloc: authBloc,
        ),
      );
      await tester.pumpAndSettle();

      expect(
        find.text('Something went wrong. Please try again.'),
        findsOneWidget,
      );
      expect(find.widgetWithText(ElevatedButton, 'Retry'), findsOneWidget);
    });

    testWidgets('tapping a listing card navigates to ListingDetailScreen', (
      tester,
    ) async {
      final apiClient = await _fakeApiClientWithListings();
      final authBloc = AuthBloc(auth: _MockAuthRepository());

      await tester.pumpWidget(
        _wrap(
          const MarketplaceHomeScreen(),
          apiClient: apiClient,
          authBloc: authBloc,
        ),
      );
      await tester.pumpAndSettle();

      expect(find.text('Wildflower Honey'), findsOneWidget);

      await tester.tap(find.text('Wildflower Honey'));
      await tester.pumpAndSettle();

      expect(find.byType(ListingDetailScreen), findsOneWidget);
    });

    testWidgets(
        'reloads the feed after returning from ListingDetailScreen (so an '
        'edit is reflected)', (tester) async {
      final adapter = _ListingsHttpClientAdapter();
      final apiClient = await _fakeApiClientWithListings(adapter: adapter);
      final authBloc = AuthBloc(auth: _MockAuthRepository());

      await tester.pumpWidget(_wrap(
        const MarketplaceHomeScreen(),
        apiClient: apiClient,
        authBloc: authBloc,
      ));
      await tester.pumpAndSettle();

      // Anonymous, so no mine=true request is fired alongside the feed load.
      expect(adapter.listingsRequestCount, 1);

      await tester.tap(find.text('Wildflower Honey'));
      await tester.pumpAndSettle();
      expect(find.byType(ListingDetailScreen), findsOneWidget);

      await tester.pageBack();
      await tester.pumpAndSettle();

      expect(find.byType(ListingDetailScreen), findsNothing);
      expect(adapter.listingsRequestCount, 2);
    });

    testWidgets('listing card shows "Free" instead of "0.00" when price is 0', (
      tester,
    ) async {
      final apiClient = await _fakeApiClientWithListings(
        adapter: _ListingsHttpClientAdapter(price: 0),
      );
      final authBloc = AuthBloc(auth: _MockAuthRepository());

      await tester.pumpWidget(
        _wrap(
          const MarketplaceHomeScreen(),
          apiClient: apiClient,
          authBloc: authBloc,
        ),
      );
      await tester.pumpAndSettle();

      expect(find.text('Free'), findsOneWidget);
      expect(find.text('0.00'), findsNothing);
    });

    testWidgets(
      'no bottom loading spinner when everything fits on one page',
      (tester) async {
        final apiClient = await _fakeApiClientWithListings(
          adapter: _ListingsHttpClientAdapter(total: 1),
        );
        final authBloc = AuthBloc(auth: _MockAuthRepository());

        await tester.pumpWidget(
          _wrap(
            const MarketplaceHomeScreen(),
            apiClient: apiClient,
            authBloc: authBloc,
          ),
        );
        await tester.pumpAndSettle();

        expect(find.byType(CircularProgressIndicator), findsNothing);
      },
    );

    testWidgets(
      'shows a bottom loading spinner when there are more listings to load',
      (tester) async {
        final adapter = _ListingsHttpClientAdapter(total: 45);
        final apiClient = await _fakeApiClientWithListings(adapter: adapter);
        final authBloc = AuthBloc(auth: _MockAuthRepository());

        await tester.pumpWidget(
          _wrap(
            const MarketplaceHomeScreen(),
            apiClient: apiClient,
            authBloc: authBloc,
          ),
        );
        // hasMore is true, so the trailing spinner animates forever and
        // pumpAndSettle would never return; pump a bounded number of frames
        // to let the initial load's futures resolve instead.
        await tester.pump();
        await tester.pump(const Duration(milliseconds: 50));
        await tester.pump(const Duration(milliseconds: 50));

        expect(find.byType(CircularProgressIndicator), findsOneWidget);
      },
    );

    testWidgets(
      'scrolling near the bottom of the feed triggers loadMore and requests '
      'the next offset',
      (tester) async {
        final adapter = _ListingsHttpClientAdapter(total: 45, itemsPerPage: 20);
        final apiClient = await _fakeApiClientWithListings(adapter: adapter);
        final authBloc = AuthBloc(auth: _MockAuthRepository());

        await tester.pumpWidget(
          _wrap(
            const MarketplaceHomeScreen(),
            apiClient: apiClient,
            authBloc: authBloc,
          ),
        );
        await tester.pump();
        await tester.pump(const Duration(milliseconds: 50));
        await tester.pump(const Duration(milliseconds: 50));

        final requestCountBeforeScroll = adapter.listingsRequestCount;

        await tester.drag(find.byType(ListView), const Offset(0, -100000));
        await tester.pump();
        await tester.pump(const Duration(milliseconds: 50));
        await tester.pump(const Duration(milliseconds: 50));

        final offsets = adapter.listingsRequests
            .where((r) => r.queryParameters['mine'] != true)
            .map((r) => r.queryParameters['offset'])
            .toList();
        expect(offsets, contains(20));
        expect(
          adapter.listingsRequestCount,
          greaterThan(requestCountBeforeScroll),
        );
      },
    );

    testWidgets('non-owner: favorite heart is shown on the listing card', (
      tester,
    ) async {
      final apiClient = await _fakeApiClientWithListings();
      await _tokenStorage.save(access: _jwtWithSub(99), refresh: 'refresh');
      final authBloc = AuthBloc(auth: _MockAuthRepository())
        ..emit(AuthAuthenticated());

      await tester.pumpWidget(
        _wrap(
          const MarketplaceHomeScreen(),
          apiClient: apiClient,
          authBloc: authBloc,
        ),
      );
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.favorite_border), findsOneWidget);
    });

    testWidgets(
      'owner: no favorite heart on the listing card — cannot favorite own '
      'listing',
      (tester) async {
        final apiClient = await _fakeApiClientWithListings();
        await _tokenStorage.save(access: _jwtWithSub(1), refresh: 'refresh');
        final authBloc = AuthBloc(auth: _MockAuthRepository())
          ..emit(AuthAuthenticated());

        await tester.pumpWidget(
          _wrap(
            const MarketplaceHomeScreen(),
            apiClient: apiClient,
            authBloc: authBloc,
          ),
        );
        await tester.pumpAndSettle();

        expect(find.byIcon(Icons.favorite_border), findsNothing);
        expect(find.byIcon(Icons.favorite), findsNothing);
      },
    );

    testWidgets('unauthenticated: the create-listing "+" button is not shown', (
      tester,
    ) async {
      final apiClient = await _fakeApiClient();
      final authBloc = AuthBloc(auth: _MockAuthRepository());

      await tester.pumpWidget(
        _wrap(
          const MarketplaceHomeScreen(),
          apiClient: apiClient,
          authBloc: authBloc,
        ),
      );
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.add), findsNothing);
    });

    testWidgets(
      'authenticated: tapping the "+" button opens CreateListingScreen '
      'and refreshes the feed on return',
      (tester) async {
        final adapter = _ListingsHttpClientAdapter();
        final apiClient = await _fakeApiClientWithListings(adapter: adapter);
        await _tokenStorage.save(access: _jwtWithSub(1), refresh: 'refresh');
        final authBloc = AuthBloc(auth: _MockAuthRepository())
          ..emit(AuthAuthenticated());

        await tester.pumpWidget(
          _wrap(
            const MarketplaceHomeScreen(),
            apiClient: apiClient,
            authBloc: authBloc,
          ),
        );
        await tester.pumpAndSettle();

        expect(adapter.listingsRequestCount, 2);

        await tester.tap(find.byIcon(Icons.add));
        await tester.pumpAndSettle();

        expect(find.byType(CreateListingScreen), findsOneWidget);

        await tester.pageBack();
        await tester.pumpAndSettle();

        expect(find.byType(CreateListingScreen), findsNothing);
        expect(adapter.listingsRequestCount, 4);
      },
    );

    testWidgets(
      'the map button is shown in the bottom banner regardless of auth '
      'and stays disabled',
      (tester) async {
        final apiClient = await _fakeApiClient();
        final authBloc = AuthBloc(auth: _MockAuthRepository());

        await tester.pumpWidget(
          _wrap(
            const MarketplaceHomeScreen(),
            apiClient: apiClient,
            authBloc: authBloc,
          ),
        );
        await tester.pumpAndSettle();

        final mapButton = tester.widget<IconButton>(
          find.widgetWithIcon(IconButton, Icons.map_outlined),
        );
        expect(mapButton.onPressed, isNull);
      },
    );

    testWidgets(
      'authenticated: the "+" button is shown next to the map button in '
      'the bottom banner',
      (tester) async {
        final apiClient = await _fakeApiClient();
        final authBloc = AuthBloc(auth: _MockAuthRepository())
          ..emit(AuthAuthenticated());

        await tester.pumpWidget(
          _wrap(
            const MarketplaceHomeScreen(),
            apiClient: apiClient,
            authBloc: authBloc,
          ),
        );
        await tester.pumpAndSettle();

        expect(find.byIcon(Icons.add), findsOneWidget);
        expect(find.byIcon(Icons.map_outlined), findsOneWidget);
      },
    );

    testWidgets(
      'typing in the search field debounces and calls setKeyword without '
      'pressing enter',
      (tester) async {
        final adapter = _ListingsHttpClientAdapter();
        final apiClient = await _fakeApiClientWithListings(adapter: adapter);
        final authBloc = AuthBloc(auth: _MockAuthRepository());

        await tester.pumpWidget(
          _wrap(
            const MarketplaceHomeScreen(),
            apiClient: apiClient,
            authBloc: authBloc,
          ),
        );
        await tester.pumpAndSettle();

        // Anonymous, so no mine=true request is fired alongside the feed load.
        expect(adapter.listingsRequestCount, 1);

        await tester.enterText(find.byType(TextField), 'honey');
        await tester.pump(const Duration(milliseconds: 200));
        expect(adapter.listingsRequestCount, 1);

        await tester.pump(const Duration(milliseconds: 250));
        await tester.pump();
        expect(adapter.listingsRequestCount, 2);
        expect(
          adapter.listingsRequests
              .lastWhere((r) => r.queryParameters['mine'] != true)
              .queryParameters['keyword'],
          'honey',
        );
      },
    );

    testWidgets('unauthenticated: the "My listings" icon is not shown', (
      tester,
    ) async {
      final apiClient = await _fakeApiClient();
      final authBloc = AuthBloc(auth: _MockAuthRepository());

      await tester.pumpWidget(
        _wrap(
          const MarketplaceHomeScreen(),
          apiClient: apiClient,
          authBloc: authBloc,
        ),
      );
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.list_alt_outlined), findsNothing);
    });

    testWidgets(
      'authenticated: tapping the "My listings" icon opens MyListingsScreen '
      'and refreshes the feed on return',
      (tester) async {
        final adapter = _ListingsHttpClientAdapter();
        final apiClient = await _fakeApiClientWithListings(adapter: adapter);
        await _tokenStorage.save(access: _jwtWithSub(1), refresh: 'refresh');
        final authBloc = AuthBloc(auth: _MockAuthRepository())
          ..emit(AuthAuthenticated());

        await tester.pumpWidget(
          _wrap(
            const MarketplaceHomeScreen(),
            apiClient: apiClient,
            authBloc: authBloc,
          ),
        );
        await tester.pumpAndSettle();

        expect(adapter.listingsRequestCount, 2);
        expect(find.byIcon(Icons.list_alt_outlined), findsOneWidget);

        await tester.tap(find.byIcon(Icons.list_alt_outlined));
        await tester.pumpAndSettle();

        expect(find.byType(MyListingsScreen), findsOneWidget);
        // MyListingsScreen fetches its own listings on open.
        expect(adapter.listingsRequestCount, 3);

        await tester.pageBack();
        await tester.pumpAndSettle();

        expect(find.byType(MyListingsScreen), findsNothing);
        // MarketplaceHomeScreen reloads its feed when the pushed route returns.
        expect(adapter.listingsRequestCount, 5);
      },
    );

    testWidgets(
      'authenticated but with no own listings: the "My listings" icon is '
      'not shown',
      (tester) async {
        final apiClient = await _fakeApiClientWithListings(
          adapter: _ListingsHttpClientAdapter(mineTotal: 0),
        );
        await _tokenStorage.save(access: _jwtWithSub(1), refresh: 'refresh');
        final authBloc = AuthBloc(auth: _MockAuthRepository())
          ..emit(AuthAuthenticated());

        await tester.pumpWidget(
          _wrap(
            const MarketplaceHomeScreen(),
            apiClient: apiClient,
            authBloc: authBloc,
          ),
        );
        await tester.pumpAndSettle();

        expect(find.byIcon(Icons.list_alt_outlined), findsNothing);
      },
    );

    testWidgets('unauthenticated: the "Favorites" icon is not shown', (
      tester,
    ) async {
      final apiClient = await _fakeApiClient();
      final authBloc = AuthBloc(auth: _MockAuthRepository());

      await tester.pumpWidget(
        _wrap(
          const MarketplaceHomeScreen(),
          apiClient: apiClient,
          authBloc: authBloc,
        ),
      );
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.bookmark_border), findsNothing);
    });

    testWidgets(
      'authenticated but with no favorites: the "Favorites" icon is not '
      'shown',
      (tester) async {
        final apiClient = await _fakeApiClientWithListings(
          adapter: _ListingsHttpClientAdapter(favoritesCount: 0),
        );
        await _tokenStorage.save(access: _jwtWithSub(1), refresh: 'refresh');
        final authBloc = AuthBloc(auth: _MockAuthRepository())
          ..emit(AuthAuthenticated());

        await tester.pumpWidget(
          _wrap(
            const MarketplaceHomeScreen(),
            apiClient: apiClient,
            authBloc: authBloc,
          ),
        );
        await tester.pumpAndSettle();

        expect(find.byIcon(Icons.bookmark_border), findsNothing);
      },
    );

    testWidgets(
      'authenticated: tapping the "Favorites" icon opens FavoritesScreen '
      'and refreshes the feed on return',
      (tester) async {
        final adapter = _ListingsHttpClientAdapter(favoritesCount: 1);
        final apiClient = await _fakeApiClientWithListings(adapter: adapter);
        await _tokenStorage.save(access: _jwtWithSub(1), refresh: 'refresh');
        final authBloc = AuthBloc(auth: _MockAuthRepository())
          ..emit(AuthAuthenticated());

        await tester.pumpWidget(
          _wrap(
            const MarketplaceHomeScreen(),
            apiClient: apiClient,
            authBloc: authBloc,
          ),
        );
        await tester.pumpAndSettle();

        expect(adapter.listingsRequestCount, 2);
        expect(find.byIcon(Icons.bookmark_border), findsOneWidget);

        await tester.tap(find.byIcon(Icons.bookmark_border));
        await tester.pumpAndSettle();

        expect(find.byType(FavoritesScreen), findsOneWidget);

        await tester.pageBack();
        await tester.pumpAndSettle();

        expect(find.byType(FavoritesScreen), findsNothing);
        // MarketplaceHomeScreen reloads its feed when the pushed route returns.
        expect(adapter.listingsRequestCount, 4);
      },
    );
  });
}
