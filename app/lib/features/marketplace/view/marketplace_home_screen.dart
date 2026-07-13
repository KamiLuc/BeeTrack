import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/api/api_client.dart';
import '../../../core/theme/app_layout.dart';
import '../../../core/widgets/app_drawer.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../features/auth/bloc/auth_bloc.dart';
import '../../../l10n/app_localizations.dart';
import '../cubit/marketplace_cubit.dart';
import '../data/listing_category.dart';
import '../data/listing_model.dart';
import '../data/listing_price.dart';
import '../data/listing_repository.dart';
import 'create_listing_screen.dart';
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
    await Navigator.of(context).push(
      MaterialPageRoute(builder: (_) => const CreateListingScreen()),
    );
    if (context.mounted) context.read<MarketplaceCubit>().load();
  }

  Future<void> _openMyListings(BuildContext context) async {
    await Navigator.of(context).push(
      MaterialPageRoute(builder: (_) => const MyListingsScreen()),
    );
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
            actions: [
              if (isAuthenticated) const ProfileIconButton(),
            ],
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
                              context
                                  .read<MarketplaceCubit>()
                                  .setKeyword(value.trim());
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
              _MarketplaceBanner(
                l10n: l10n,
                isAuthenticated: isAuthenticated,
                onAdd: () => _openCreateListing(context),
                onMyListings: () => _openMyListings(context),
                onMap: null,
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
  final VoidCallback onAdd;
  final VoidCallback onMyListings;
  final VoidCallback? onMap;

  const _MarketplaceBanner({
    required this.l10n,
    required this.isAuthenticated,
    required this.onAdd,
    required this.onMyListings,
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
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                children: [
                  if (isAuthenticated) ...[
                    IconButton(
                      icon: const Icon(Icons.add),
                      iconSize: 28,
                      tooltip: l10n.marketplaceCreateScreenTitle,
                      onPressed: onAdd,
                    ),
                    IconButton(
                      icon: const Icon(Icons.list_alt_outlined),
                      iconSize: 28,
                      tooltip: l10n.myListingsTitle,
                      onPressed: onMyListings,
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

class _ListingsFeed extends StatelessWidget {
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
      final hasMore = state.items.length < state.total;
      return RefreshIndicator(
        onRefresh: () => context.read<MarketplaceCubit>().load(),
        child: Center(
          child: ConstrainedBox(
            constraints: AppLayout.formConstraints(context),
            child: ListView.builder(
              physics: const AlwaysScrollableScrollPhysics(),
              padding: const EdgeInsets.fromLTRB(16, 4, 16, 16),
              itemCount: state.items.length + (hasMore ? 1 : 0),
              itemBuilder: (context, index) {
                if (index >= state.items.length) {
                  return Padding(
                    padding: const EdgeInsets.symmetric(vertical: 12),
                    child: Center(
                      child: state.loadingMore
                          ? const CircularProgressIndicator()
                          : OutlinedButton(
                              onPressed: () =>
                                  context.read<MarketplaceCubit>().loadMore(),
                              child: Text(l10n.generalLoadMore),
                            ),
                    ),
                  );
                }
                return Padding(
                  padding: const EdgeInsets.only(bottom: 8),
                  child: _ListingCard(listing: state.items[index]),
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

  const _ListingCard({required this.listing});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;

    return Card(
      margin: EdgeInsets.zero,
      child: InkWell(
        onTap: () => Navigator.of(context).push(
          MaterialPageRoute(
            builder: (_) => ListingDetailScreen(listing: listing),
          ),
        ),
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.fromLTRB(16, 12, 16, 12),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Expanded(
                    child: Text(
                      listing.title,
                      style: textTheme.titleMedium,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  const SizedBox(width: 8),
                  Text(
                    listingPriceLabel(l10n, listing.price),
                    style: textTheme.titleSmall?.copyWith(
                      color: colorScheme.primary,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 6),
              Wrap(
                spacing: 10,
                runSpacing: 4,
                children: [
                  _InfoChip(
                    icon: listingCategoryIcon(listing.category),
                    label: listingCategoryLabel(l10n, listing.category),
                  ),
                  if (listing.address.isNotEmpty)
                    _InfoChip(
                      icon: Icons.location_on_outlined,
                      label: listing.address,
                    ),
                ],
              ),
              if (listing.description.isNotEmpty) ...[
                const SizedBox(height: 8),
                Text(
                  listing.description,
                  style: textTheme.bodySmall?.copyWith(
                    color: colorScheme.onSurfaceVariant,
                  ),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
              ],
            ],
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
