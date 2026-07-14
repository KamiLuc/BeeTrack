import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/storage/token_storage.dart';
import '../data/favorites_repository.dart';
import '../data/listing_model.dart';
import '../data/listing_repository.dart';

part 'marketplace_state.dart';

const int _pageSize = 20;

class MarketplaceCubit extends Cubit<MarketplaceState> {
  final ListingRepository _repo;
  final FavoritesRepository _favoritesRepo;
  final TokenStorage _tokenStorage;

  String? _category;
  String _keyword = '';
  double? _priceMin;
  double? _priceMax;

  MarketplaceCubit({
    required ListingRepository repo,
    required FavoritesRepository favoritesRepo,
    required TokenStorage tokenStorage,
  }) : _repo = repo,
       _favoritesRepo = favoritesRepo,
       _tokenStorage = tokenStorage,
       super(MarketplaceInitial());

  Future<Set<int>> _loadFavoriteIds() async {
    try {
      final favorites = await _favoritesRepo.listFavorites();
      return favorites.map((l) => l.id).toSet();
    } catch (_) {
      return {};
    }
  }

  Future<bool> _loadHasOwnListings() async {
    if (_tokenStorage.accessToken == null) return false;
    try {
      final result = await _repo.searchListings(mine: true, limit: 1);
      return result.total > 0;
    } catch (_) {
      return false;
    }
  }

  Future<void> load() async {
    emit(MarketplaceLoading());
    try {
      final result = await _repo.searchListings(
        category: _category,
        keyword: _keyword.isEmpty ? null : _keyword,
        priceMin: _priceMin,
        priceMax: _priceMax,
        limit: _pageSize,
        offset: 0,
      );
      final favoriteIds = await _loadFavoriteIds();
      final hasOwnListings = await _loadHasOwnListings();
      emit(
        MarketplaceLoaded(
          items: result.items,
          total: result.total,
          hasMore: result.items.length < result.total,
          category: _category,
          keyword: _keyword,
          priceMin: _priceMin,
          priceMax: _priceMax,
          favoriteIds: favoriteIds,
          hasOwnListings: hasOwnListings,
        ),
      );
    } catch (_) {
      emit(MarketplaceError());
    }
  }

  /// Fetches the next page of results and appends it to the current list.
  /// No-ops if already loading more or if there's nothing left to fetch.
  Future<void> loadMore() async {
    final current = state;
    if (current is! MarketplaceLoaded) return;
    if (current.isLoadingMore || !current.hasMore) return;

    emit(current.copyWith(isLoadingMore: true));
    try {
      final result = await _repo.searchListings(
        category: _category,
        keyword: _keyword.isEmpty ? null : _keyword,
        priceMin: _priceMin,
        priceMax: _priceMax,
        limit: _pageSize,
        offset: current.items.length,
      );
      final items = [...current.items, ...result.items];
      emit(
        current.copyWith(
          items: items,
          total: result.total,
          hasMore: items.length < result.total,
          isLoadingMore: false,
        ),
      );
    } catch (_) {
      emit(current.copyWith(isLoadingMore: false));
    }
  }

  void setCategory(String? category) {
    _category = category;
    load();
  }

  void setKeyword(String keyword) {
    _keyword = keyword;
    load();
  }

  void setPriceRange(double? min, double? max) {
    _priceMin = min;
    _priceMax = max;
    load();
  }

  Future<void> toggleFavorite(int listingId) async {
    final current = state;
    if (current is! MarketplaceLoaded) return;
    final wasFavorite = current.favoriteIds.contains(listingId);
    final optimistic = Set<int>.of(current.favoriteIds);
    wasFavorite ? optimistic.remove(listingId) : optimistic.add(listingId);
    emit(current.copyWith(favoriteIds: optimistic));
    try {
      if (wasFavorite) {
        await _favoritesRepo.removeFavorite(listingId);
      } else {
        await _favoritesRepo.addFavorite(listingId);
      }
    } catch (_) {
      final latest = state;
      if (latest is MarketplaceLoaded) {
        emit(latest.copyWith(favoriteIds: current.favoriteIds));
      }
    }
  }
}
