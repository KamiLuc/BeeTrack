part of 'marketplace_cubit.dart';

sealed class MarketplaceState {}

final class MarketplaceInitial extends MarketplaceState {}

final class MarketplaceLoading extends MarketplaceState {}

final class MarketplaceError extends MarketplaceState {}

final class MarketplaceLoaded extends MarketplaceState {
  final List<Listing> items;
  final int total;
  final bool hasMore;
  final bool isLoadingMore;
  final String? category;
  final String keyword;
  final Set<int> favoriteIds;
  final bool hasOwnListings;

  MarketplaceLoaded({
    required this.items,
    required this.total,
    required this.hasMore,
    this.isLoadingMore = false,
    required this.category,
    required this.keyword,
    this.favoriteIds = const {},
    this.hasOwnListings = false,
  });

  MarketplaceLoaded copyWith({
    List<Listing>? items,
    int? total,
    bool? hasMore,
    bool? isLoadingMore,
    Set<int>? favoriteIds,
  }) {
    return MarketplaceLoaded(
      items: items ?? this.items,
      total: total ?? this.total,
      hasMore: hasMore ?? this.hasMore,
      isLoadingMore: isLoadingMore ?? this.isLoadingMore,
      category: category,
      keyword: keyword,
      favoriteIds: favoriteIds ?? this.favoriteIds,
      hasOwnListings: hasOwnListings,
    );
  }
}
