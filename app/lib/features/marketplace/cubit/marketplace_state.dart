part of 'marketplace_cubit.dart';

sealed class MarketplaceState {}

final class MarketplaceInitial extends MarketplaceState {}

final class MarketplaceLoading extends MarketplaceState {}

final class MarketplaceError extends MarketplaceState {}

final class MarketplaceLoaded extends MarketplaceState {
  final List<Listing> items;
  final int currentPage;
  final int totalPages;
  final String? category;
  final String keyword;
  final Set<int> favoriteIds;
  final bool hasOwnListings;

  MarketplaceLoaded({
    required this.items,
    required this.currentPage,
    required this.totalPages,
    required this.category,
    required this.keyword,
    this.favoriteIds = const {},
    this.hasOwnListings = false,
  });

  MarketplaceLoaded copyWith({List<Listing>? items, Set<int>? favoriteIds}) {
    return MarketplaceLoaded(
      items: items ?? this.items,
      currentPage: currentPage,
      totalPages: totalPages,
      category: category,
      keyword: keyword,
      favoriteIds: favoriteIds ?? this.favoriteIds,
      hasOwnListings: hasOwnListings,
    );
  }
}
