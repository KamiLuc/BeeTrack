import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';

import '../../../core/api/api_client.dart';
import '../../../core/theme/app_layout.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../l10n/app_localizations.dart';
import '../data/favorites_repository.dart';
import '../data/listing_category.dart';
import '../data/listing_model.dart';
import '../data/listing_price.dart';
import 'listing_detail_screen.dart';

class FavoritesScreen extends StatefulWidget {
  const FavoritesScreen({super.key});

  @override
  State<FavoritesScreen> createState() => _FavoritesScreenState();
}

class _FavoritesScreenState extends State<FavoritesScreen> {
  late final FavoritesRepository _repo;
  List<Listing>? _favorites;
  bool _loading = true;
  bool _hasError = false;
  final Set<int> _busy = {};
  final Set<int> _unfavoritedIds = {};

  @override
  void initState() {
    super.initState();
    _repo = FavoritesRepository(api: context.read<ApiClient>());
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _hasError = false;
    });
    try {
      final result = await _repo.listFavorites();
      if (!mounted) return;
      setState(() {
        _favorites = result;
        _unfavoritedIds.clear();
        _loading = false;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() {
        _hasError = true;
        _loading = false;
      });
    }
  }

  Future<void> _openDetail(Listing listing) async {
    await Navigator.of(
      context,
    ).push(MaterialPageRoute(builder: (_) => ListingDetailScreen(listing: listing)));
    if (mounted) _load();
  }

  Future<void> _toggleFavorite(Listing listing) async {
    final l10n = AppLocalizations.of(context)!;
    final wasUnfavorited = _unfavoritedIds.contains(listing.id);
    setState(() {
      _busy.add(listing.id);
      if (wasUnfavorited) {
        _unfavoritedIds.remove(listing.id);
      } else {
        _unfavoritedIds.add(listing.id);
      }
    });
    try {
      if (wasUnfavorited) {
        await _repo.addFavorite(listing.id);
      } else {
        await _repo.removeFavorite(listing.id);
      }
      if (mounted) setState(() => _busy.remove(listing.id));
    } catch (_) {
      if (mounted) {
        setState(() {
          _busy.remove(listing.id);
          if (wasUnfavorited) {
            _unfavoritedIds.add(listing.id);
          } else {
            _unfavoritedIds.remove(listing.id);
          }
        });
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text(l10n.generalError)));
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.favoritesTitle),
        actions: const [ProfileIconButton()],
      ),
      body: _buildBody(l10n),
    );
  }

  Widget _buildBody(AppLocalizations l10n) {
    if (_loading) return const Center(child: CircularProgressIndicator());
    if (_hasError) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              l10n.generalError,
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 12),
            ElevatedButton(onPressed: _load, child: Text(l10n.generalRetry)),
          ],
        ),
      );
    }
    final favorites = _favorites ?? [];
    if (favorites.isEmpty) {
      return Center(child: Text(l10n.favoritesEmpty));
    }
    return RefreshIndicator(
      onRefresh: _load,
      child: Center(
        child: ConstrainedBox(
          constraints: AppLayout.formConstraints(context),
          child: ListView.builder(
            physics: const AlwaysScrollableScrollPhysics(),
            padding: const EdgeInsets.all(16),
            itemCount: favorites.length,
            itemBuilder: (context, index) {
              final listing = favorites[index];
              return Padding(
                padding: const EdgeInsets.only(bottom: 8),
                child: _FavoriteListingCard(
                  listing: listing,
                  isFavorite: !_unfavoritedIds.contains(listing.id),
                  busy: _busy.contains(listing.id),
                  onTap: () => _openDetail(listing),
                  onToggleFavorite: () => _toggleFavorite(listing),
                ),
              );
            },
          ),
        ),
      ),
    );
  }
}

class _FavoriteListingCard extends StatelessWidget {
  final Listing listing;
  final bool isFavorite;
  final bool busy;
  final VoidCallback onTap;
  final VoidCallback onToggleFavorite;

  const _FavoriteListingCard({
    required this.listing,
    required this.isFavorite,
    required this.busy,
    required this.onTap,
    required this.onToggleFavorite,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;

    return Card(
      margin: EdgeInsets.zero,
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: onTap,
        child: Stack(
          children: [
            Padding(
              padding: const EdgeInsets.all(10),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _FavoriteThumbnail(listing: listing),
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
            Positioned(
              bottom: 0,
              right: 0,
              child: busy
                  ? const Padding(
                      padding: EdgeInsets.all(12),
                      child: SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      ),
                    )
                  : IconButton(
                      icon: Icon(
                        isFavorite ? Icons.favorite : Icons.favorite_border,
                      ),
                      color: isFavorite ? colorScheme.primary : null,
                      tooltip: isFavorite
                          ? l10n.marketplaceFavoriteRemove
                          : l10n.marketplaceFavoriteAdd,
                      onPressed: onToggleFavorite,
                    ),
            ),
          ],
        ),
      ),
    );
  }
}

class _FavoriteThumbnail extends StatelessWidget {
  final Listing listing;

  const _FavoriteThumbnail({required this.listing});

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
