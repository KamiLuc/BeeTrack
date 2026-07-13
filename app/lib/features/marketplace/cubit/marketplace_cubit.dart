import 'package:flutter_bloc/flutter_bloc.dart';

import '../data/favorites_repository.dart';
import '../data/listing_model.dart';
import '../data/listing_repository.dart';

part 'marketplace_state.dart';

const int _pageSize = 20;

class MarketplaceCubit extends Cubit<MarketplaceState> {
  final ListingRepository _repo;
  final FavoritesRepository _favoritesRepo;

  String? _category;
  String _keyword = '';

  MarketplaceCubit({
    required ListingRepository repo,
    required FavoritesRepository favoritesRepo,
  }) : _repo = repo,
       _favoritesRepo = favoritesRepo,
       super(MarketplaceInitial());

  Future<Set<int>> _loadFavoriteIds() async {
    try {
      final favorites = await _favoritesRepo.listFavorites();
      return favorites.map((l) => l.id).toSet();
    } catch (_) {
      return {};
    }
  }

  Future<void> load() => _goToPage(1);

  Future<void> goToPage(int page) => _goToPage(page);

  Future<void> _goToPage(int page) async {
    emit(MarketplaceLoading());
    try {
      final result = await _repo.searchListings(
        category: _category,
        keyword: _keyword.isEmpty ? null : _keyword,
        limit: _pageSize,
        offset: (page - 1) * _pageSize,
      );
      final favoriteIds = await _loadFavoriteIds();
      final totalPages = (result.total / _pageSize).ceil().clamp(1, 999999);
      emit(
        MarketplaceLoaded(
          items: result.items,
          currentPage: page,
          totalPages: totalPages,
          category: _category,
          keyword: _keyword,
          favoriteIds: favoriteIds,
        ),
      );
    } catch (_) {
      emit(MarketplaceError());
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
