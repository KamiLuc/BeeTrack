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
  await tester.enterText(find.byType(TextField), '$sum');
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

    testWidgets('no pagination banner when everything fits on one page', (
      tester,
    ) async {
      final (apiClient, adapter) = await _fakeApiClient();
      adapter.listings = [_listingJson()];

      await tester.pumpWidget(_wrap(apiClient));
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.chevron_left), findsNothing);
      expect(find.byIcon(Icons.chevron_right), findsNothing);
    });

    testWidgets(
      'shows numbered pagination and requests the right offset when a '
      'page is tapped',
      (tester) async {
        final (apiClient, adapter) = await _fakeApiClient();
        adapter.listings = [_listingJson()];
        adapter.totalOverride = 45;

        await tester.pumpWidget(_wrap(apiClient));
        await tester.pumpAndSettle();

        // total: 45, pageSize: 20 -> 3 pages.
        expect(find.text('1'), findsOneWidget);
        expect(find.text('2'), findsOneWidget);
        expect(find.text('3'), findsOneWidget);

        await tester.tap(find.text('2'));
        await tester.pumpAndSettle();

        final lastRequest = adapter.requests
            .where((r) => r.path.endsWith('/listings') && r.method == 'GET')
            .last;
        expect(lastRequest.queryParameters['offset'], 20);
        expect(lastRequest.queryParameters['limit'], 20);
      },
    );

    testWidgets('tapping a card does nothing — edit is only via the menu', (
      tester,
    ) async {
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
      'deleting the last item on a non-first page navigates back a page',
      (tester) async {
        final (apiClient, adapter) = await _fakeApiClient();
        adapter.listings = [_listingJson()];
        adapter.totalOverride = 21;

        await tester.pumpWidget(_wrap(apiClient));
        await tester.pumpAndSettle();

        await tester.tap(find.text('2'));
        await tester.pumpAndSettle();

        await tester.tap(find.byIcon(Icons.more_vert));
        await tester.pumpAndSettle();
        await tester.tap(find.text(l10n.generalDelete));
        await tester.pumpAndSettle();
        await _solveDeletePuzzle(tester, l10n);

        final searchRequests = adapter.requests.where(
          (r) => r.path.endsWith('/listings') && r.method == 'GET',
        );
        expect(searchRequests.last.queryParameters['offset'], 0);
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

        await tester.enterText(find.byType(TextField), '-1');
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
