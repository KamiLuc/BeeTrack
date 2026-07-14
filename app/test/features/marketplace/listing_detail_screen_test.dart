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

/// Solves the delete-confirmation math puzzle (see `showDeleteDialog` /
/// `withPuzzle: true`) by reading the displayed "a + b = " prompt, typing
/// the sum, and tapping the confirm button.
Future<void> _solveDeletePuzzle(
  WidgetTester tester,
  AppLocalizations l10n,
) async {
  final prompt = tester
      .widgetList<Text>(find.byType(Text))
      .firstWhere((t) => RegExp(r'^\d+ \+ \d+ = $').hasMatch(t.data ?? ''));
  final match = RegExp(r'^(\d+) \+ (\d+) = $').firstMatch(prompt.data!)!;
  final sum = int.parse(match.group(1)!) + int.parse(match.group(2)!);
  await tester.enterText(find.byType(TextField), '$sum');
  await tester.tap(find.text(l10n.generalDelete).last);
  await tester.pumpAndSettle();
}

/// Builds a minimal (unsigned) JWT carrying the given `sub` claim, so
/// [TokenStorage.userId] can be exercised without a real auth flow.
String _jwtWithSub(int sub) {
  String segment(Object data) =>
      base64Url.encode(utf8.encode(jsonEncode(data))).replaceAll('=', '');
  return '${segment({'alg': 'none'})}.${segment({'sub': sub})}.signature';
}

/// Records requests and returns configurable responses so favorite
/// add/remove/list calls can be asserted and success/failure controlled.
class _RecordingAdapter implements HttpClientAdapter {
  final List<RequestOptions> requests = [];
  List<Map<String, dynamic>> favoriteItems = [];
  bool failMutations = false;
  Map<String, dynamic>? refetchedListingJson;

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

    if (RegExp(r'/listings/\d+/favorite$').hasMatch(options.path) &&
        options.method == 'GET') {
      return ResponseBody.fromString(
        jsonEncode({'is_favorite': favoriteItems.isNotEmpty}),
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

    if (RegExp(r'/listings/\d+$').hasMatch(options.path) &&
        options.method == 'GET' &&
        refetchedListingJson != null) {
      return ResponseBody.fromString(
        jsonEncode(refetchedListingJson),
        200,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );
    }

    if (RegExp(r'/listings/\d+$').hasMatch(options.path) &&
        options.method == 'DELETE') {
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

    if (RegExp(r'/listings/\d+$').hasMatch(options.path) &&
        options.method == 'PATCH') {
      return ResponseBody.fromString(
        jsonEncode({
          ..._listingJson(_listing()),
          ...(options.data as Map<String, dynamic>),
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

  @override
  void close({bool force = false}) {}
}

/// The [TokenStorage] backing the current test's [ApiClient], also handed to
/// [_wrap] via a [RepositoryProvider] so `context.read<TokenStorage>()`
/// resolves just like it does in the real widget tree (see main.dart).
late TokenStorage _tokenStorage;

/// [userId], when given, is embedded as the `sub` claim of a fake access
/// token so ListingDetailScreen's "am I the owner" check can be exercised.
Future<(ApiClient, _RecordingAdapter)> _fakeApiClient({int? userId}) async {
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();
  _tokenStorage = TokenStorage(prefs);
  if (userId != null) {
    await _tokenStorage.save(access: _jwtWithSub(userId), refresh: 'refresh');
  }
  final apiClient = ApiClient(storage: _tokenStorage, baseUrl: 'http://test');
  final adapter = _RecordingAdapter();
  apiClient.dio.httpClientAdapter = adapter;
  return (apiClient, adapter);
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

Listing _listing({
  int id = 1,
  String? apiaryName,
  List<ListingImage> images = const [],
  double? price = 42.5,
  String category = 'honey',
}) => Listing(
  id: id,
  userId: 5,
  title: 'Wildflower Honey',
  description: 'Fresh honey from the meadow.',
  category: category,
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
    testWidgets(
      'renders listing title, price, chips and description; contact is '
      'revealed on tap',
      (tester) async {
        final (apiClient, _) = await _fakeApiClient();
        final listing = _listing();

        await tester.pumpWidget(
          _wrap(ListingDetailScreen(listing: listing), apiClient: apiClient),
        );
        await tester.pumpAndSettle();

        expect(find.text('Wildflower Honey'), findsWidgets);
        expect(find.text('42.50'), findsOneWidget);
        expect(find.text('Krakow'), findsOneWidget);
        expect(find.textContaining('10 jars'), findsOneWidget);
        expect(find.text('Fresh honey from the meadow.'), findsOneWidget);

        expect(find.text('123456789'), findsNothing);
        expect(find.text('seller@example.com'), findsNothing);

        await tester.tap(find.widgetWithText(OutlinedButton, 'Call'));
        await tester.pumpAndSettle();
        expect(find.text('123456789'), findsOneWidget);

        await tester.tap(find.widgetWithText(OutlinedButton, 'Write'));
        await tester.pumpAndSettle();
        expect(find.text('seller@example.com'), findsOneWidget);
      },
    );

    testWidgets('shows "Free" instead of "0.00" when price is 0', (
      tester,
    ) async {
      final (apiClient, _) = await _fakeApiClient();

      await tester.pumpWidget(
        _wrap(
          ListingDetailScreen(listing: _listing(price: 0)),
          apiClient: apiClient,
        ),
      );
      await tester.pumpAndSettle();

      expect(find.text('Free'), findsOneWidget);
      expect(find.text('0.00'), findsNothing);
    });

    testWidgets('shows apiary section only when apiaryName is set', (
      tester,
    ) async {
      final (apiClient, _) = await _fakeApiClient();

      await tester.pumpWidget(
        _wrap(
          ListingDetailScreen(listing: _listing(apiaryName: 'Sunny Meadow')),
          apiClient: apiClient,
        ),
      );
      await tester.pumpAndSettle();

      expect(find.text('Sunny Meadow'), findsOneWidget);

      final (apiClient2, _) = await _fakeApiClient();
      await tester.pumpWidget(
        _wrap(ListingDetailScreen(listing: _listing()), apiClient: apiClient2),
      );
      await tester.pumpAndSettle();

      expect(find.text('Sunny Meadow'), findsNothing);
    });

    testWidgets('shows placeholder icon when listing has no images', (
      tester,
    ) async {
      final (apiClient, _) = await _fakeApiClient();

      await tester.pumpWidget(
        _wrap(ListingDetailScreen(listing: _listing()), apiClient: apiClient),
      );
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.image_not_supported_outlined), findsOneWidget);
      expect(find.byType(PageView), findsNothing);
    });

    testWidgets(
      'shows carousel with nav arrows and expand button for multiple images',
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

        await tester.pumpWidget(
          _wrap(
            ListingDetailScreen(listing: _listing(images: images)),
            apiClient: apiClient,
          ),
        );
        await tester.pumpAndSettle();

        expect(find.byType(PageView), findsOneWidget);
        expect(find.byIcon(Icons.image_not_supported_outlined), findsNothing);
        expect(find.byIcon(Icons.chevron_left), findsOneWidget);
        expect(find.byIcon(Icons.chevron_right), findsOneWidget);
        expect(find.byIcon(Icons.open_in_full), findsOneWidget);

        await tester.tap(find.byIcon(Icons.open_in_full));
        await tester.pumpAndSettle();

        expect(find.byIcon(Icons.close), findsOneWidget);
        await tester.tap(find.byIcon(Icons.close));
        await tester.pumpAndSettle();
      },
    );

    testWidgets('unauthenticated: no favorite action shown in AppBar', (
      tester,
    ) async {
      final (apiClient, _) = await _fakeApiClient();
      final authBloc = AuthBloc(auth: _MockAuthRepository());

      await tester.pumpWidget(
        _wrap(
          ListingDetailScreen(listing: _listing()),
          apiClient: apiClient,
          authBloc: authBloc,
        ),
      );
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.favorite), findsNothing);
      expect(find.byIcon(Icons.favorite_border), findsNothing);
    });

    testWidgets(
      'authenticated: shows favorite icon and reflects existing favorite',
      (tester) async {
        final (apiClient, adapter) = await _fakeApiClient();
        final listing = _listing();
        adapter.favoriteItems = [_listingJson(listing)];
        final authBloc = AuthBloc(auth: _MockAuthRepository())
          ..emit(AuthAuthenticated());

        await tester.pumpWidget(
          _wrap(
            ListingDetailScreen(listing: listing),
            apiClient: apiClient,
            authBloc: authBloc,
          ),
        );
        await tester.pumpAndSettle();

        expect(find.byIcon(Icons.favorite), findsOneWidget);
      },
    );

    testWidgets(
      'authenticated: tapping favorite toggles icon and calls add endpoint',
      (tester) async {
        final (apiClient, adapter) = await _fakeApiClient();
        final authBloc = AuthBloc(auth: _MockAuthRepository())
          ..emit(AuthAuthenticated());

        await tester.pumpWidget(
          _wrap(
            ListingDetailScreen(listing: _listing()),
            apiClient: apiClient,
            authBloc: authBloc,
          ),
        );
        await tester.pumpAndSettle();

        expect(find.byIcon(Icons.favorite_border), findsOneWidget);

        await tester.tap(find.byIcon(Icons.favorite_border));
        await tester.pumpAndSettle();

        expect(find.byIcon(Icons.favorite), findsOneWidget);
        expect(
          adapter.requests.any(
            (r) =>
                r.method == 'POST' && r.path.contains('/listings/1/favorite'),
          ),
          isTrue,
        );
      },
    );

    testWidgets(
      'authenticated: failed toggle rolls back icon and shows snackbar',
      (tester) async {
        final (apiClient, adapter) = await _fakeApiClient();
        adapter.failMutations = true;
        final authBloc = AuthBloc(auth: _MockAuthRepository())
          ..emit(AuthAuthenticated());

        await tester.pumpWidget(
          _wrap(
            ListingDetailScreen(listing: _listing()),
            apiClient: apiClient,
            authBloc: authBloc,
          ),
        );
        await tester.pumpAndSettle();

        await tester.tap(find.byIcon(Icons.favorite_border));
        await tester.pumpAndSettle();

        expect(find.byIcon(Icons.favorite_border), findsOneWidget);
        expect(find.byType(SnackBar), findsOneWidget);
      },
    );

    testWidgets('non-owner: no edit or delete button shown', (tester) async {
      final (apiClient, _) = await _fakeApiClient(userId: 99);

      await tester.pumpWidget(
        _wrap(ListingDetailScreen(listing: _listing()), apiClient: apiClient),
      );
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.edit_outlined), findsNothing);
      expect(find.byIcon(Icons.delete_outline), findsNothing);
    });

    testWidgets('owner: no favorite icon shown — cannot favorite own listing', (
      tester,
    ) async {
      final (apiClient, _) = await _fakeApiClient(userId: 5);
      final authBloc = AuthBloc(auth: _MockAuthRepository())
        ..emit(AuthAuthenticated());

      await tester.pumpWidget(
        _wrap(
          ListingDetailScreen(listing: _listing()),
          apiClient: apiClient,
          authBloc: authBloc,
        ),
      );
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.favorite_border), findsNothing);
      expect(find.byIcon(Icons.favorite), findsNothing);
    });

    testWidgets('non-owner: favorite icon is shown', (tester) async {
      final (apiClient, _) = await _fakeApiClient(userId: 99);
      final authBloc = AuthBloc(auth: _MockAuthRepository())
        ..emit(AuthAuthenticated());

      await tester.pumpWidget(
        _wrap(
          ListingDetailScreen(listing: _listing()),
          apiClient: apiClient,
          authBloc: authBloc,
        ),
      );
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.favorite_border), findsOneWidget);
    });

    testWidgets(
      'owner: shows edit button at the bottom, opens CreateListingScreen '
      'prefilled, and refetches the listing after a successful edit',
      (tester) async {
        final (apiClient, adapter) = await _fakeApiClient(userId: 5);
        final listing = _listing(category: 'HONEY');
        adapter.refetchedListingJson = _listingJson(listing).map(
          (key, value) =>
              MapEntry(key, key == 'title' ? 'Updated Title' : value),
        );

        await tester.pumpWidget(
          _wrap(ListingDetailScreen(listing: listing), apiClient: apiClient),
        );
        await tester.pumpAndSettle();

        expect(find.byIcon(Icons.edit_outlined), findsOneWidget);
        expect(
          find.descendant(
            of: find.byType(AppBar),
            matching: find.byIcon(Icons.edit_outlined),
          ),
          findsNothing,
        );
        expect(
          find.byWidgetPredicate(
            (w) => w is Scaffold && w.bottomNavigationBar != null,
          ),
          findsNothing,
        );

        await tester.tap(find.byIcon(Icons.edit_outlined));
        await tester.pumpAndSettle();

        expect(find.text('Edit listing'), findsOneWidget);
        expect(find.text('Wildflower Honey'), findsWidgets);

        final saveFinder = find.widgetWithText(ElevatedButton, 'Save');
        await tester.ensureVisible(saveFinder);
        await tester.tap(saveFinder);
        await tester.pumpAndSettle();

        expect(find.text('Updated Title'), findsWidgets);
      },
    );

    testWidgets(
      'owner: carousel resets to first image after editing photos, so nav '
      'arrows and image shown stay in sync with the refreshed listing',
      (tester) async {
        final (apiClient, adapter) = await _fakeApiClient(userId: 5);
        final oldImages = [
          ListingImage(
            id: 1,
            listingId: 1,
            url: '/uploads/old-a.jpg',
            displayOrder: 0,
            createdAt: DateTime(2026, 1, 1),
          ),
          ListingImage(
            id: 2,
            listingId: 1,
            url: '/uploads/old-b.jpg',
            displayOrder: 1,
            createdAt: DateTime(2026, 1, 1),
          ),
        ];
        final newImages = [
          ListingImage(
            id: 3,
            listingId: 1,
            url: '/uploads/new-a.jpg',
            displayOrder: 0,
            createdAt: DateTime(2026, 6, 1),
          ),
          ListingImage(
            id: 4,
            listingId: 1,
            url: '/uploads/new-b.jpg',
            displayOrder: 1,
            createdAt: DateTime(2026, 6, 1),
          ),
        ];
        final listing = _listing(images: oldImages, category: 'HONEY');
        adapter.refetchedListingJson = {
          ..._listingJson(listing),
          'images': newImages
              .map(
                (img) => {
                  'id': img.id,
                  'listing_id': img.listingId,
                  'url': img.url,
                  'display_order': img.displayOrder,
                  'created_at': img.createdAt.toIso8601String(),
                },
              )
              .toList(),
        };

        await tester.pumpWidget(
          _wrap(ListingDetailScreen(listing: listing), apiClient: apiClient),
        );
        await tester.pumpAndSettle();

        expect(find.image(NetworkImage('http://test/uploads/old-a.jpg')),
            findsOneWidget);

        await tester.tap(find.byIcon(Icons.chevron_right));
        await tester.pumpAndSettle();

        expect(find.image(NetworkImage('http://test/uploads/old-b.jpg')),
            findsOneWidget);
        final leftArrowBefore = tester.widget<InkWell>(
          find.ancestor(
            of: find.byIcon(Icons.chevron_left),
            matching: find.byType(InkWell),
          ),
        );
        expect(leftArrowBefore.onTap, isNotNull);

        await tester.tap(find.byIcon(Icons.edit_outlined));
        await tester.pumpAndSettle();

        final saveFinder = find.widgetWithText(ElevatedButton, 'Save');
        await tester.ensureVisible(saveFinder);
        await tester.tap(saveFinder);
        await tester.pumpAndSettle();

        expect(find.image(NetworkImage('http://test/uploads/new-a.jpg')),
            findsOneWidget);
        expect(find.image(NetworkImage('http://test/uploads/old-a.jpg')),
            findsNothing);
        expect(find.image(NetworkImage('http://test/uploads/old-b.jpg')),
            findsNothing);

        final leftArrowAfter = tester.widget<InkWell>(
          find.ancestor(
            of: find.byIcon(Icons.chevron_left),
            matching: find.byType(InkWell),
          ),
        );
        expect(leftArrowAfter.onTap, isNull);
      },
    );

    testWidgets(
      'owner: delete button requires solving the confirmation puzzle, then '
      'pops back',
      (tester) async {
        final l10n = await AppLocalizations.delegate.load(const Locale('en'));
        final (apiClient, adapter) = await _fakeApiClient(userId: 5);
        final listing = _listing();

        await tester.pumpWidget(
          _wrap(
            Builder(
              builder: (context) => Scaffold(
                body: Center(
                  child: ElevatedButton(
                    onPressed: () => Navigator.of(context).push(
                      MaterialPageRoute(
                        builder: (_) => ListingDetailScreen(listing: listing),
                      ),
                    ),
                    child: const Text('open'),
                  ),
                ),
              ),
            ),
            apiClient: apiClient,
          ),
        );
        await tester.pumpAndSettle();

        await tester.tap(find.text('open'));
        await tester.pumpAndSettle();
        expect(find.byType(ListingDetailScreen), findsOneWidget);

        await tester.tap(find.byIcon(Icons.delete_outline));
        await tester.pumpAndSettle();

        expect(find.text(l10n.marketplaceDeleteConfirm), findsOneWidget);

        // Cancelling must not delete anything.
        await tester.tap(find.text(l10n.generalCancel));
        await tester.pumpAndSettle();
        expect(find.byType(ListingDetailScreen), findsOneWidget);
        expect(adapter.requests.any((r) => r.method == 'DELETE'), isFalse);

        await tester.tap(find.byIcon(Icons.delete_outline));
        await tester.pumpAndSettle();
        await _solveDeletePuzzle(tester, l10n);

        expect(find.byType(ListingDetailScreen), findsNothing);
        expect(
          adapter.requests.any(
            (r) =>
                r.method == 'DELETE' &&
                r.path.contains('/listings/${listing.id}'),
          ),
          isTrue,
        );
      },
    );

    testWidgets(
      'owner: wrong puzzle answer shows an error and does not delete',
      (tester) async {
        final l10n = await AppLocalizations.delegate.load(const Locale('en'));
        final (apiClient, adapter) = await _fakeApiClient(userId: 5);
        final listing = _listing();

        await tester.pumpWidget(
          _wrap(ListingDetailScreen(listing: listing), apiClient: apiClient),
        );
        await tester.pumpAndSettle();

        await tester.tap(find.byIcon(Icons.delete_outline));
        await tester.pumpAndSettle();

        await tester.enterText(find.byType(TextField), '-1');
        await tester.tap(find.text(l10n.generalDelete).last);
        await tester.pumpAndSettle();

        expect(find.text(l10n.deletePuzzleWrong), findsOneWidget);
        expect(find.byType(ListingDetailScreen), findsOneWidget);
        expect(adapter.requests.any((r) => r.method == 'DELETE'), isFalse);
      },
    );

    testWidgets(
      'owner: delete failure shows an error snackbar and stays on screen',
      (tester) async {
        final l10n = await AppLocalizations.delegate.load(const Locale('en'));
        final (apiClient, adapter) = await _fakeApiClient(userId: 5);
        adapter.failMutations = true;
        final listing = _listing();

        await tester.pumpWidget(
          _wrap(ListingDetailScreen(listing: listing), apiClient: apiClient),
        );
        await tester.pumpAndSettle();

        await tester.tap(find.byIcon(Icons.delete_outline));
        await tester.pumpAndSettle();
        await _solveDeletePuzzle(tester, l10n);

        expect(find.byType(ListingDetailScreen), findsOneWidget);
        expect(find.text(l10n.generalError), findsOneWidget);
      },
    );
  });
}
