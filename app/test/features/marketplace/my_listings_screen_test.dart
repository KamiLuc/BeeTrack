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
  await tester.enterText(
    find.descendant(
      of: find.byType(AlertDialog),
      matching: find.byType(TextField),
    ),
    '$sum',
  );
  await tester.tap(find.text(l10n.generalDelete).last);
  await tester.pumpAndSettle();
}

/// Records requests and serves configurable responses for the endpoints
/// MyListingsScreen and its actions call.
class _RecordingAdapter implements HttpClientAdapter {
  final List<RequestOptions> requests = [];
  List<Map<String, dynamic>> listings = [];

  /// When set, overrides the response's `total` (which otherwise defaults to
  /// `listings.length`) so pagination can be exercised without needing 20+
  /// fixture listings.
  int? totalOverride;
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
      return _json({
        'items': listings,
        'total': totalOverride ?? listings.length,
      });
    }

    if (options.path.contains('/hide') && options.method == 'PATCH') {
      if (failMutations) {
        throw DioException(requestOptions: options, message: 'hide failed');
      }
      final id = int.parse(
        RegExp(r'/listings/(\d+)/hide').firstMatch(options.path)!.group(1)!,
      );
      final hidden = (options.data as Map<String, dynamic>)['hidden'] as bool;
      final listing = listings.firstWhere((l) => l['id'] == id);
      return _json({...listing, 'is_hidden': hidden});
    }

    if (RegExp(r'/listings/\d+$').hasMatch(options.path) &&
        options.method == 'DELETE') {
      if (failMutations) {
        throw DioException(requestOptions: options, message: 'delete failed');
      }
      final id = int.parse(
        RegExp(r'/listings/(\d+)$').firstMatch(options.path)!.group(1)!,
      );
      listings = listings.where((l) => l['id'] != id).toList();
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
  String status = 'approved',
  String? rejectionReason,
  Map<String, dynamic>? honeyBatch,
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
  'is_hidden': isHidden,
  'status': status,
  'rejection_reason': rejectionReason,
  'created_at': DateTime(2026, 1, 1).toIso8601String(),
  'updated_at': DateTime(2026, 1, 1).toIso8601String(),
  'images': [],
  if (honeyBatch != null) 'honey_batch_id': honeyBatch['id'],
  if (honeyBatch != null) 'honey_batch': honeyBatch,
};

void main() {
  late AppLocalizations l10n;
  setUpAll(() async {
    l10n = await AppLocalizations.delegate.load(const Locale('en'));
  });

  group('MyListingsScreen', () {
    testWidgets('shows the empty state when the user has no listings', (
      tester,
    ) async {
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
        await tester.tap(
          find.widgetWithText(ElevatedButton, l10n.generalRetry),
        );
        await tester.pumpAndSettle();

        expect(find.text('Wildflower Honey'), findsOneWidget);
      },
    );

    testWidgets(
      'on a narrow screen, the card switches to a stacked layout with the '
      'image on top, without overflowing',
      (tester) async {
        final originalSize = tester.view.physicalSize;
        final originalRatio = tester.view.devicePixelRatio;
        tester.view.physicalSize = const Size(360, 800);
        tester.view.devicePixelRatio = 1.0;
        addTearDown(() {
          tester.view.physicalSize = originalSize;
          tester.view.devicePixelRatio = originalRatio;
        });

        final (apiClient, adapter) = await _fakeApiClient();
        adapter.listings = [_listingJson()];

        await tester.pumpWidget(_wrap(apiClient));
        await tester.pumpAndSettle();

        expect(tester.takeException(), isNull);
        expect(find.text('Wildflower Honey'), findsOneWidget);
      },
    );

    testWidgets('shows the certified badge for a listing with a confirmed honey batch', (
      tester,
    ) async {
      final (apiClient, adapter) = await _fakeApiClient();
      adapter.listings = [
        _listingJson(
          honeyBatch: {
            'id': 3,
            'honey_type': 'Wildflower',
            'gathering_date': DateTime(2026, 1, 1).toIso8601String(),
            'amount_grams': 2500,
            'processing_method': 'raw',
            'certification_status': 'confirmed',
            'has_pdf': true,
            'verification_url': 'https://example.com/verify/tok',
          },
        ),
      ];

      await tester.pumpWidget(_wrap(apiClient));
      await tester.pumpAndSettle();

      expect(find.text(l10n.marketplaceCertifiedBadge), findsOneWidget);
    });

    testWidgets('renders a listing with the hidden badge when hidden', (
      tester,
    ) async {
      final (apiClient, adapter) = await _fakeApiClient();
      adapter.listings = [_listingJson(isHidden: true)];

      await tester.pumpWidget(_wrap(apiClient));
      await tester.pumpAndSettle();

      expect(find.text('Wildflower Honey'), findsOneWidget);
      expect(find.text(l10n.marketplaceHiddenBadge), findsOneWidget);
    });

    testWidgets('shows the live badge for an approved listing', (
      tester,
    ) async {
      final (apiClient, adapter) = await _fakeApiClient();
      adapter.listings = [_listingJson(status: 'approved')];

      await tester.pumpWidget(_wrap(apiClient));
      await tester.pumpAndSettle();

      expect(find.text(l10n.marketplaceStatusApproved), findsOneWidget);
    });

    testWidgets('shows the pending badge for a pending listing', (
      tester,
    ) async {
      final (apiClient, adapter) = await _fakeApiClient();
      adapter.listings = [_listingJson(status: 'pending')];

      await tester.pumpWidget(_wrap(apiClient));
      await tester.pumpAndSettle();

      expect(find.text(l10n.marketplaceStatusPending), findsOneWidget);
    });

    testWidgets(
      'shows the rejected badge with the reason as a tooltip',
      (tester) async {
        final (apiClient, adapter) = await _fakeApiClient();
        adapter.listings = [
          _listingJson(status: 'rejected', rejectionReason: 'Blurry photos'),
        ];

        await tester.pumpWidget(_wrap(apiClient));
        await tester.pumpAndSettle();

        expect(find.text(l10n.marketplaceStatusRejected), findsOneWidget);
        expect(
          find.byWidgetPredicate(
            (w) => w is Tooltip && w.message == 'Blurry photos',
          ),
          findsOneWidget,
        );
      },
    );

    testWidgets(
      'does not request more when everything fits in one load',
      (tester) async {
        final (apiClient, adapter) = await _fakeApiClient();
        adapter.listings = [_listingJson()];

        await tester.pumpWidget(_wrap(apiClient));
        await tester.pumpAndSettle();

        final requestCountBeforeScroll = adapter.requests
            .where((r) => r.path.endsWith('/listings') && r.method == 'GET')
            .length;

        await tester.drag(find.byType(ListView), const Offset(0, -100000));
        await tester.pumpAndSettle();

        final requestCountAfterScroll = adapter.requests
            .where((r) => r.path.endsWith('/listings') && r.method == 'GET')
            .length;
        expect(requestCountAfterScroll, requestCountBeforeScroll);
      },
    );

    testWidgets(
      'scrolling near the bottom of the list triggers loading more and '
      'requests the next offset',
      (tester) async {
        final (apiClient, adapter) = await _fakeApiClient();
        adapter.listings = [
          for (var i = 1; i <= 20; i++)
            _listingJson(id: i, title: 'Listing $i'),
        ];
        adapter.totalOverride = 45;

        await tester.pumpWidget(_wrap(apiClient));
        await tester.pump();
        await tester.pump(const Duration(milliseconds: 50));
        await tester.pump(const Duration(milliseconds: 50));

        await tester.drag(find.byType(ListView), const Offset(0, -100000));
        await tester.pump();
        await tester.pump(const Duration(milliseconds: 50));
        await tester.pump(const Duration(milliseconds: 50));

        final offsets = adapter.requests
            .where((r) => r.path.endsWith('/listings') && r.method == 'GET')
            .map((r) => r.queryParameters['offset'])
            .toList();
        expect(offsets, contains(20));
      },
    );

    testWidgets('tapping a card opens the listing detail view', (
      tester,
    ) async {
      final (apiClient, adapter) = await _fakeApiClient();
      adapter.listings = [_listingJson()];

      await tester.pumpWidget(_wrap(apiClient));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Wildflower Honey'));
      await tester.pumpAndSettle();

      expect(find.byType(ListingDetailScreen), findsOneWidget);
      expect(find.byType(CreateListingScreen), findsNothing);
    });

    testWidgets('add button opens CreateListingScreen for a new listing', (
      tester,
    ) async {
      final (apiClient, adapter) = await _fakeApiClient();
      adapter.listings = [_listingJson()];

      await tester.pumpWidget(_wrap(apiClient));
      await tester.pumpAndSettle();

      await tester.tap(find.byIcon(Icons.add));
      await tester.pumpAndSettle();

      expect(find.byType(CreateListingScreen), findsOneWidget);
      expect(find.text(l10n.marketplaceCreateScreenTitle), findsOneWidget);
    });

    testWidgets(
      'edit action opens CreateListingScreen prefilled and reloads on '
      'successful edit',
      (tester) async {
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
      },
    );

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

    testWidgets('delete action removes the card only after confirming', (
      tester,
    ) async {
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
      await _solveDeletePuzzle(tester, l10n);

      expect(find.text('Wildflower Honey'), findsNothing);
      expect(find.text(l10n.myListingsEmpty), findsOneWidget);
    });

    testWidgets(
      'deleting an item with more results available still fetches the '
      'right offset for the next page',
      (tester) async {
        final (apiClient, adapter) = await _fakeApiClient();
        adapter.listings = [
          _listingJson(id: 1, title: 'Wildflower Honey'),
          for (var i = 2; i <= 20; i++) _listingJson(id: i, title: 'Listing $i'),
        ];
        adapter.totalOverride = 21;

        await tester.pumpWidget(_wrap(apiClient));
        await tester.pump();
        await tester.pump(const Duration(milliseconds: 50));
        await tester.pump(const Duration(milliseconds: 50));

        await tester.tap(find.byIcon(Icons.more_vert).first);
        await tester.pump();
        await tester.pump(const Duration(milliseconds: 300));
        await tester.tap(find.text(l10n.generalDelete));
        await tester.pump();
        await tester.pump(const Duration(milliseconds: 300));

        final prompt = tester
            .widgetList<Text>(find.byType(Text))
            .firstWhere((t) => RegExp(r'^\d+ \+ \d+ = $').hasMatch(t.data ?? ''));
        final match = RegExp(r'^(\d+) \+ (\d+) = $').firstMatch(prompt.data!)!;
        final sum = int.parse(match.group(1)!) + int.parse(match.group(2)!);
        await tester.enterText(
          find.descendant(
            of: find.byType(AlertDialog),
            matching: find.byType(TextField),
          ),
          '$sum',
        );
        await tester.tap(find.text(l10n.generalDelete).last);
        await tester.pump();
        await tester.pump(const Duration(milliseconds: 50));
        await tester.pump(const Duration(milliseconds: 50));

        expect(find.text('Wildflower Honey'), findsNothing);
        expect(find.text('Listing 2'), findsOneWidget);

        await tester.drag(find.byType(ListView), const Offset(0, -100000));
        await tester.pump();
        await tester.pump(const Duration(milliseconds: 50));
        await tester.pump(const Duration(milliseconds: 50));

        final offsets = adapter.requests
            .where((r) => r.path.endsWith('/listings') && r.method == 'GET')
            .map((r) => r.queryParameters['offset'])
            .toList();
        expect(offsets, contains(19));
      },
    );

    testWidgets(
      'typing in the search field reloads with the keyword after a debounce',
      (tester) async {
        final (apiClient, adapter) = await _fakeApiClient();
        adapter.listings = [_listingJson()];

        await tester.pumpWidget(_wrap(apiClient));
        await tester.pumpAndSettle();

        await tester.enterText(find.byType(TextField), 'wildflower');
        await tester.pump(const Duration(milliseconds: 500));
        await tester.pumpAndSettle();

        final searchRequests = adapter.requests.where(
          (r) => r.path.endsWith('/listings') && r.method == 'GET',
        );
        expect(searchRequests.last.queryParameters['keyword'], 'wildflower');
        expect(searchRequests.last.queryParameters['mine'], true);
      },
    );

    testWidgets(
      'selecting a category reloads with the category filter applied',
      (tester) async {
        final (apiClient, adapter) = await _fakeApiClient();
        adapter.listings = [_listingJson()];

        await tester.pumpWidget(_wrap(apiClient));
        await tester.pumpAndSettle();

        await tester.tap(find.byType(DropdownButtonFormField<String?>));
        await tester.pumpAndSettle();
        await tester.tap(find.text(l10n.marketplaceCategoryHoney).last);
        await tester.pumpAndSettle();

        final searchRequests = adapter.requests.where(
          (r) => r.path.endsWith('/listings') && r.method == 'GET',
        );
        expect(searchRequests.last.queryParameters['category'], 'HONEY');
      },
    );

    testWidgets(
      'wrong puzzle answer shows an error and does not delete',
      (tester) async {
        final (apiClient, adapter) = await _fakeApiClient();
        adapter.listings = [_listingJson()];

        await tester.pumpWidget(_wrap(apiClient));
        await tester.pumpAndSettle();

        await tester.tap(find.byIcon(Icons.more_vert));
        await tester.pumpAndSettle();
        await tester.tap(find.text(l10n.generalDelete));
        await tester.pumpAndSettle();

        await tester.enterText(
          find.descendant(
            of: find.byType(AlertDialog),
            matching: find.byType(TextField),
          ),
          '-1',
        );
        await tester.tap(find.text(l10n.generalDelete).last);
        await tester.pumpAndSettle();

        expect(find.text(l10n.deletePuzzleWrong), findsOneWidget);
        expect(find.text('Wildflower Honey'), findsOneWidget);
        expect(adapter.requests.any((r) => r.method == 'DELETE'), isFalse);
      },
    );

    testWidgets(
      'delete failure shows an error snackbar and keeps the card',
      (tester) async {
        final (apiClient, adapter) = await _fakeApiClient();
        adapter.listings = [_listingJson()];
        adapter.failMutations = true;

        await tester.pumpWidget(_wrap(apiClient));
        await tester.pumpAndSettle();

        await tester.tap(find.byIcon(Icons.more_vert));
        await tester.pumpAndSettle();
        await tester.tap(find.text(l10n.generalDelete));
        await tester.pumpAndSettle();
        await _solveDeletePuzzle(tester, l10n);

        expect(find.text('Wildflower Honey'), findsOneWidget);
        expect(find.text(l10n.generalError), findsOneWidget);
      },
    );
  });
}
