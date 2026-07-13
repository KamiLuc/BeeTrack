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
import 'package:app/features/marketplace/view/listing_detail_screen.dart';
import 'package:app/features/marketplace/view/marketplace_home_screen.dart';
import 'package:app/l10n/app_localizations.dart';

class _MockAuthRepository extends Mock implements AuthRepository {}

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

Future<ApiClient> _fakeApiClient() async {
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();
  final apiClient = ApiClient(storage: TokenStorage(prefs), baseUrl: 'http://test');
  apiClient.dio.httpClientAdapter = _FailingHttpClientAdapter();
  return apiClient;
}

/// Returns a single listing for `/api/v1/listings` so tap-to-navigate can be
/// exercised without a real backend.
class _ListingsHttpClientAdapter implements HttpClientAdapter {
  _ListingsHttpClientAdapter({this.price = 20.0});

  final double? price;
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
    if (options.path.contains('/listings') && options.method == 'GET') {
      listingsRequestCount++;
      listingsRequests.add(options);
    }
    if (options.path.contains('/listings')) {
      return ResponseBody.fromString(
        jsonEncode({
          'items': [
            {
              'id': 7,
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
            }
          ],
          'total': 1,
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

Future<ApiClient> _fakeApiClientWithListings({_ListingsHttpClientAdapter? adapter}) async {
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();
  final apiClient = ApiClient(storage: TokenStorage(prefs), baseUrl: 'http://test');
  apiClient.dio.httpClientAdapter = adapter ?? _ListingsHttpClientAdapter();
  return apiClient;
}

Widget _wrap(
  Widget child, {
  required ApiClient apiClient,
  AuthBloc? authBloc,
}) =>
    RepositoryProvider<ApiClient>.value(
      value: apiClient,
      child: BlocProvider<AuthBloc>.value(
        value: authBloc ?? AuthBloc(auth: _MockAuthRepository()),
        child: MaterialApp(
          localizationsDelegates: AppLocalizations.localizationsDelegates,
          supportedLocales: AppLocalizations.supportedLocales,
          locale: const Locale('en'),
          home: child,
        ),
      ),
    );

void main() {
  group('MarketplaceHomeScreen', () {
    testWidgets('unauthenticated: shows marketplace and drawer with login option',
        (tester) async {
      final apiClient = await _fakeApiClient();
      final authBloc = AuthBloc(auth: _MockAuthRepository());
      await tester.pumpWidget(_wrap(
        const MarketplaceHomeScreen(),
        apiClient: apiClient,
        authBloc: authBloc,
      ));
      await tester.pumpAndSettle();

      expect(find.text('Marketplace'), findsOneWidget);

      await tester.tap(find.byTooltip('Open navigation menu'));
      await tester.pumpAndSettle();

      expect(find.widgetWithText(ListTile, 'Log in'), findsOneWidget);
      expect(find.widgetWithText(ListTile, 'Apiaries'), findsOneWidget);
    });

    testWidgets(
        'unauthenticated: while browsing, Marketplace tile is selected and '
        'locked Apiaries tile is not, and tapping Apiaries calls onLogin',
        (tester) async {
      final apiClient = await _fakeApiClient();
      final authBloc = AuthBloc(auth: _MockAuthRepository());
      var loginTapped = false;
      await tester.pumpWidget(_wrap(
        MarketplaceHomeScreen(onLogin: () => loginTapped = true),
        apiClient: apiClient,
        authBloc: authBloc,
      ));
      await tester.pumpAndSettle();

      await tester.tap(find.byTooltip('Open navigation menu'));
      await tester.pumpAndSettle();

      final apiariesTile =
          tester.widget<ListTile>(find.widgetWithText(ListTile, 'Apiaries'));
      expect(apiariesTile.selected, isFalse);

      final marketplaceTile = tester
          .widget<ListTile>(find.widgetWithText(ListTile, 'Marketplace'));
      expect(marketplaceTile.selected, isTrue);

      await tester.tap(find.widgetWithText(ListTile, 'Apiaries'));
      await tester.pumpAndSettle();

      expect(loginTapped, isTrue);
    });

    testWidgets('unauthenticated: tapping Log in invokes onLogin callback',
        (tester) async {
      final apiClient = await _fakeApiClient();
      final authBloc = AuthBloc(auth: _MockAuthRepository());
      var loginTapped = false;
      await tester.pumpWidget(_wrap(
        MarketplaceHomeScreen(onLogin: () => loginTapped = true),
        apiClient: apiClient,
        authBloc: authBloc,
      ));
      await tester.pumpAndSettle();

      await tester.tap(find.byTooltip('Open navigation menu'));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(ListTile, 'Log in'));
      await tester.pumpAndSettle();

      expect(loginTapped, isTrue);
    });

    testWidgets('authenticated: shows marketplace with apiaries option',
        (tester) async {
      final apiClient = await _fakeApiClient();
      final authBloc = AuthBloc(auth: _MockAuthRepository())
        ..emit(AuthAuthenticated());

      await tester.pumpWidget(_wrap(
        const MarketplaceHomeScreen(),
        apiClient: apiClient,
        authBloc: authBloc,
      ));
      await tester.pumpAndSettle();

      expect(find.text('Marketplace'), findsOneWidget);

      await tester.tap(find.byTooltip('Open navigation menu'));
      await tester.pumpAndSettle();

      expect(find.widgetWithText(ListTile, 'Apiaries'), findsOneWidget);
      expect(find.widgetWithText(ListTile, 'Log out'), findsOneWidget);
    });

    testWidgets('re-selecting marketplace just closes drawer',
        (tester) async {
      final apiClient = await _fakeApiClient();
      final authBloc = AuthBloc(auth: _MockAuthRepository())
        ..emit(AuthAuthenticated());
      var selected = <AppSection>[];

      await tester.pumpWidget(_wrap(
        MarketplaceHomeScreen(onSelectSection: selected.add),
        apiClient: apiClient,
        authBloc: authBloc,
      ));
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

      await tester.pumpWidget(_wrap(
        MarketplaceHomeScreen(onSelectSection: selected.add),
        apiClient: apiClient,
        authBloc: authBloc,
      ));
      await tester.pumpAndSettle();

      await tester.tap(find.byTooltip('Open navigation menu'));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(ListTile, 'Apiaries'));
      await tester.pumpAndSettle();

      expect(selected, [AppSection.apiaries]);
    });

    testWidgets('authenticated: tapping Log out dispatches LogoutRequested',
        (tester) async {
      final apiClient = await _fakeApiClient();
      final repo = _MockAuthRepository();
      when(() => repo.logout()).thenAnswer((_) async {});
      final authBloc = AuthBloc(auth: repo)..emit(AuthAuthenticated());

      await tester.pumpWidget(_wrap(
        const MarketplaceHomeScreen(),
        apiClient: apiClient,
        authBloc: authBloc,
      ));
      await tester.pumpAndSettle();

      await tester.tap(find.byTooltip('Open navigation menu'));
      await tester.pumpAndSettle();

      await tester.tap(find.widgetWithText(ListTile, 'Log out'));
      await tester.pumpAndSettle();

      expect(authBloc.state, isA<AuthUnauthenticated>());
    });

    testWidgets('shows retry option when listings fail to load', (tester) async {
      final apiClient = await _fakeApiClient();
      final authBloc = AuthBloc(auth: _MockAuthRepository());

      await tester.pumpWidget(_wrap(
        const MarketplaceHomeScreen(),
        apiClient: apiClient,
        authBloc: authBloc,
      ));
      await tester.pumpAndSettle();

      expect(find.text('Something went wrong. Please try again.'), findsOneWidget);
      expect(find.widgetWithText(ElevatedButton, 'Retry'), findsOneWidget);
    });

    testWidgets('tapping a listing card navigates to ListingDetailScreen',
        (tester) async {
      final apiClient = await _fakeApiClientWithListings();
      final authBloc = AuthBloc(auth: _MockAuthRepository());

      await tester.pumpWidget(_wrap(
        const MarketplaceHomeScreen(),
        apiClient: apiClient,
        authBloc: authBloc,
      ));
      await tester.pumpAndSettle();

      expect(find.text('Wildflower Honey'), findsOneWidget);

      await tester.tap(find.text('Wildflower Honey'));
      await tester.pumpAndSettle();

      expect(find.byType(ListingDetailScreen), findsOneWidget);
    });

    testWidgets('listing card shows "Free" instead of "0.00" when price is 0',
        (tester) async {
      final apiClient =
          await _fakeApiClientWithListings(adapter: _ListingsHttpClientAdapter(price: 0));
      final authBloc = AuthBloc(auth: _MockAuthRepository());

      await tester.pumpWidget(_wrap(
        const MarketplaceHomeScreen(),
        apiClient: apiClient,
        authBloc: authBloc,
      ));
      await tester.pumpAndSettle();

      expect(find.text('Free'), findsOneWidget);
      expect(find.text('0.00'), findsNothing);
    });

    testWidgets('unauthenticated: the create-listing "+" button is not shown',
        (tester) async {
      final apiClient = await _fakeApiClient();
      final authBloc = AuthBloc(auth: _MockAuthRepository());

      await tester.pumpWidget(_wrap(
        const MarketplaceHomeScreen(),
        apiClient: apiClient,
        authBloc: authBloc,
      ));
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.add), findsNothing);
    });

    testWidgets(
        'authenticated: tapping the "+" button opens CreateListingScreen '
        'and refreshes the feed on return', (tester) async {
      final adapter = _ListingsHttpClientAdapter();
      final apiClient = await _fakeApiClientWithListings(adapter: adapter);
      final authBloc = AuthBloc(auth: _MockAuthRepository())
        ..emit(AuthAuthenticated());

      await tester.pumpWidget(_wrap(
        const MarketplaceHomeScreen(),
        apiClient: apiClient,
        authBloc: authBloc,
      ));
      await tester.pumpAndSettle();

      expect(adapter.listingsRequestCount, 1);

      await tester.tap(find.byIcon(Icons.add));
      await tester.pumpAndSettle();

      expect(find.byType(CreateListingScreen), findsOneWidget);

      await tester.pageBack();
      await tester.pumpAndSettle();

      expect(find.byType(CreateListingScreen), findsNothing);
      expect(adapter.listingsRequestCount, 2);
    });

    testWidgets(
        'the map button is shown in the bottom banner regardless of auth '
        'and stays disabled', (tester) async {
      final apiClient = await _fakeApiClient();
      final authBloc = AuthBloc(auth: _MockAuthRepository());

      await tester.pumpWidget(_wrap(
        const MarketplaceHomeScreen(),
        apiClient: apiClient,
        authBloc: authBloc,
      ));
      await tester.pumpAndSettle();

      final mapButton = tester.widget<IconButton>(
        find.widgetWithIcon(IconButton, Icons.map_outlined),
      );
      expect(mapButton.onPressed, isNull);
    });

    testWidgets(
        'authenticated: the "+" button is shown next to the map button in '
        'the bottom banner', (tester) async {
      final apiClient = await _fakeApiClient();
      final authBloc = AuthBloc(auth: _MockAuthRepository())
        ..emit(AuthAuthenticated());

      await tester.pumpWidget(_wrap(
        const MarketplaceHomeScreen(),
        apiClient: apiClient,
        authBloc: authBloc,
      ));
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.add), findsOneWidget);
      expect(find.byIcon(Icons.map_outlined), findsOneWidget);
    });

    testWidgets(
        'typing in the search field debounces and calls setKeyword without '
        'pressing enter', (tester) async {
      final adapter = _ListingsHttpClientAdapter();
      final apiClient = await _fakeApiClientWithListings(adapter: adapter);
      final authBloc = AuthBloc(auth: _MockAuthRepository());

      await tester.pumpWidget(_wrap(
        const MarketplaceHomeScreen(),
        apiClient: apiClient,
        authBloc: authBloc,
      ));
      await tester.pumpAndSettle();

      expect(adapter.listingsRequestCount, 1);

      await tester.enterText(find.byType(TextField), 'honey');
      await tester.pump(const Duration(milliseconds: 200));
      expect(adapter.listingsRequestCount, 1);

      await tester.pump(const Duration(milliseconds: 250));
      await tester.pump();
      expect(adapter.listingsRequestCount, 2);
      expect(adapter.listingsRequests.last.queryParameters['keyword'], 'honey');
    });
  });
}
