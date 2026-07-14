import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';

import '../../../core/api/api_client.dart';
import '../../../core/storage/token_storage.dart';
import '../../../core/theme/app_layout.dart';
import '../../../core/widgets/app_drawer.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../features/auth/bloc/auth_bloc.dart';
import '../../../l10n/app_localizations.dart';
import '../cubit/marketplace_cubit.dart';
import '../data/favorites_repository.dart';
import '../data/listing_category.dart';
import '../data/listing_model.dart';
import '../data/listing_price.dart';
import '../data/listing_repository.dart';
import 'create_listing_screen.dart';
import 'favorites_screen.dart';
import 'listing_detail_screen.dart';
import 'my_listings_screen.dart';

class MarketplaceHomeScreen extends StatelessWidget {
  /// Called when an authenticated user picks a section from the drawer.
  final ValueChanged<AppSection>? onSelectSection;

  /// Called when an unauthenticated user taps "Log in" in the drawer.
  final VoidCallback? onLogin;

  const MarketplaceHomeScreen({super.key, this.onSelectSection, this.onLogin});

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => MarketplaceCubit(
        repo: ListingRepository(api: context.read<ApiClient>()),
        favoritesRepo: FavoritesRepository(api: context.read<ApiClient>()),
        tokenStorage: context.read<TokenStorage>(),
      )..load(),
      child: _MarketplaceView(
        onSelectSection: onSelectSection,
        onLogin: onLogin,
      ),
    );
  }
}

class _MarketplaceView extends StatefulWidget {
  final ValueChanged<AppSection>? onSelectSection;
  final VoidCallback? onLogin;

  const _MarketplaceView({this.onSelectSection, this.onLogin});

  @override
  State<_MarketplaceView> createState() => _MarketplaceViewState();
}

class _MarketplaceViewState extends State<_MarketplaceView> {
  final _searchController = TextEditingController();
  Timer? _debounce;

  @override
  void dispose() {
    _debounce?.cancel();
    _searchController.dispose();
    super.dispose();
  }

  void _onSearchChanged(String value) {
    _debounce?.cancel();
    _debounce = Timer(const Duration(milliseconds: 400), () {
      context.read<MarketplaceCubit>().setKeyword(value.trim());
    });
  }

  Future<void> _openCreateListing(BuildContext context) async {
    await Navigator.of(
      context,
    ).push(MaterialPageRoute(builder: (_) => const CreateListingScreen()));
    if (context.mounted) context.read<MarketplaceCubit>().load();
  }

  Future<void> _openMyListings(BuildContext context) async {
    await Navigator.of(
      context,
    ).push(MaterialPageRoute(builder: (_) => const MyListingsScreen()));
    if (context.mounted) context.read<MarketplaceCubit>().load();
  }

  Future<void> _openFavorites(BuildContext context) async {
    await Navigator.of(
      context,
    ).push(MaterialPageRoute(builder: (_) => const FavoritesScreen()));
    if (context.mounted) context.read<MarketplaceCubit>().load();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return BlocBuilder<AuthBloc, AuthState>(
      builder: (context, authState) {
        final isAuthenticated = authState is AuthAuthenticated;
        final drawer = isAuthenticated
            ? AuthenticatedAppDrawer(
                current: AppSection.marketplace,
                onSelect: widget.onSelectSection ?? (_) {},
              )
            : UnauthenticatedAppDrawer(
                onMarketplace: () {},
                onLogin: widget.onLogin ?? () {},
              );

        return Scaffold(
          appBar: AppBar(
            title: Text(l10n.marketplaceTitle),
            actions: [if (isAuthenticated) const ProfileIconButton()],
          ),
          drawer: drawer,
          body: Column(
            children: [
              Expanded(
                child: Center(
                  child: ConstrainedBox(
                    constraints: const BoxConstraints(maxWidth: 600),
                    child: Column(
                      children: [
                        Padding(
                          padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
                          child: TextField(
                            controller: _searchController,
                            decoration: InputDecoration(
                              hintText: l10n.marketplaceSearchHint,
                              prefixIcon: const Icon(Icons.search),
                              isDense: true,
                              border: OutlineInputBorder(
                                borderRadius: BorderRadius.circular(12),
                              ),
                            ),
                            textInputAction: TextInputAction.search,
                            onChanged: _onSearchChanged,
                            onSubmitted: (value) {
                              _debounce?.cancel();
                              context.read<MarketplaceCubit>().setKeyword(
                                value.trim(),
                              );
                            },
                          ),
                        ),
                        const _CategoryDropdown(),
                        const SizedBox(height: 4),
                        Expanded(child: _ListingsFeed()),
                      ],
                    ),
                  ),
                ),
              ),
              BlocBuilder<MarketplaceCubit, MarketplaceState>(
                builder: (context, state) {
                  final loaded = state is MarketplaceLoaded ? state : null;
                  return _MarketplaceBanner(
                    l10n: l10n,
                    isAuthenticated: isAuthenticated,
                    hasOwnListings: loaded?.hasOwnListings ?? false,
                    hasFavorites: loaded?.favoriteIds.isNotEmpty ?? false,
                    onAdd: () => _openCreateListing(context),
                    onMyListings: () => _openMyListings(context),
                    onFavorites: () => _openFavorites(context),
                    onMap: null,
                  );
                },
              ),
            ],
          ),
        );
      },
    );
  }
}

class _MarketplaceBanner extends StatelessWidget {
  final AppLocalizations l10n;
  final bool isAuthenticated;
  final bool hasOwnListings;
  final bool hasFavorites;
  final VoidCallback onAdd;
  final VoidCallback onMyListings;
  final VoidCallback onFavorites;
  final VoidCallback? onMap;

  const _MarketplaceBanner({
    required this.l10n,
    required this.isAuthenticated,
    required this.hasOwnListings,
    required this.hasFavorites,
    required this.onAdd,
    required this.onMyListings,
    required this.onFavorites,
    required this.onMap,
  });

  @override
  Widget build(BuildContext context) {
    final bannerWidth = AppLayout.bannerWidth(context);

    return SafeArea(
      top: false,
      child: Padding(
        padding: const EdgeInsets.only(bottom: 8),
        child: Center(
          child: SizedBox(
            width: bannerWidth,
            child: Container(
              decoration: const BoxDecoration(
                color: Colors.amber,
                borderRadius: BorderRadius.only(
                  topLeft: Radius.circular(20),
                  topRight: Radius.circular(20),
                ),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black26,
                    blurRadius: 8,
                    offset: Offset(0, -2),
                  ),
                ],
              ),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                    children: [
                      if (isAuthenticated) ...[
                        IconButton(
                          icon: const Icon(Icons.add),
                          iconSize: 28,
                          tooltip: l10n.marketplaceCreateScreenTitle,
                          onPressed: onAdd,
                        ),
                        if (hasOwnListings)
                          IconButton(
                            icon: const Icon(Icons.list_alt_outlined),
                            iconSize: 28,
                            tooltip: l10n.myListingsTitle,
                            onPressed: onMyListings,
                          ),
                        if (hasFavorites)
                          IconButton(
                            icon: const Icon(Icons.bookmark_border),
                            iconSize: 28,
                            tooltip: l10n.favoritesTitle,
                            onPressed: onFavorites,
                          ),
                      ],
                      IconButton(
                        icon: const Icon(Icons.map_outlined),
                        iconSize: 28,
                        tooltip: l10n.marketplaceMapTooltip,
                        onPressed: onMap,
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _CategoryDropdown extends StatefulWidget {
  const _CategoryDropdown();

  @override
  State<_CategoryDropdown> createState() => _CategoryDropdownState();
}

class _CategoryDropdownState extends State<_CategoryDropdown> {
  String? _lastSelectedCategory;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final state = context.watch<MarketplaceCubit>().state;

    String? selectedCategory;
    if (state is MarketplaceLoaded) {
      selectedCategory = state.category;
      _lastSelectedCategory = selectedCategory;
    } else {
      selectedCategory = _lastSelectedCategory;
    }

    Widget categoryItem(String? category) {
      return Padding(
        padding: const EdgeInsets.symmetric(horizontal: 12),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(listingCategoryIcon(category), size: 20),
            const SizedBox(width: 8),
            Text(
              category == null
                  ? l10n.marketplaceCategoryAll
                  : listingCategoryLabel(l10n, category),
            ),
          ],
        ),
      );
    }

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Container(
        decoration: BoxDecoration(
          border: Border.all(color: Theme.of(context).colorScheme.outline),
          borderRadius: BorderRadius.circular(12),
        ),
        child: DropdownButton<String?>(
          isExpanded: true,
          value: selectedCategory,
          underline: const SizedBox(),
          onChanged: (category) {
            FocusScope.of(context).unfocus();
            context.read<MarketplaceCubit>().setCategory(category);
          },
          selectedItemBuilder: (BuildContext context) {
            return [
              categoryItem(null),
              for (final category in listingCategories) categoryItem(category),
            ];
          },
          items: [
            DropdownMenuItem(value: null, child: categoryItem(null)),
            for (final category in listingCategories)
              DropdownMenuItem(value: category, child: categoryItem(category)),
          ],
        ),
      ),
    );
  }
}

class _ListingsFeed extends StatefulWidget {
  @override
  State<_ListingsFeed> createState() => _ListingsFeedState();
}

class _ListingsFeedState extends State<_ListingsFeed> {
  final _scrollController = ScrollController();

  @override
  void initState() {
    super.initState();
    _scrollController.addListener(_onScroll);
  }

  @override
  void dispose() {
    _scrollController.removeListener(_onScroll);
    _scrollController.dispose();
    super.dispose();
  }

  void _onScroll() {
    const threshold = 300.0;
    if (_scrollController.position.pixels >=
        _scrollController.position.maxScrollExtent - threshold) {
      context.read<MarketplaceCubit>().loadMore();
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final state = context.watch<MarketplaceCubit>().state;

    if (state is MarketplaceLoading || state is MarketplaceInitial) {
      return const Center(child: CircularProgressIndicator());
    }
    if (state is MarketplaceError) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              l10n.generalError,
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 12),
            ElevatedButton(
              onPressed: () => context.read<MarketplaceCubit>().load(),
              child: Text(l10n.generalRetry),
            ),
          ],
        ),
      );
    }
    if (state is MarketplaceLoaded) {
      if (state.items.isEmpty) {
        return Center(child: Text(l10n.marketplaceEmpty));
      }
      return RefreshIndicator(
        onRefresh: () => context.read<MarketplaceCubit>().load(),
        child: Center(
          child: ConstrainedBox(
            constraints: AppLayout.formConstraints(context),
            child: ListView.builder(
              controller: _scrollController,
              physics: const AlwaysScrollableScrollPhysics(),
              padding: const EdgeInsets.fromLTRB(16, 4, 16, 16),
              itemCount: state.items.length + (state.hasMore ? 1 : 0),
              itemBuilder: (context, index) {
                if (index >= state.items.length) {
                  return const Padding(
                    padding: EdgeInsets.symmetric(vertical: 16),
                    child: Center(child: CircularProgressIndicator()),
                  );
                }
                final listing = state.items[index];
                return Padding(
                  padding: const EdgeInsets.only(bottom: 8),
                  child: _ListingCard(
                    listing: listing,
                    isFavorite: state.favoriteIds.contains(listing.id),
                  ),
                );
              },
            ),
          ),
        ),
      );
    }
    return const SizedBox.shrink();
  }
}

class _ListingCard extends StatelessWidget {
  final Listing listing;
  final bool isFavorite;

  const _ListingCard({required this.listing, required this.isFavorite});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;
    final isAuthenticated =
        context.watch<AuthBloc>().state is AuthAuthenticated;
    final isOwner = listing.userId == context.read<TokenStorage>().userId;

    return Card(
      margin: EdgeInsets.zero,
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: () async {
          await Navigator.of(context).push(
            MaterialPageRoute(
              builder: (_) => ListingDetailScreen(listing: listing),
            ),
          );
          if (context.mounted) {
            context.read<MarketplaceCubit>().load();
          }
        },
        child: Stack(
          children: [
            Padding(
              padding: const EdgeInsets.all(10),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _ListingThumbnail(listing: listing),
                  const SizedBox(width: 12),
                  Expanded(
                    child: SizedBox(
                      height: 92,
                      child: Padding(
                        padding: const EdgeInsets.only(right: 72),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Text(
                              listing.title,
                              style: textTheme.titleMedium,
                              maxLines: 2,
                              overflow: TextOverflow.ellipsis,
                            ),
                            Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                if (listing.address.isNotEmpty)
                                  Text(
                                    listing.address,
                                    style: textTheme.bodySmall?.copyWith(
                                      color: colorScheme.onSurfaceVariant,
                                    ),
                                    overflow: TextOverflow.ellipsis,
                                  ),
                                Text(
                                  l10n.marketplacePostedOn(
                                    DateFormat.yMMMd(
                                      Localizations.localeOf(
                                        context,
                                      ).toString(),
                                    ).add_Hm().format(listing.createdAt),
                                  ),
                                  style: textTheme.bodySmall?.copyWith(
                                    color: colorScheme.onSurfaceVariant,
                                  ),
                                ),
                                const SizedBox(height: 6),
                                _InfoChip(
                                  icon: listingCategoryIcon(listing.category),
                                  label: listingCategoryLabel(
                                    l10n,
                                    listing.category,
                                  ),
                                ),
                              ],
                            ),
                          ],
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ),
            Positioned(
              top: 10,
              right: 8,
              child: Text(
                listingPriceLabel(l10n, listing.price),
                style: textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                  color: colorScheme.primary,
                ),
              ),
            ),
            if (isAuthenticated && !isOwner)
              Positioned(
                bottom: 0,
                right: 0,
                child: IconButton(
                  icon: Icon(
                    isFavorite ? Icons.favorite : Icons.favorite_border,
                  ),
                  color: isFavorite ? colorScheme.primary : null,
                  onPressed: () => context
                      .read<MarketplaceCubit>()
                      .toggleFavorite(listing.id),
                ),
              ),
          ],
        ),
      ),
    );
  }
}

class _ListingThumbnail extends StatelessWidget {
  final Listing listing;

  const _ListingThumbnail({required this.listing});

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final image = listing.images.isEmpty ? null : listing.images.first;
    final baseUrl = context.read<ApiClient>().baseUrl;

    return ClipRRect(
      borderRadius: BorderRadius.circular(8),
      child: SizedBox(
        width: 110,
        height: 92,
        child: image == null
            ? Container(
                color: colorScheme.surfaceContainerHighest,
                alignment: Alignment.center,
                child: Icon(
                  Icons.image_not_supported_outlined,
                  color: colorScheme.onSurfaceVariant,
                ),
              )
            : Image.network(
                '$baseUrl${image.url}',
                fit: BoxFit.cover,
                errorBuilder: (context, error, stack) => Container(
                  color: colorScheme.surfaceContainerHighest,
                  alignment: Alignment.center,
                  child: Icon(
                    Icons.broken_image_outlined,
                    color: colorScheme.onSurfaceVariant,
                  ),
                ),
              ),
      ),
    );
  }
}

class _InfoChip extends StatelessWidget {
  final IconData icon;
  final String label;

  const _InfoChip({required this.icon, required this.label});

  @override
  Widget build(BuildContext context) {
    final color = Theme.of(context).colorScheme.onSurfaceVariant;
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 14, color: color),
        const SizedBox(width: 4),
        Text(
          label,
          style: Theme.of(context).textTheme.bodySmall?.copyWith(color: color),
        ),
      ],
    );
  }
}
