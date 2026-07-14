import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:app/core/storage/token_storage.dart';
import 'package:app/features/marketplace/cubit/marketplace_cubit.dart';
import 'package:app/features/marketplace/data/favorites_repository.dart';
import 'package:app/features/marketplace/data/listing_model.dart';
import 'package:app/features/marketplace/data/listing_repository.dart';

class MockListingRepository extends Mock implements ListingRepository {}

class MockFavoritesRepository extends Mock implements FavoritesRepository {}

/// Builds a [TokenStorage] backed by mock [SharedPreferences], optionally
/// pre-populated with an access token so `accessToken` returns non-null.
Future<TokenStorage> _tokenStorageWithToken({bool loggedIn = true}) async {
  SharedPreferences.setMockInitialValues(
    loggedIn ? {'access_token': 'fake.jwt.token'} : {},
  );
  final prefs = await SharedPreferences.getInstance();
  return TokenStorage(prefs);
}

Listing _listing(int id) => Listing(
  id: id,
  userId: 1,
  title: 'Listing $id',
  description: '',
  category: 'HONEY',
  quantity: '',
  address: '',
  contactPhone: '',
  contactEmail: '',
  isHidden: false,
  createdAt: DateTime.utc(2025, 1, 1),
  updatedAt: DateTime.utc(2025, 1, 1),
  images: const [],
);

void main() {
  late MockListingRepository repo;
  late MockFavoritesRepository favoritesRepo;
  late TokenStorage tokenStorage;
  late MarketplaceCubit cubit;

  setUp(() async {
    repo = MockListingRepository();
    favoritesRepo = MockFavoritesRepository();
    tokenStorage = await _tokenStorageWithToken();
    when(() => favoritesRepo.listFavorites()).thenAnswer((_) async => []);
    when(
      () => repo.searchListings(mine: true, limit: any(named: 'limit')),
    ).thenAnswer((_) async => ListingSearchResult(items: [], total: 0));
    cubit = MarketplaceCubit(
      repo: repo,
      favoritesRepo: favoritesRepo,
      tokenStorage: tokenStorage,
    );
  });

  tearDown(() => cubit.close());

  group('load', () {
    blocTest<MarketplaceCubit, MarketplaceState>(
      'emits [Loading, Loaded] on success',
      build: () {
        when(
          () => repo.searchListings(
            category: any(named: 'category'),
            keyword: any(named: 'keyword'),
            limit: any(named: 'limit'),
          ),
        ).thenAnswer(
          (_) async => ListingSearchResult(items: [_listing(1)], total: 1),
        );
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [
        isA<MarketplaceLoading>(),
        isA<MarketplaceLoaded>().having((s) => s.items.length, 'length', 1),
      ],
    );

    blocTest<MarketplaceCubit, MarketplaceState>(
      'emits [Loading, Error] on failure',
      build: () {
        when(
          () => repo.searchListings(
            category: any(named: 'category'),
            keyword: any(named: 'keyword'),
            limit: any(named: 'limit'),
          ),
        ).thenThrow(Exception('network error'));
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [isA<MarketplaceLoading>(), isA<MarketplaceError>()],
    );
  });

  group('setCategory / setKeyword', () {
    blocTest<MarketplaceCubit, MarketplaceState>(
      'reloads with the chosen category filter',
      build: () {
        when(
          () => repo.searchListings(
            category: any(named: 'category'),
            keyword: any(named: 'keyword'),
            limit: any(named: 'limit'),
          ),
        ).thenAnswer((_) async => ListingSearchResult(items: [], total: 0));
        return cubit;
      },
      act: (c) => c.setCategory('HONEY'),
      verify: (_) {
        verify(
          () => repo.searchListings(
            category: 'HONEY',
            keyword: null,
            limit: any(named: 'limit'),
          ),
        ).called(1);
      },
    );

    blocTest<MarketplaceCubit, MarketplaceState>(
      'reloads with the given keyword',
      build: () {
        when(
          () => repo.searchListings(
            category: any(named: 'category'),
            keyword: any(named: 'keyword'),
            limit: any(named: 'limit'),
          ),
        ).thenAnswer((_) async => ListingSearchResult(items: [], total: 0));
        return cubit;
      },
      act: (c) => c.setKeyword('honey'),
      verify: (_) {
        verify(
          () => repo.searchListings(
            category: null,
            keyword: 'honey',
            limit: any(named: 'limit'),
          ),
        ).called(1);
      },
    );
  });

  group('load favorites', () {
    blocTest<MarketplaceCubit, MarketplaceState>(
      'populates favoriteIds from the favorites repository',
      build: () {
        when(
          () => repo.searchListings(
            category: any(named: 'category'),
            keyword: any(named: 'keyword'),
            limit: any(named: 'limit'),
          ),
        ).thenAnswer(
          (_) async =>
              ListingSearchResult(items: [_listing(1), _listing(2)], total: 2),
        );
        when(
          () => favoritesRepo.listFavorites(),
        ).thenAnswer((_) async => [_listing(2)]);
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [
        isA<MarketplaceLoading>(),
        isA<MarketplaceLoaded>().having((s) => s.favoriteIds, 'favoriteIds', {
          2,
        }),
      ],
    );

    blocTest<MarketplaceCubit, MarketplaceState>(
      'treats a favorites-fetch failure as no favorites, not an error',
      build: () {
        when(
          () => repo.searchListings(
            category: any(named: 'category'),
            keyword: any(named: 'keyword'),
            limit: any(named: 'limit'),
          ),
        ).thenAnswer(
          (_) async => ListingSearchResult(items: [_listing(1)], total: 1),
        );
        when(
          () => favoritesRepo.listFavorites(),
        ).thenThrow(Exception('unauthenticated'));
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [
        isA<MarketplaceLoading>(),
        isA<MarketplaceLoaded>().having(
          (s) => s.favoriteIds,
          'favoriteIds',
          <int>{},
        ),
      ],
    );
  });

  group('toggleFavorite', () {
    blocTest<MarketplaceCubit, MarketplaceState>(
      'optimistically adds and calls the repository',
      build: () {
        when(
          () => repo.searchListings(
            category: any(named: 'category'),
            keyword: any(named: 'keyword'),
            limit: any(named: 'limit'),
          ),
        ).thenAnswer(
          (_) async => ListingSearchResult(items: [_listing(1)], total: 1),
        );
        when(() => favoritesRepo.listFavorites()).thenAnswer((_) async => []);
        when(() => favoritesRepo.addFavorite(1)).thenAnswer((_) async {});
        return cubit;
      },
      act: (c) async {
        await c.load();
        await c.toggleFavorite(1);
      },
      expect: () => [
        isA<MarketplaceLoading>(),
        isA<MarketplaceLoaded>().having(
          (s) => s.favoriteIds,
          'favoriteIds',
          <int>{},
        ),
        isA<MarketplaceLoaded>().having((s) => s.favoriteIds, 'favoriteIds', {
          1,
        }),
      ],
      verify: (_) {
        verify(() => favoritesRepo.addFavorite(1)).called(1);
      },
    );

    blocTest<MarketplaceCubit, MarketplaceState>(
      'reverts the optimistic update when the repository call fails',
      build: () {
        when(
          () => repo.searchListings(
            category: any(named: 'category'),
            keyword: any(named: 'keyword'),
            limit: any(named: 'limit'),
          ),
        ).thenAnswer(
          (_) async => ListingSearchResult(items: [_listing(1)], total: 1),
        );
        when(() => favoritesRepo.listFavorites()).thenAnswer((_) async => []);
        when(
          () => favoritesRepo.addFavorite(1),
        ).thenThrow(Exception('network error'));
        return cubit;
      },
      act: (c) async {
        await c.load();
        await c.toggleFavorite(1);
      },
      expect: () => [
        isA<MarketplaceLoading>(),
        isA<MarketplaceLoaded>().having(
          (s) => s.favoriteIds,
          'favoriteIds',
          <int>{},
        ),
        isA<MarketplaceLoaded>().having((s) => s.favoriteIds, 'favoriteIds', {
          1,
        }),
        isA<MarketplaceLoaded>().having(
          (s) => s.favoriteIds,
          'favoriteIds',
          <int>{},
        ),
      ],
    );
  });

  group('load / loadMore pagination', () {
    blocTest<MarketplaceCubit, MarketplaceState>(
      'load() fetches offset 0 and computes hasMore from total',
      build: () {
        when(
          () => repo.searchListings(
            category: any(named: 'category'),
            keyword: any(named: 'keyword'),
            limit: any(named: 'limit'),
            offset: any(named: 'offset'),
          ),
        ).thenAnswer(
          (_) async => ListingSearchResult(items: [_listing(1)], total: 45),
        );
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [
        isA<MarketplaceLoading>(),
        isA<MarketplaceLoaded>()
            .having((s) => s.items.length, 'items.length', 1)
            .having((s) => s.total, 'total', 45)
            .having((s) => s.hasMore, 'hasMore', true),
      ],
      verify: (_) {
        verify(
          () => repo.searchListings(
            category: any(named: 'category'),
            keyword: any(named: 'keyword'),
            limit: 20,
            offset: 0,
          ),
        ).called(1);
      },
    );

    blocTest<MarketplaceCubit, MarketplaceState>(
      'loadMore() requests the next offset and appends results without '
      'losing or duplicating existing items',
      build: () {
        when(
          () => repo.searchListings(
            category: any(named: 'category'),
            keyword: any(named: 'keyword'),
            limit: any(named: 'limit'),
            offset: 0,
          ),
        ).thenAnswer(
          (_) async => ListingSearchResult(
            items: [_listing(1), _listing(2)],
            total: 3,
          ),
        );
        when(
          () => repo.searchListings(
            category: any(named: 'category'),
            keyword: any(named: 'keyword'),
            limit: any(named: 'limit'),
            offset: 2,
          ),
        ).thenAnswer(
          (_) async => ListingSearchResult(items: [_listing(3)], total: 3),
        );
        return cubit;
      },
      act: (c) async {
        await c.load();
        await c.loadMore();
      },
      expect: () => [
        isA<MarketplaceLoading>(),
        isA<MarketplaceLoaded>()
            .having((s) => s.items.map((i) => i.id), 'ids', [1, 2])
            .having((s) => s.hasMore, 'hasMore', true),
        isA<MarketplaceLoaded>().having(
          (s) => s.isLoadingMore,
          'isLoadingMore',
          true,
        ),
        isA<MarketplaceLoaded>()
            .having((s) => s.items.map((i) => i.id), 'ids', [1, 2, 3])
            .having((s) => s.total, 'total', 3)
            .having((s) => s.hasMore, 'hasMore', false)
            .having((s) => s.isLoadingMore, 'isLoadingMore', false),
      ],
      verify: (_) {
        verify(
          () => repo.searchListings(
            category: any(named: 'category'),
            keyword: any(named: 'keyword'),
            limit: 20,
            offset: 2,
          ),
        ).called(1);
      },
    );

    blocTest<MarketplaceCubit, MarketplaceState>(
      'loadMore() is a no-op when hasMore is false',
      build: () {
        when(
          () => repo.searchListings(
            category: any(named: 'category'),
            keyword: any(named: 'keyword'),
            limit: any(named: 'limit'),
            offset: any(named: 'offset'),
          ),
        ).thenAnswer(
          (_) async => ListingSearchResult(items: [_listing(1)], total: 1),
        );
        return cubit;
      },
      act: (c) async {
        await c.load();
        await c.loadMore();
      },
      expect: () => [
        isA<MarketplaceLoading>(),
        isA<MarketplaceLoaded>().having((s) => s.hasMore, 'hasMore', false),
      ],
      verify: (_) {
        verify(
          () => repo.searchListings(
            category: any(named: 'category'),
            keyword: any(named: 'keyword'),
            limit: any(named: 'limit'),
            offset: any(named: 'offset'),
          ),
        ).called(1);
      },
    );

    blocTest<MarketplaceCubit, MarketplaceState>(
      'loadMore() is a no-op when already isLoadingMore',
      build: () {
        when(
          () => repo.searchListings(
            category: any(named: 'category'),
            keyword: any(named: 'keyword'),
            limit: any(named: 'limit'),
            offset: 0,
          ),
        ).thenAnswer(
          (_) async => ListingSearchResult(items: [_listing(1)], total: 2),
        );
        when(
          () => repo.searchListings(
            category: any(named: 'category'),
            keyword: any(named: 'keyword'),
            limit: any(named: 'limit'),
            offset: 1,
          ),
        ).thenAnswer(
          (_) async => ListingSearchResult(items: [_listing(2)], total: 2),
        );
        return cubit;
      },
      act: (c) async {
        await c.load();
        final first = c.loadMore();
        final second = c.loadMore();
        await Future.wait([first, second]);
      },
      verify: (_) {
        verify(
          () => repo.searchListings(
            category: any(named: 'category'),
            keyword: any(named: 'keyword'),
            limit: any(named: 'limit'),
            offset: 1,
          ),
        ).called(1);
      },
    );

    blocTest<MarketplaceCubit, MarketplaceState>(
      'loadMore() failure keeps existing items and does not emit '
      'MarketplaceError',
      build: () {
        when(
          () => repo.searchListings(
            category: any(named: 'category'),
            keyword: any(named: 'keyword'),
            limit: any(named: 'limit'),
            offset: 0,
          ),
        ).thenAnswer(
          (_) async => ListingSearchResult(items: [_listing(1)], total: 2),
        );
        when(
          () => repo.searchListings(
            category: any(named: 'category'),
            keyword: any(named: 'keyword'),
            limit: any(named: 'limit'),
            offset: 1,
          ),
        ).thenThrow(Exception('network error'));
        return cubit;
      },
      act: (c) async {
        await c.load();
        await c.loadMore();
      },
      expect: () => [
        isA<MarketplaceLoading>(),
        isA<MarketplaceLoaded>()
            .having((s) => s.items.length, 'items.length', 1)
            .having((s) => s.hasMore, 'hasMore', true),
        isA<MarketplaceLoaded>().having(
          (s) => s.isLoadingMore,
          'isLoadingMore',
          true,
        ),
        isA<MarketplaceLoaded>()
            .having((s) => s.items.length, 'items.length', 1)
            .having((s) => s.isLoadingMore, 'isLoadingMore', false),
      ],
    );
  });

  group('load hasOwnListings', () {
    blocTest<MarketplaceCubit, MarketplaceState>(
      'is true when the mine=true search returns a positive total',
      build: () {
        when(
          () => repo.searchListings(
            category: any(named: 'category'),
            keyword: any(named: 'keyword'),
            limit: any(named: 'limit'),
          ),
        ).thenAnswer(
          (_) async => ListingSearchResult(items: [_listing(1)], total: 1),
        );
        when(
          () => repo.searchListings(mine: true, limit: any(named: 'limit')),
        ).thenAnswer(
          (_) async => ListingSearchResult(items: [_listing(1)], total: 3),
        );
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [
        isA<MarketplaceLoading>(),
        isA<MarketplaceLoaded>().having(
          (s) => s.hasOwnListings,
          'hasOwnListings',
          true,
        ),
      ],
    );

    blocTest<MarketplaceCubit, MarketplaceState>(
      'is false when the mine=true search returns a zero total',
      build: () {
        when(
          () => repo.searchListings(
            category: any(named: 'category'),
            keyword: any(named: 'keyword'),
            limit: any(named: 'limit'),
          ),
        ).thenAnswer(
          (_) async => ListingSearchResult(items: [_listing(1)], total: 1),
        );
        when(
          () => repo.searchListings(mine: true, limit: any(named: 'limit')),
        ).thenAnswer((_) async => ListingSearchResult(items: [], total: 0));
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [
        isA<MarketplaceLoading>(),
        isA<MarketplaceLoaded>().having(
          (s) => s.hasOwnListings,
          'hasOwnListings',
          false,
        ),
      ],
    );

    blocTest<MarketplaceCubit, MarketplaceState>(
      'is false when the mine=true search throws',
      build: () {
        when(
          () => repo.searchListings(
            category: any(named: 'category'),
            keyword: any(named: 'keyword'),
            limit: any(named: 'limit'),
          ),
        ).thenAnswer(
          (_) async => ListingSearchResult(items: [_listing(1)], total: 1),
        );
        when(
          () => repo.searchListings(mine: true, limit: any(named: 'limit')),
        ).thenThrow(Exception('network error'));
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [
        isA<MarketplaceLoading>(),
        isA<MarketplaceLoaded>().having(
          (s) => s.hasOwnListings,
          'hasOwnListings',
          false,
        ),
      ],
    );

    late TokenStorage loggedOutStorage;

    blocTest<MarketplaceCubit, MarketplaceState>(
      'skips the mine=true search entirely when logged out',
      setUp: () async {
        loggedOutStorage = await _tokenStorageWithToken(loggedIn: false);
      },
      build: () {
        when(
          () => repo.searchListings(
            category: any(named: 'category'),
            keyword: any(named: 'keyword'),
            limit: any(named: 'limit'),
          ),
        ).thenAnswer(
          (_) async => ListingSearchResult(items: [_listing(1)], total: 1),
        );
        return MarketplaceCubit(
          repo: repo,
          favoritesRepo: favoritesRepo,
          tokenStorage: loggedOutStorage,
        );
      },
      act: (c) => c.load(),
      expect: () => [
        isA<MarketplaceLoading>(),
        isA<MarketplaceLoaded>().having(
          (s) => s.hasOwnListings,
          'hasOwnListings',
          false,
        ),
      ],
      verify: (_) {
        verifyNever(
          () => repo.searchListings(mine: true, limit: any(named: 'limit')),
        );
      },
    );
  });
}
