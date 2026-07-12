import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:app/features/marketplace/cubit/marketplace_cubit.dart';
import 'package:app/features/marketplace/data/listing_model.dart';
import 'package:app/features/marketplace/data/listing_repository.dart';

class MockListingRepository extends Mock implements ListingRepository {}

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
  late MarketplaceCubit cubit;

  setUp(() {
    repo = MockListingRepository();
    cubit = MarketplaceCubit(repo: repo);
  });

  tearDown(() => cubit.close());

  group('load', () {
    blocTest<MarketplaceCubit, MarketplaceState>(
      'emits [Loading, Loaded] on success',
      build: () {
        when(() => repo.searchListings(
              category: any(named: 'category'),
              keyword: any(named: 'keyword'),
              limit: any(named: 'limit'),
            )).thenAnswer(
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
        when(() => repo.searchListings(
              category: any(named: 'category'),
              keyword: any(named: 'keyword'),
              limit: any(named: 'limit'),
            )).thenThrow(Exception('network error'));
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
        when(() => repo.searchListings(
              category: any(named: 'category'),
              keyword: any(named: 'keyword'),
              limit: any(named: 'limit'),
            )).thenAnswer(
          (_) async => ListingSearchResult(items: [], total: 0),
        );
        return cubit;
      },
      act: (c) => c.setCategory('HONEY'),
      verify: (_) {
        verify(() => repo.searchListings(
              category: 'HONEY',
              keyword: null,
              limit: any(named: 'limit'),
            )).called(1);
      },
    );

    blocTest<MarketplaceCubit, MarketplaceState>(
      'reloads with the given keyword',
      build: () {
        when(() => repo.searchListings(
              category: any(named: 'category'),
              keyword: any(named: 'keyword'),
              limit: any(named: 'limit'),
            )).thenAnswer(
          (_) async => ListingSearchResult(items: [], total: 0),
        );
        return cubit;
      },
      act: (c) => c.setKeyword('honey'),
      verify: (_) {
        verify(() => repo.searchListings(
              category: null,
              keyword: 'honey',
              limit: any(named: 'limit'),
            )).called(1);
      },
    );
  });

  group('loadMore', () {
    blocTest<MarketplaceCubit, MarketplaceState>(
      'appends items and keeps total',
      build: () {
        when(() => repo.searchListings(
              category: any(named: 'category'),
              keyword: any(named: 'keyword'),
              limit: any(named: 'limit'),
            )).thenAnswer(
          (_) async => ListingSearchResult(items: [_listing(1)], total: 2),
        );
        when(() => repo.searchListings(
              category: any(named: 'category'),
              keyword: any(named: 'keyword'),
              limit: any(named: 'limit'),
              offset: any(named: 'offset'),
            )).thenAnswer(
          (_) async => ListingSearchResult(items: [_listing(2)], total: 2),
        );
        return cubit;
      },
      act: (c) async {
        await c.load();
        await c.loadMore();
      },
      expect: () => [
        isA<MarketplaceLoading>(),
        isA<MarketplaceLoaded>().having((s) => s.items.length, 'length', 1),
        isA<MarketplaceLoaded>().having((s) => s.loadingMore, 'loadingMore', true),
        isA<MarketplaceLoaded>()
            .having((s) => s.items.length, 'length', 2)
            .having((s) => s.loadingMore, 'loadingMore', false),
      ],
    );

    blocTest<MarketplaceCubit, MarketplaceState>(
      'does nothing when all items are already loaded',
      build: () {
        when(() => repo.searchListings(
              category: any(named: 'category'),
              keyword: any(named: 'keyword'),
              limit: any(named: 'limit'),
            )).thenAnswer(
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
        isA<MarketplaceLoaded>().having((s) => s.items.length, 'length', 1),
      ],
    );
  });
}
