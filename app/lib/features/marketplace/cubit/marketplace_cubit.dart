import 'package:flutter_bloc/flutter_bloc.dart';

import '../data/listing_model.dart';
import '../data/listing_repository.dart';

part 'marketplace_state.dart';

const int _pageSize = 20;

class MarketplaceCubit extends Cubit<MarketplaceState> {
  final ListingRepository _repo;

  String? _category;
  String _keyword = '';

  MarketplaceCubit({required this._repo}) : super(MarketplaceInitial());

  Future<void> load() async {
    emit(MarketplaceLoading());
    try {
      final result = await _repo.searchListings(
        category: _category,
        keyword: _keyword.isEmpty ? null : _keyword,
        limit: _pageSize,
      );
      emit(MarketplaceLoaded(
        items: result.items,
        total: result.total,
        category: _category,
        keyword: _keyword,
      ));
    } catch (_) {
      emit(MarketplaceError());
    }
  }

  Future<void> loadMore() async {
    final current = state;
    if (current is! MarketplaceLoaded || current.items.length >= current.total) {
      return;
    }
    emit(current.copyWith(loadingMore: true));
    try {
      final result = await _repo.searchListings(
        category: _category,
        keyword: _keyword.isEmpty ? null : _keyword,
        limit: _pageSize,
        offset: current.items.length,
      );
      emit(current.copyWith(
        items: [...current.items, ...result.items],
        total: result.total,
        loadingMore: false,
      ));
    } catch (_) {
      emit(current.copyWith(loadingMore: false));
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
}
