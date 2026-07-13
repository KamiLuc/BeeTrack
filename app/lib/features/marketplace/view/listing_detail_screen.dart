import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';

import '../../../core/api/api_client.dart';
import '../../../features/auth/bloc/auth_bloc.dart';
import '../../../l10n/app_localizations.dart';
import '../data/favorites_repository.dart';
import '../data/listing_category.dart';
import '../data/listing_model.dart';
import '../data/listing_price.dart';

class ListingDetailScreen extends StatefulWidget {
  final Listing listing;

  const ListingDetailScreen({super.key, required this.listing});

  @override
  State<ListingDetailScreen> createState() => _ListingDetailScreenState();
}

class _ListingDetailScreenState extends State<ListingDetailScreen> {
  bool _isFavorite = false;
  bool _favoriteBusy = false;
  final _imagePageController = PageController();
  int _imagePage = 0;

  @override
  void initState() {
    super.initState();
    if (context.read<AuthBloc>().state is AuthAuthenticated) {
      _loadFavoriteStatus();
    }
  }

  @override
  void dispose() {
    _imagePageController.dispose();
    super.dispose();
  }

  Future<void> _loadFavoriteStatus() async {
    try {
      final favorites = await FavoritesRepository(api: context.read<ApiClient>())
          .listFavorites();
      if (!mounted) return;
      setState(() {
        _isFavorite = favorites.any((l) => l.id == widget.listing.id);
      });
    } catch (_) {}
  }

  Future<void> _toggleFavorite() async {
    final l10n = AppLocalizations.of(context)!;
    final repo = FavoritesRepository(api: context.read<ApiClient>());
    final next = !_isFavorite;
    setState(() {
      _isFavorite = next;
      _favoriteBusy = true;
    });
    try {
      if (next) {
        await repo.addFavorite(widget.listing.id);
      } else {
        await repo.removeFavorite(widget.listing.id);
      }
    } catch (_) {
      if (mounted) {
        setState(() => _isFavorite = !next);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.generalError)),
        );
      }
    } finally {
      if (mounted) setState(() => _favoriteBusy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final listing = widget.listing;

    return Scaffold(
      appBar: AppBar(
        title: Text(listing.title, overflow: TextOverflow.ellipsis),
        actions: [
          BlocBuilder<AuthBloc, AuthState>(
            builder: (context, authState) {
              if (authState is! AuthAuthenticated) return const SizedBox.shrink();
              return IconButton(
                icon: Icon(_isFavorite ? Icons.favorite : Icons.favorite_border),
                color: _isFavorite ? Theme.of(context).colorScheme.primary : null,
                tooltip: _isFavorite
                    ? l10n.marketplaceFavoriteRemove
                    : l10n.marketplaceFavoriteAdd,
                onPressed: _favoriteBusy ? null : _toggleFavorite,
              );
            },
          ),
        ],
      ),
      body: SingleChildScrollView(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 600),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _ImageCarousel(
                  images: listing.images,
                  controller: _imagePageController,
                  page: _imagePage,
                  onPageChanged: (page) => setState(() => _imagePage = page),
                ),
                Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Expanded(
                            child: Text(
                              listing.title,
                              style: Theme.of(context).textTheme.titleLarge,
                            ),
                          ),
                          const SizedBox(width: 8),
                          Text(
                            listingPriceLabel(l10n, listing.price),
                            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                                  color: Theme.of(context).colorScheme.primary,
                                ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 8),
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
                          if (listing.quantity.isNotEmpty)
                            _InfoChip(
                              icon: Icons.inventory_2_outlined,
                              label: '${l10n.marketplaceQuantityLabel}: ${listing.quantity}',
                            ),
                        ],
                      ),
                      const SizedBox(height: 8),
                      Text(
                        l10n.marketplacePostedOn(
                          DateFormat.yMMMd(
                            Localizations.localeOf(context).toString(),
                          ).format(listing.createdAt),
                        ),
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: Theme.of(context).colorScheme.onSurfaceVariant,
                            ),
                      ),
                      if (listing.description.isNotEmpty) ...[
                        const SizedBox(height: 16),
                        _DetailSection(
                          icon: Icons.description_outlined,
                          title: l10n.marketplaceDescriptionLabel,
                          child: Text(
                            listing.description,
                            style: Theme.of(context).textTheme.bodyMedium,
                          ),
                        ),
                      ],
                      const SizedBox(height: 12),
                      _DetailSection(
                        icon: Icons.contact_phone_outlined,
                        title: l10n.marketplaceContactLabel,
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            if (listing.contactPhone.isNotEmpty)
                              _ContactRow(
                                icon: Icons.phone_outlined,
                                text: listing.contactPhone,
                              ),
                            if (listing.contactEmail.isNotEmpty)
                              _ContactRow(
                                icon: Icons.email_outlined,
                                text: listing.contactEmail,
                              ),
                          ],
                        ),
                      ),
                      if (listing.apiaryName != null &&
                          listing.apiaryName!.isNotEmpty) ...[
                        const SizedBox(height: 12),
                        _DetailSection(
                          icon: Icons.location_city_outlined,
                          title: l10n.marketplaceApiaryLabel,
                          child: Text(
                            listing.apiaryName!,
                            style: Theme.of(context).textTheme.bodyMedium,
                          ),
                        ),
                      ],
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _ImageCarousel extends StatelessWidget {
  final List<ListingImage> images;
  final PageController controller;
  final int page;
  final ValueChanged<int> onPageChanged;

  const _ImageCarousel({
    required this.images,
    required this.controller,
    required this.page,
    required this.onPageChanged,
  });

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final baseUrl = context.read<ApiClient>().baseUrl;

    if (images.isEmpty) {
      return Container(
        height: 240,
        color: colorScheme.surfaceContainerHighest,
        alignment: Alignment.center,
        child: Icon(
          Icons.image_not_supported_outlined,
          size: 48,
          color: colorScheme.onSurfaceVariant,
        ),
      );
    }

    return Stack(
      alignment: Alignment.bottomCenter,
      children: [
        SizedBox(
          height: 240,
          child: PageView.builder(
            controller: controller,
            itemCount: images.length,
            onPageChanged: onPageChanged,
            itemBuilder: (context, index) => Image.network(
              '$baseUrl${images[index].url}',
              fit: BoxFit.cover,
              width: double.infinity,
              errorBuilder: (context, error, stack) => Container(
                color: colorScheme.surfaceContainerHighest,
                alignment: Alignment.center,
                child: Icon(
                  Icons.broken_image_outlined,
                  size: 48,
                  color: colorScheme.onSurfaceVariant,
                ),
              ),
            ),
          ),
        ),
        if (images.length > 1)
          Padding(
            padding: const EdgeInsets.only(bottom: 8),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                for (var i = 0; i < images.length; i++)
                  Container(
                    margin: const EdgeInsets.symmetric(horizontal: 3),
                    width: 7,
                    height: 7,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      color: i == page
                          ? colorScheme.primary
                          : colorScheme.surface.withValues(alpha: 0.7),
                    ),
                  ),
              ],
            ),
          ),
      ],
    );
  }
}

class _DetailSection extends StatelessWidget {
  final IconData icon;
  final String title;
  final Widget child;

  const _DetailSection({
    required this.icon,
    required this.title,
    required this.child,
  });

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;

    return Card(
      margin: EdgeInsets.zero,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(icon, size: 20, color: colorScheme.primary),
                const SizedBox(width: 8),
                Expanded(child: Text(title, style: textTheme.titleMedium)),
              ],
            ),
            const SizedBox(height: 12),
            child,
          ],
        ),
      ),
    );
  }
}

class _ContactRow extends StatelessWidget {
  final IconData icon;
  final String text;

  const _ContactRow({required this.icon, required this.text});

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    return Padding(
      padding: const EdgeInsets.only(bottom: 4),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 16, color: colorScheme.onSurfaceVariant),
          const SizedBox(width: 8),
          Text(text, style: Theme.of(context).textTheme.bodyMedium),
        ],
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
