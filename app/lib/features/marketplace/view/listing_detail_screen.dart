import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../../core/api/api_client.dart';
import '../../../core/storage/token_storage.dart';
import '../../../core/theme/app_layout.dart';
import '../../../core/widgets/delete_dialog.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../features/auth/bloc/auth_bloc.dart';
import '../../../l10n/app_localizations.dart';
import '../../apiary/data/apiary_model.dart';
import '../../apiary/view/apiaries_map_screen.dart';
import '../data/favorites_repository.dart';
import '../data/listing_category.dart';
import '../data/listing_model.dart';
import '../data/listing_price.dart';
import '../data/listing_repository.dart';
import 'create_listing_screen.dart';

class ListingDetailScreen extends StatefulWidget {
  final Listing listing;

  const ListingDetailScreen({super.key, required this.listing});

  @override
  State<ListingDetailScreen> createState() => _ListingDetailScreenState();
}

class _ListingDetailScreenState extends State<ListingDetailScreen> {
  late Listing _listing;
  bool _isFavorite = false;
  bool _favoriteBusy = false;
  final _imagePageController = PageController();
  int _imagePage = 0;
  bool _showPhone = false;
  bool _showEmail = false;
  bool _hideBusy = false;

  @override
  void initState() {
    super.initState();
    _listing = widget.listing;
    if (context.read<AuthBloc>().state is AuthAuthenticated) {
      _loadFavoriteStatus();
    }
    _loadFullListing();
  }

  /// The listing passed in may have come from the feed's search results,
  /// which don't include detail-only fields (attached apiary's GPS/hive
  /// count), so refetch the full listing by id once the screen opens.
  Future<void> _loadFullListing() async {
    try {
      final fresh = await ListingRepository(
        api: context.read<ApiClient>(),
      ).getListing(_listing.id);
      if (mounted) setState(() => _listing = fresh);
    } catch (_) {}
  }

  @override
  void didUpdateWidget(covariant ListingDetailScreen oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (!identical(oldWidget.listing, widget.listing)) {
      _listing = widget.listing;
    }
  }

  @override
  void dispose() {
    _imagePageController.dispose();
    super.dispose();
  }

  Future<void> _loadFavoriteStatus() async {
    try {
      final isFavorite = await FavoritesRepository(
        api: context.read<ApiClient>(),
      ).checkFavorite(_listing.id);
      if (!mounted) return;
      setState(() => _isFavorite = isFavorite);
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
        await repo.addFavorite(_listing.id);
      } else {
        await repo.removeFavorite(_listing.id);
      }
    } catch (_) {
      if (mounted) {
        setState(() => _isFavorite = !next);
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text(l10n.generalError)));
      }
    } finally {
      if (mounted) setState(() => _favoriteBusy = false);
    }
  }

  Future<void> _openEdit() async {
    final updated = await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (_) => CreateListingScreen(existingListing: _listing),
      ),
    );
    if ((updated ?? false) && mounted) {
      try {
        final fresh = await ListingRepository(
          api: context.read<ApiClient>(),
        ).getListing(_listing.id);
        if (mounted) {
          setState(() {
            _listing = fresh;
            _imagePage = 0;
          });
          if (_imagePageController.hasClients) {
            _imagePageController.jumpToPage(0);
          }
        }
      } catch (_) {}
    }
  }

  void _openApiaryMap() {
    final listing = _listing;
    final l10n = AppLocalizations.of(context)!;
    final apiary = Apiary(
      id: listing.apiaryId!,
      name: listing.apiaryName ?? '',
      lat: listing.apiaryLat,
      lng: listing.apiaryLng,
      gridRows: 0,
      gridCols: 0,
      hiveCount: listing.apiaryHiveCount,
      userRole: '',
    );
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => ApiariesMapScreen(
          apiaries: [apiary],
          title: l10n.apiaryLocationTitle,
        ),
      ),
    );
  }

  Future<void> _toggleHidden() async {
    final l10n = AppLocalizations.of(context)!;
    setState(() => _hideBusy = true);
    try {
      final updated = await ListingRepository(
        api: context.read<ApiClient>(),
      ).hideListing(_listing.id, hidden: !_listing.isHidden);
      if (mounted) setState(() => _listing = updated);
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text(l10n.generalError)));
      }
    } finally {
      if (mounted) setState(() => _hideBusy = false);
    }
  }

  Future<void> _delete() async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await showDeleteDialog(
      context,
      title: l10n.marketplaceDeleteConfirm,
      warning: l10n.marketplaceDeleteWarning,
      l10n: l10n,
      withPuzzle: true,
    );
    if (!confirmed || !mounted) return;
    try {
      await ListingRepository(
        api: context.read<ApiClient>(),
      ).deleteListing(_listing.id);
      if (mounted) Navigator.of(context).pop();
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text(l10n.generalError)));
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final listing = _listing;
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;
    final isOwner = listing.userId == context.read<TokenStorage>().userId;

    final isAuthenticated = context.read<AuthBloc>().state is AuthAuthenticated;

    return Scaffold(
      appBar: AppBar(
        title: Text(listing.title, overflow: TextOverflow.ellipsis),
        actions: [if (isAuthenticated) const ProfileIconButton()],
      ),
      body: Column(
        children: [
          Expanded(
            child: SingleChildScrollView(
              child: Center(
                child: ConstrainedBox(
                  constraints: const BoxConstraints(maxWidth: 600),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Padding(
                        padding: const EdgeInsets.fromLTRB(12, 12, 12, 0),
                        child: _ImageCarousel(
                          images: listing.images,
                          controller: _imagePageController,
                          page: _imagePage,
                          onPageChanged: (page) =>
                              setState(() => _imagePage = page),
                        ),
                      ),
                      Padding(
                        padding: const EdgeInsets.all(16),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Card(
                              margin: EdgeInsets.zero,
                              child: Padding(
                                padding: const EdgeInsets.all(16),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Row(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        Expanded(
                                          child: Text(
                                            l10n.marketplacePostedOn(
                                              DateFormat.yMMMd(
                                                Localizations.localeOf(
                                                  context,
                                                ).toString(),
                                              ).add_Hm().format(
                                                listing.createdAt,
                                              ),
                                            ),
                                            style: textTheme.bodySmall
                                                ?.copyWith(
                                                  color: colorScheme
                                                      .onSurfaceVariant,
                                                ),
                                          ),
                                        ),
                                        if (!isOwner)
                                          BlocBuilder<AuthBloc, AuthState>(
                                            builder: (context, authState) {
                                              if (authState
                                                  is! AuthAuthenticated) {
                                                return const SizedBox.shrink();
                                              }
                                              return InkWell(
                                                customBorder:
                                                    const CircleBorder(),
                                                onTap: _favoriteBusy
                                                    ? null
                                                    : _toggleFavorite,
                                                child: Padding(
                                                  padding: const EdgeInsets.all(
                                                    4,
                                                  ),
                                                  child: Icon(
                                                    _isFavorite
                                                        ? Icons.favorite
                                                        : Icons.favorite_border,
                                                    color: _isFavorite
                                                        ? colorScheme.primary
                                                        : colorScheme
                                                              .onSurfaceVariant,
                                                  ),
                                                ),
                                              );
                                            },
                                          ),
                                      ],
                                    ),
                                    const SizedBox(height: 12),
                                    Text(
                                      listing.title,
                                      style: textTheme.headlineSmall?.copyWith(
                                        fontWeight: FontWeight.bold,
                                      ),
                                    ),
                                    const SizedBox(height: 8),
                                    Text(
                                      listingPriceLabel(l10n, listing.price),
                                      style: textTheme.headlineSmall?.copyWith(
                                        fontWeight: FontWeight.bold,
                                        color: colorScheme.primary,
                                      ),
                                    ),
                                    const SizedBox(height: 12),
                                    Wrap(
                                      spacing: 10,
                                      runSpacing: 4,
                                      children: [
                                        _InfoChip(
                                          icon: listingCategoryIcon(
                                            listing.category,
                                          ),
                                          label: listingCategoryLabel(
                                            l10n,
                                            listing.category,
                                          ),
                                        ),
                                        if (listing.address.isNotEmpty)
                                          _InfoChip(
                                            icon: Icons.location_on_outlined,
                                            label: listing.address,
                                          ),
                                        if (listing.quantity.isNotEmpty)
                                          _InfoChip(
                                            icon: Icons.inventory_2_outlined,
                                            label:
                                                '${l10n.marketplaceQuantityLabel}: ${listing.quantity}',
                                          ),
                                      ],
                                    ),
                                  ],
                                ),
                              ),
                            ),
                            if (listing.contactPhone.isNotEmpty ||
                                listing.contactEmail.isNotEmpty) ...[
                              const SizedBox(height: 20),
                              Row(
                                children: [
                                  if (listing.contactPhone.isNotEmpty)
                                    Expanded(
                                      child: _RevealButton(
                                        icon: Icons.call_outlined,
                                        label: l10n.marketplaceCallButton,
                                        value: listing.contactPhone,
                                        revealed: _showPhone,
                                        onReveal: () =>
                                            setState(() => _showPhone = true),
                                      ),
                                    ),
                                  if (listing.contactPhone.isNotEmpty &&
                                      listing.contactEmail.isNotEmpty)
                                    const SizedBox(width: 12),
                                  if (listing.contactEmail.isNotEmpty)
                                    Expanded(
                                      child: _RevealButton(
                                        icon: Icons.email_outlined,
                                        label: l10n.marketplaceWriteButton,
                                        value: listing.contactEmail,
                                        revealed: _showEmail,
                                        onReveal: () =>
                                            setState(() => _showEmail = true),
                                      ),
                                    ),
                                ],
                              ),
                            ],
                            if (listing.description.isNotEmpty) ...[
                              const SizedBox(height: 20),
                              _DetailSection(
                                icon: Icons.description_outlined,
                                title: l10n.marketplaceDescriptionLabel,
                                child: Text(
                                  listing.description,
                                  style: textTheme.bodyMedium,
                                ),
                              ),
                            ],
                            if (listing.apiaryName != null &&
                                listing.apiaryName!.isNotEmpty) ...[
                              const SizedBox(height: 12),
                              _DetailSection(
                                icon: Icons.location_city_outlined,
                                title: l10n.marketplaceApiaryLabel,
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      listing.apiaryName!,
                                      style: textTheme.bodyMedium,
                                    ),
                                    const SizedBox(height: 8),
                                    Row(
                                      children: [
                                        _InfoChip(
                                          icon: Icons.hive_outlined,
                                          label: l10n.hiveCount(
                                            listing.apiaryHiveCount,
                                          ),
                                        ),
                                        if (listing.apiaryLat != null &&
                                            listing.apiaryLng != null) ...[
                                          const Spacer(),
                                          OutlinedButton.icon(
                                            onPressed: _openApiaryMap,
                                            icon: const Icon(
                                              Icons.map_outlined,
                                              size: 18,
                                            ),
                                            label: Text(
                                              l10n.apiaryMapTooltip,
                                            ),
                                          ),
                                        ],
                                      ],
                                    ),
                                  ],
                                ),
                              ),
                            ],
                            if (listing.honeyBatch != null) ...[
                              const SizedBox(height: 12),
                              _HoneyBatchSection(batch: listing.honeyBatch!),
                            ],
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
          if (isOwner)
            _EditBanner(
              onEdit: _openEdit,
              onDelete: _delete,
              isHidden: listing.isHidden,
              hideBusy: _hideBusy,
              onToggleHidden: _toggleHidden,
            ),
        ],
      ),
    );
  }
}

class _EditBanner extends StatelessWidget {
  final VoidCallback onEdit;
  final VoidCallback onDelete;
  final bool isHidden;
  final bool hideBusy;
  final VoidCallback onToggleHidden;

  const _EditBanner({
    required this.onEdit,
    required this.onDelete,
    required this.isHidden,
    required this.hideBusy,
    required this.onToggleHidden,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
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
                  IconButton(
                    icon: const Icon(Icons.edit_outlined),
                    iconSize: 28,
                    tooltip: l10n.generalEdit,
                    onPressed: onEdit,
                  ),
                  hideBusy
                      ? const SizedBox(
                          width: 28,
                          height: 28,
                          child: Padding(
                            padding: EdgeInsets.all(2),
                            child: CircularProgressIndicator(strokeWidth: 2),
                          ),
                        )
                      : IconButton(
                          icon: Icon(
                            isHidden
                                ? Icons.visibility_outlined
                                : Icons.visibility_off_outlined,
                          ),
                          iconSize: 28,
                          tooltip: isHidden
                              ? l10n.marketplaceShowListing
                              : l10n.marketplaceHideListing,
                          onPressed: onToggleHidden,
                        ),
                  IconButton(
                    icon: const Icon(Icons.delete_outline),
                    iconSize: 28,
                    tooltip: l10n.generalDelete,
                    onPressed: onDelete,
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

class _RevealButton extends StatelessWidget {
  final IconData icon;
  final String label;
  final String value;
  final bool revealed;
  final VoidCallback onReveal;

  const _RevealButton({
    required this.icon,
    required this.label,
    required this.value,
    required this.revealed,
    required this.onReveal,
  });

  @override
  Widget build(BuildContext context) {
    return OutlinedButton.icon(
      onPressed: revealed ? null : onReveal,
      icon: Icon(icon, size: 18),
      label: Text(revealed ? value : label, overflow: TextOverflow.ellipsis),
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

  void _openFullscreen(BuildContext context, String baseUrl) {
    Navigator.of(context).push(
      PageRouteBuilder(
        opaque: true,
        barrierColor: Colors.black,
        pageBuilder: (context, _, __) => _FullscreenGallery(
          images: images,
          baseUrl: baseUrl,
          initialPage: page,
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final baseUrl = context.read<ApiClient>().baseUrl;

    if (images.isEmpty) {
      return ClipRRect(
        borderRadius: BorderRadius.circular(16),
        child: Container(
          height: 280,
          color: colorScheme.surfaceContainerHighest,
          alignment: Alignment.center,
          child: Icon(
            Icons.image_not_supported_outlined,
            size: 48,
            color: colorScheme.onSurfaceVariant,
          ),
        ),
      );
    }

    return ClipRRect(
      borderRadius: BorderRadius.circular(16),
      child: Stack(
        alignment: Alignment.center,
        children: [
          SizedBox(
            height: 280,
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
          if (images.length > 1) ...[
            Positioned(
              left: 8,
              child: _NavArrow(
                icon: Icons.chevron_left,
                onTap: page > 0
                    ? () => controller.previousPage(
                        duration: const Duration(milliseconds: 200),
                        curve: Curves.easeOut,
                      )
                    : null,
              ),
            ),
            Positioned(
              right: 8,
              child: _NavArrow(
                icon: Icons.chevron_right,
                onTap: page < images.length - 1
                    ? () => controller.nextPage(
                        duration: const Duration(milliseconds: 200),
                        curve: Curves.easeOut,
                      )
                    : null,
              ),
            ),
          ],
          Positioned(
            right: 12,
            bottom: 12,
            child: _CarouselIconButton(
              icon: Icons.open_in_full,
              onTap: () => _openFullscreen(context, baseUrl),
            ),
          ),
        ],
      ),
    );
  }
}

class _NavArrow extends StatelessWidget {
  final IconData icon;
  final VoidCallback? onTap;

  const _NavArrow({required this.icon, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return _CarouselIconButton(icon: icon, onTap: onTap);
  }
}

class _CarouselIconButton extends StatelessWidget {
  final IconData icon;
  final VoidCallback? onTap;

  const _CarouselIconButton({required this.icon, required this.onTap});

  @override
  Widget build(BuildContext context) {
    final enabled = onTap != null;
    return Material(
      color: Colors.white.withValues(alpha: enabled ? 0.85 : 0.4),
      shape: const CircleBorder(),
      child: InkWell(
        customBorder: const CircleBorder(),
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.all(8),
          child: Icon(
            icon,
            size: 20,
            color: enabled ? Colors.black87 : Colors.black38,
          ),
        ),
      ),
    );
  }
}

class _FullscreenGallery extends StatefulWidget {
  final List<ListingImage> images;
  final String baseUrl;
  final int initialPage;

  const _FullscreenGallery({
    required this.images,
    required this.baseUrl,
    required this.initialPage,
  });

  @override
  State<_FullscreenGallery> createState() => _FullscreenGalleryState();
}

class _FullscreenGalleryState extends State<_FullscreenGallery> {
  late final PageController _controller;

  @override
  void initState() {
    super.initState();
    _controller = PageController(initialPage: widget.initialPage);
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black,
      body: Stack(
        children: [
          PageView.builder(
            controller: _controller,
            itemCount: widget.images.length,
            itemBuilder: (context, index) => InteractiveViewer(
              child: Center(
                child: Image.network(
                  '${widget.baseUrl}${widget.images[index].url}',
                  fit: BoxFit.contain,
                  errorBuilder: (context, error, stack) => const Icon(
                    Icons.broken_image_outlined,
                    size: 48,
                    color: Colors.white54,
                  ),
                ),
              ),
            ),
          ),
          Positioned(
            top: 8,
            right: 8,
            child: SafeArea(
              child: IconButton(
                icon: const Icon(Icons.close, color: Colors.white),
                onPressed: () => Navigator.of(context).pop(),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _HoneyBatchSection extends StatelessWidget {
  final ListingHoneyBatch batch;

  const _HoneyBatchSection({required this.batch});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final dateLabel = DateFormat.yMMMd(
      Localizations.localeOf(context).toString(),
    ).format(batch.gatheringDate);

    return _DetailSection(
      icon: Icons.verified_outlined,
      title: l10n.marketplaceHoneyBatchSectionTitle,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Wrap(
            spacing: 10,
            runSpacing: 4,
            children: [
              _InfoChip(icon: Icons.water_drop_outlined, label: batch.honeyType),
              _InfoChip(
                icon: Icons.event_outlined,
                label: dateLabel,
              ),
              _InfoChip(
                icon: Icons.scale_outlined,
                label: '${batch.amountKg.toStringAsFixed(1)} kg',
              ),
            ],
          ),
          const SizedBox(height: 12),
          SizedBox(
            width: double.infinity,
            child: OutlinedButton.icon(
              onPressed: () => _launchExternal(batch.verificationUrl),
              icon: const Icon(Icons.link, size: 18),
              label: Text(
                l10n.honeyBatchOpenPublicPage,
                textAlign: TextAlign.center,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

Future<void> _launchExternal(String url) {
  return launchUrl(Uri.parse(url), webOnlyWindowName: '_blank');
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
