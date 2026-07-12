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
import '../data/listing_repository.dart';

class MarketplaceHomeScreen extends StatelessWidget {
  /// Called when an authenticated user picks a section from the drawer.
  final ValueChanged<AppSection>? onSelectSection;

  /// Called when an unauthenticated user taps "Log in" in the drawer.
  final VoidCallback? onLogin;

  const MarketplaceHomeScreen({
    super.key,
    this.onSelectSection,
    this.onLogin,
  });

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => MarketplaceCubit(
        repo: ListingRepository(api: context.read<ApiClient>()),
      )..load(),
      child: _MarketplaceView(onSelectSection: onSelectSection, onLogin: onLogin),
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

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
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
              IconButton(
                icon: const Icon(Icons.map_outlined),
                tooltip: l10n.marketplaceMapTooltip,
                onPressed: null,
              ),
              if (isAuthenticated) const ProfileIconButton(),
            ],
          ),
          drawer: drawer,
          body: Column(
            children: [
              Padding(
                padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
                child: TextField(
                  controller: _searchController,
                  decoration: InputDecoration(
                    hintText: l10n.marketplaceSearchHint,
                    prefixIcon: const Icon(Icons.search),
                    isDense: true,
                    border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
                  ),
                  textInputAction: TextInputAction.search,
                  onSubmitted: (value) =>
                      context.read<MarketplaceCubit>().setKeyword(value.trim()),
                ),
              ),
              const _CategoryChips(),
              const SizedBox(height: 4),
              Expanded(child: _ListingsFeed()),
            ],
          ),
        );
      },
    );
  }
}

class _CategoryChips extends StatelessWidget {
  const _CategoryChips();

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final state = context.watch<MarketplaceCubit>().state;
    final selectedCategory = state is MarketplaceLoaded ? state.category : null;

    return SizedBox(
      height: 40,
      child: ListView(
        scrollDirection: Axis.horizontal,
        padding: const EdgeInsets.symmetric(horizontal: 12),
        children: [
          _CategoryChip(
            label: l10n.marketplaceCategoryAll,
            selected: selectedCategory == null,
            onSelected: () => context.read<MarketplaceCubit>().setCategory(null),
          ),
          for (final category in listingCategories)
            _CategoryChip(
              label: listingCategoryLabel(l10n, category),
              selected: selectedCategory == category,
              onSelected: () => context.read<MarketplaceCubit>().setCategory(category),
            ),
        ],
      ),
    );
  }
}

class _CategoryChip extends StatelessWidget {
  final String label;
  final bool selected;
  final VoidCallback onSelected;

  const _CategoryChip({
    required this.label,
    required this.selected,
    required this.onSelected,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 4),
      child: ChoiceChip(
        label: Text(label),
        selected: selected,
        onSelected: (_) => onSelected(),
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
            Text(l10n.generalError, style: Theme.of(context).textTheme.titleMedium),
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
                              onPressed: () => context.read<MarketplaceCubit>().loadMore(),
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
                  listing.price != null
                      ? '${listing.price!.toStringAsFixed(2)} zł'
                      : l10n.marketplacePriceOnRequest,
                  style: textTheme.titleSmall?.copyWith(color: colorScheme.primary),
                ),
              ],
            ),
            const SizedBox(height: 6),
            Wrap(
              spacing: 10,
              runSpacing: 4,
              children: [
                _InfoChip(
                  icon: Icons.sell_outlined,
                  label: listingCategoryLabel(l10n, listing.category),
                ),
                if (listing.address.isNotEmpty)
                  _InfoChip(icon: Icons.location_on_outlined, label: listing.address),
              ],
            ),
            if (listing.description.isNotEmpty) ...[
              const SizedBox(height: 8),
              Text(
                listing.description,
                style: textTheme.bodySmall?.copyWith(color: colorScheme.onSurfaceVariant),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ],
          ],
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
        Text(label, style: Theme.of(context).textTheme.bodySmall?.copyWith(color: color)),
      ],
    );
  }
}
