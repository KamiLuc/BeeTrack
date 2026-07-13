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
import 'package:app/features/auth/bloc/auth_bloc.dart';
import 'package:app/features/auth/data/auth_repository.dart';
import 'package:app/features/marketplace/data/listing_model.dart';
import 'package:app/features/marketplace/view/listing_detail_screen.dart';
import 'package:app/l10n/app_localizations.dart';

class _MockAuthRepository extends Mock implements AuthRepository {}

/// Records requests and returns configurable responses so favorite
/// add/remove/list calls can be asserted and success/failure controlled.
class _RecordingAdapter implements HttpClientAdapter {
  final List<RequestOptions> requests = [];
  List<Map<String, dynamic>> favoriteItems = [];
  bool failMutations = false;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    requests.add(options);

    if (options.path.contains('/favorites') && options.method == 'GET') {
      return ResponseBody.fromString(
        jsonEncode({'items': favoriteItems}),
        200,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );
    }

    if (options.path.contains('/favorite') &&
        (options.method == 'POST' || options.method == 'DELETE')) {
      if (failMutations) {
        throw DioException(
          requestOptions: options,
          response: Response(requestOptions: options, statusCode: 500),
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

    return ResponseBody.fromString(
      jsonEncode({}),
      200,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}

Future<(ApiClient, _RecordingAdapter)> _fakeApiClient() async {
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();
  final apiClient =
      ApiClient(storage: TokenStorage(prefs), baseUrl: 'http://test');
  final adapter = _RecordingAdapter();
  apiClient.dio.httpClientAdapter = adapter;
  return (apiClient, adapter);
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

Listing _listing({
  int id = 1,
  String? apiaryName,
  List<ListingImage> images = const [],
  double? price = 42.5,
}) =>
    Listing(
      id: id,
      userId: 5,
      title: 'Wildflower Honey',
      description: 'Fresh honey from the meadow.',
      category: 'honey',
      price: price,
      quantity: '10 jars',
      address: 'Krakow',
      apiaryName: apiaryName,
      contactPhone: '123456789',
      contactEmail: 'seller@example.com',
      isHidden: false,
      createdAt: DateTime(2026, 5, 1),
      updatedAt: DateTime(2026, 5, 1),
      images: images,
    );

Map<String, dynamic> _listingJson(Listing listing) => {
      'id': listing.id,
      'user_id': listing.userId,
      'title': listing.title,
      'description': listing.description,
      'category': listing.category,
      'price': listing.price,
      'quantity': listing.quantity,
      'address': listing.address,
      'apiary_name': listing.apiaryName,
      'contact_phone': listing.contactPhone,
      'contact_email': listing.contactEmail,
      'is_hidden': listing.isHidden,
      'created_at': listing.createdAt.toIso8601String(),
      'updated_at': listing.updatedAt.toIso8601String(),
      'images': const [],
    };

void main() {
  group('ListingDetailScreen', () {
    testWidgets('renders listing title, price, chips, description and contact',
        (tester) async {
      final (apiClient, _) = await _fakeApiClient();
      final listing = _listing();

      await tester.pumpWidget(_wrap(
        ListingDetailScreen(listing: listing),
        apiClient: apiClient,
      ));
      await tester.pumpAndSettle();

      expect(find.text('Wildflower Honey'), findsWidgets);
      expect(find.text('42.50 zł'), findsOneWidget);
      expect(find.text('Krakow'), findsOneWidget);
      expect(find.textContaining('10 jars'), findsOneWidget);
      expect(find.text('Fresh honey from the meadow.'), findsOneWidget);
      expect(find.text('123456789'), findsOneWidget);
      expect(find.text('seller@example.com'), findsOneWidget);
    });

    testWidgets('shows apiary section only when apiaryName is set',
        (tester) async {
      final (apiClient, _) = await _fakeApiClient();

      await tester.pumpWidget(_wrap(
        ListingDetailScreen(listing: _listing(apiaryName: 'Sunny Meadow')),
        apiClient: apiClient,
      ));
      await tester.pumpAndSettle();

      expect(find.text('Sunny Meadow'), findsOneWidget);

      final (apiClient2, _) = await _fakeApiClient();
      await tester.pumpWidget(_wrap(
        ListingDetailScreen(listing: _listing()),
        apiClient: apiClient2,
      ));
      await tester.pumpAndSettle();

      expect(find.text('Sunny Meadow'), findsNothing);
    });

    testWidgets('shows placeholder icon when listing has no images',
        (tester) async {
      final (apiClient, _) = await _fakeApiClient();

      await tester.pumpWidget(_wrap(
        ListingDetailScreen(listing: _listing()),
        apiClient: apiClient,
      ));
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.image_not_supported_outlined), findsOneWidget);
      expect(find.byType(PageView), findsNothing);
    });

    testWidgets('shows carousel with dot indicators for multiple images',
        (tester) async {
      final (apiClient, _) = await _fakeApiClient();
      final images = [
        ListingImage(
          id: 1,
          listingId: 1,
          url: '/uploads/a.jpg',
          displayOrder: 0,
          createdAt: DateTime(2026, 1, 1),
        ),
        ListingImage(
          id: 2,
          listingId: 1,
          url: '/uploads/b.jpg',
          displayOrder: 1,
          createdAt: DateTime(2026, 1, 1),
        ),
      ];

      await tester.pumpWidget(_wrap(
        ListingDetailScreen(listing: _listing(images: images)),
        apiClient: apiClient,
      ));
      await tester.pumpAndSettle();

      expect(find.byType(PageView), findsOneWidget);
      expect(find.byIcon(Icons.image_not_supported_outlined), findsNothing);
    });

    testWidgets('unauthenticated: no favorite action shown in AppBar',
        (tester) async {
      final (apiClient, _) = await _fakeApiClient();
      final authBloc = AuthBloc(auth: _MockAuthRepository());

      await tester.pumpWidget(_wrap(
        ListingDetailScreen(listing: _listing()),
        apiClient: apiClient,
        authBloc: authBloc,
      ));
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.favorite), findsNothing);
      expect(find.byIcon(Icons.favorite_border), findsNothing);
    });

    testWidgets('authenticated: shows favorite icon and reflects existing favorite',
        (tester) async {
      final (apiClient, adapter) = await _fakeApiClient();
      final listing = _listing();
      adapter.favoriteItems = [_listingJson(listing)];
      final authBloc = AuthBloc(auth: _MockAuthRepository())
        ..emit(AuthAuthenticated());

      await tester.pumpWidget(_wrap(
        ListingDetailScreen(listing: listing),
        apiClient: apiClient,
        authBloc: authBloc,
      ));
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.favorite), findsOneWidget);
    });

    testWidgets('authenticated: tapping favorite toggles icon and calls add endpoint',
        (tester) async {
      final (apiClient, adapter) = await _fakeApiClient();
      final authBloc = AuthBloc(auth: _MockAuthRepository())
        ..emit(AuthAuthenticated());

      await tester.pumpWidget(_wrap(
        ListingDetailScreen(listing: _listing()),
        apiClient: apiClient,
        authBloc: authBloc,
      ));
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.favorite_border), findsOneWidget);

      await tester.tap(find.byIcon(Icons.favorite_border));
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.favorite), findsOneWidget);
      expect(
        adapter.requests.any((r) =>
            r.method == 'POST' && r.path.contains('/listings/1/favorite')),
        isTrue,
      );
    });

    testWidgets('authenticated: failed toggle rolls back icon and shows snackbar',
        (tester) async {
      final (apiClient, adapter) = await _fakeApiClient();
      adapter.failMutations = true;
      final authBloc = AuthBloc(auth: _MockAuthRepository())
        ..emit(AuthAuthenticated());

      await tester.pumpWidget(_wrap(
        ListingDetailScreen(listing: _listing()),
        apiClient: apiClient,
        authBloc: authBloc,
      ));
      await tester.pumpAndSettle();

      await tester.tap(find.byIcon(Icons.favorite_border));
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.favorite_border), findsOneWidget);
      expect(find.byType(SnackBar), findsOneWidget);
    });
  });
}
