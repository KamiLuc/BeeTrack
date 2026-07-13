import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:app/features/marketplace/cubit/marketplace_cubit.dart';
import 'package:app/features/marketplace/data/favorites_repository.dart';
import 'package:app/features/marketplace/data/listing_model.dart';
import 'package:app/features/marketplace/data/listing_repository.dart';

class MockListingRepository extends Mock implements ListingRepository {}

class MockFavoritesRepository extends Mock implements FavoritesRepository {}

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
  late MarketplaceCubit cubit;

  setUp(() {
    repo = MockListingRepository();
    favoritesRepo = MockFavoritesRepository();
    when(() => favoritesRepo.listFavorites()).thenAnswer((_) async => []);
    cubit = MarketplaceCubit(repo: repo, favoritesRepo: favoritesRepo);
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

  group('goToPage', () {
    blocTest<MarketplaceCubit, MarketplaceState>(
      'load() fetches page 1 with offset 0 and computes totalPages',
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
            .having((s) => s.currentPage, 'currentPage', 1)
            .having((s) => s.totalPages, 'totalPages', 3),
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
      'goToPage(2) requests offset 20',
      build: () {
        when(
          () => repo.searchListings(
            category: any(named: 'category'),
            keyword: any(named: 'keyword'),
            limit: any(named: 'limit'),
            offset: any(named: 'offset'),
          ),
        ).thenAnswer(
          (_) async => ListingSearchResult(items: [_listing(21)], total: 45),
        );
        return cubit;
      },
      act: (c) => c.goToPage(2),
      expect: () => [
        isA<MarketplaceLoading>(),
        isA<MarketplaceLoaded>().having((s) => s.currentPage, 'currentPage', 2),
      ],
      verify: (_) {
        verify(
          () => repo.searchListings(
            category: any(named: 'category'),
            keyword: any(named: 'keyword'),
            limit: 20,
            offset: 20,
          ),
        ).called(1);
      },
    );
  });
}
