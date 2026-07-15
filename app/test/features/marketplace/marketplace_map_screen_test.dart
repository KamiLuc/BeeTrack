import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:app/core/api/api_client.dart';
import 'package:app/core/storage/token_storage.dart';
import 'package:app/features/auth/bloc/auth_bloc.dart';
import 'package:app/features/auth/data/auth_repository.dart';
import 'package:app/features/marketplace/data/listing_model.dart';
import 'package:app/features/marketplace/view/listing_detail_screen.dart';
import 'package:app/features/marketplace/view/marketplace_map_screen.dart';
import 'package:app/l10n/app_localizations.dart';

class _MockAuthRepository extends Mock implements AuthRepository {}

/// Fails every request synchronously so navigating into ListingDetailScreen
/// (which fetches favorite status when authenticated) never leaves a
/// pending network timer behind.
class _FailingHttpClientAdapter implements HttpClientAdapter {
  @override
  void close({bool force = false}) {}

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<List<int>>? requestStream,
    Future<void>? cancelFuture,
  ) {
    throw DioException(requestOptions: options, message: 'no network in tests');
  }
}

Listing _listing({
  required int id,
  required String title,
  double lat = 0,
  double lng = 0,
}) {
  final now = DateTime(2026, 1, 1);
  return Listing(
    id: id,
    userId: 1,
    title: title,
    description: 'Fresh honey.',
    category: 'honey',
    price: 20,
    quantity: '5 jars',
    address: 'Krakow',
    lat: lat,
    lng: lng,
    contactPhone: '123456789',
    contactEmail: 'seller@example.com',
    isHidden: false,
    createdAt: now,
    updatedAt: now,
    images: const [],
  );
}

Future<Widget> _wrap(Widget child) async {
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();
  final tokenStorage = TokenStorage(prefs);
  final apiClient = ApiClient(storage: tokenStorage, baseUrl: 'http://test');
  apiClient.dio.httpClientAdapter = _FailingHttpClientAdapter();

  return RepositoryProvider<ApiClient>.value(
    value: apiClient,
    child: RepositoryProvider<TokenStorage>.value(
      value: tokenStorage,
      child: BlocProvider<AuthBloc>.value(
        value: AuthBloc(auth: _MockAuthRepository()),
        child: MaterialApp(
          localizationsDelegates: AppLocalizations.localizationsDelegates,
          supportedLocales: AppLocalizations.supportedLocales,
          locale: const Locale('en'),
          home: child,
        ),
      ),
    ),
  );
}

void main() {
  group('MarketplaceMapScreen', () {
    testWidgets(
      'shows a marker for each located listing and excludes ones with '
      'lat=0/lng=0',
      (tester) async {
        final listings = [
          _listing(id: 1, title: 'Located Honey', lat: 50.06, lng: 19.94),
          _listing(id: 2, title: 'Unlocated Honey'),
        ];

        await tester.pumpWidget(
          await _wrap(MarketplaceMapScreen(listings: listings)),
        );
        await tester.pumpAndSettle();

        expect(find.byType(FlutterMap), findsOneWidget);
        expect(find.byIcon(Icons.location_pin), findsOneWidget);
        expect(find.text('Marketplace map'), findsOneWidget);

        final tooltip = tester.widget<Tooltip>(find.byType(Tooltip));
        expect(tooltip.message, 'Located Honey • 20.00');
      },
    );

    testWidgets(
      'shows the empty state when no listings have a real location',
      (tester) async {
        final listings = [
          _listing(id: 1, title: 'Unlocated Honey'),
          _listing(id: 2, title: 'Also Unlocated'),
        ];

        await tester.pumpWidget(
          await _wrap(MarketplaceMapScreen(listings: listings)),
        );
        await tester.pumpAndSettle();

        expect(
          find.text('No listings with a location match your filters'),
          findsOneWidget,
        );
        expect(find.byType(FlutterMap), findsNothing);
      },
    );

    testWidgets('shows the empty state when the listings list is empty', (
      tester,
    ) async {
      await tester.pumpWidget(
        await _wrap(const MarketplaceMapScreen(listings: [])),
      );
      await tester.pumpAndSettle();

      expect(
        find.text('No listings with a location match your filters'),
        findsOneWidget,
      );
    });

    testWidgets('tapping a marker navigates to ListingDetailScreen', (
      tester,
    ) async {
      final listing = _listing(
        id: 1,
        title: 'Located Honey',
        lat: 50.06,
        lng: 19.94,
      );

      await tester.pumpWidget(
        await _wrap(MarketplaceMapScreen(listings: [listing])),
      );
      await tester.pumpAndSettle();

      await tester.tap(find.byIcon(Icons.location_pin));
      await tester.pumpAndSettle();

      expect(find.byType(ListingDetailScreen), findsOneWidget);
    });
  });
}
