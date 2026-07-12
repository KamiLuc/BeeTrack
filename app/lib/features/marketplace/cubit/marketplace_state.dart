part of 'marketplace_cubit.dart';

sealed class MarketplaceState {}

final class MarketplaceInitial extends MarketplaceState {}

final class MarketplaceLoading extends MarketplaceState {}

final class MarketplaceError extends MarketplaceState {}

final class MarketplaceLoaded extends MarketplaceState {
  final List<Listing> items;
  final int total;
  final String? category;
  final String keyword;
  final bool loadingMore;

  MarketplaceLoaded({
    required this.items,
    required this.total,
    required this.category,
    required this.keyword,
    this.loadingMore = false,
  });

  MarketplaceLoaded copyWith({
    List<Listing>? items,
    int? total,
    bool? loadingMore,
  }) {
    return MarketplaceLoaded(
      items: items ?? this.items,
      total: total ?? this.total,
      category: category,
      keyword: keyword,
      loadingMore: loadingMore ?? this.loadingMore,
    );
  }
}
