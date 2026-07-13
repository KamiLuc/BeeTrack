import 'dart:math';

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/api/api_client.dart';
import '../../../core/theme/app_layout.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../l10n/app_localizations.dart';
import '../data/listing_category.dart';
import '../data/listing_model.dart';
import '../data/listing_price.dart';
import '../data/listing_repository.dart';
import 'create_listing_screen.dart';

const int _pageSize = 20;

class MyListingsScreen extends StatefulWidget {
  const MyListingsScreen({super.key});

  @override
  State<MyListingsScreen> createState() => _MyListingsScreenState();
}

class _MyListingsScreenState extends State<MyListingsScreen> {
  late final ListingRepository _repo;
  List<Listing>? _listings;
  int _currentPage = 1;
  int _totalPages = 1;
  bool _loading = true;
  bool _hasError = false;
  final Set<int> _busy = {};

  @override
  void initState() {
    super.initState();
    _repo = ListingRepository(api: context.read<ApiClient>());
    _load();
  }

  Future<void> _load() => _goToPage(1);

  Future<void> _goToPage(int page) async {
    setState(() {
      _loading = true;
      _hasError = false;
    });
    try {
      final result = await _repo.searchListings(
        mine: true,
        limit: _pageSize,
        offset: (page - 1) * _pageSize,
      );
      if (!mounted) return;
      setState(() {
        _listings = result.items;
        _currentPage = page;
        _totalPages = (result.total / _pageSize).ceil().clamp(1, 999999);
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

  Future<void> _openEdit(Listing listing) async {
    final updated = await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (_) => CreateListingScreen(existingListing: listing),
      ),
    );
    if ((updated ?? false) && mounted) _goToPage(_currentPage);
  }

  Future<void> _toggleHidden(Listing listing) async {
    final l10n = AppLocalizations.of(context)!;
    setState(() => _busy.add(listing.id));
    try {
      final updated = await _repo.hideListing(
        listing.id,
        hidden: !listing.isHidden,
      );
      if (!mounted) return;
      setState(() {
        _listings = [
          for (final l in _listings!)
            if (l.id == updated.id) updated else l,
        ];
        _busy.remove(listing.id);
      });
    } catch (_) {
      if (mounted) {
        setState(() => _busy.remove(listing.id));
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text(l10n.generalError)));
      }
    }
  }

  Future<void> _confirmDelete(Listing listing) async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(l10n.marketplaceDeleteConfirm),
        content: Text(l10n.marketplaceDeleteWarning),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(false),
            child: Text(l10n.generalCancel),
          ),
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(true),
            child: Text(
              l10n.generalDelete,
              style: TextStyle(color: Theme.of(ctx).colorScheme.error),
            ),
          ),
        ],
      ),
    );
    if (confirmed != true || !mounted) return;

    setState(() => _busy.add(listing.id));
    try {
      await _repo.deleteListing(listing.id);
      if (!mounted) return;
      final remaining = _listings!.where((l) => l.id != listing.id).toList();
      if (remaining.isEmpty && _currentPage > 1) {
        await _goToPage(_currentPage - 1);
      } else {
        setState(() {
          _listings = remaining;
          _busy.remove(listing.id);
        });
      }
    } catch (_) {
      if (mounted) {
        setState(() => _busy.remove(listing.id));
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
        title: Text(l10n.myListingsTitle),
        actions: const [ProfileIconButton()],
      ),
      body: Column(
        children: [
          Expanded(child: _buildBody(l10n)),
          if (_totalPages > 1)
            _PaginationBanner(
              currentPage: _currentPage,
              totalPages: _totalPages,
              onPage: _goToPage,
            ),
        ],
      ),
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
    final listings = _listings ?? [];
    if (listings.isEmpty) {
      return Center(child: Text(l10n.myListingsEmpty));
    }
    return RefreshIndicator(
      onRefresh: _load,
      child: Center(
        child: ConstrainedBox(
          constraints: AppLayout.formConstraints(context),
          child: ListView.builder(
            physics: const AlwaysScrollableScrollPhysics(),
            padding: const EdgeInsets.all(16),
            itemCount: listings.length,
            itemBuilder: (context, index) {
              final listing = listings[index];
              return Padding(
                padding: const EdgeInsets.only(bottom: 8),
                child: _MyListingCard(
                  listing: listing,
                  busy: _busy.contains(listing.id),
                  onEdit: () => _openEdit(listing),
                  onToggleHidden: () => _toggleHidden(listing),
                  onDelete: () => _confirmDelete(listing),
                ),
              );
            },
          ),
        ),
      ),
    );
  }
}

class _PaginationBanner extends StatelessWidget {
  final int currentPage;
  final int totalPages;
  final ValueChanged<int> onPage;

  const _PaginationBanner({
    required this.currentPage,
    required this.totalPages,
    required this.onPage,
  });

  List<_PageItem> _buildPageItems() {
    final cur = currentPage;
    final last = totalPages;
    final items = <_PageItem>[];

    void addPage(int p) => items.add(_PageItem.page(p));
    void addEllipsis() => items.add(_PageItem.ellipsis());

    if (last <= 5) {
      for (var i = 1; i <= last; i++) addPage(i);
    } else {
      addPage(1);
      if (cur > 3) addEllipsis();
      for (var i = max(2, cur - 1); i <= min(last - 1, cur + 1); i++) {
        addPage(i);
      }
      if (cur < last - 2) addEllipsis();
      addPage(last);
    }
    return items;
  }

  @override
  Widget build(BuildContext context) {
    final bannerWidth = AppLayout.bannerWidth(context);
    final pageItems = _buildPageItems();

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
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  IconButton(
                    icon: const Icon(Icons.chevron_left),
                    iconSize: 24,
                    onPressed: currentPage > 1
                        ? () => onPage(currentPage - 1)
                        : null,
                  ),
                  for (final item in pageItems)
                    item.isEllipsis
                        ? const Padding(
                            padding: EdgeInsets.symmetric(horizontal: 2),
                            child: Text('…', style: TextStyle(fontSize: 16)),
                          )
                        : _PageButton(
                            page: item.page!,
                            isCurrent: item.page == currentPage,
                            onTap: () => onPage(item.page!),
                          ),
                  IconButton(
                    icon: const Icon(Icons.chevron_right),
                    iconSize: 24,
                    onPressed: currentPage < totalPages
                        ? () => onPage(currentPage + 1)
                        : null,
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

class _PageItem {
  final int? page;
  final bool isEllipsis;

  const _PageItem.page(this.page) : isEllipsis = false;
  const _PageItem.ellipsis() : page = null, isEllipsis = true;
}

class _PageButton extends StatelessWidget {
  final int page;
  final bool isCurrent;
  final VoidCallback onTap;

  const _PageButton({
    required this.page,
    required this.isCurrent,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: isCurrent ? null : onTap,
      borderRadius: BorderRadius.circular(16),
      child: Container(
        width: 32,
        height: 32,
        decoration: isCurrent
            ? BoxDecoration(
                color: Colors.black26,
                borderRadius: BorderRadius.circular(16),
              )
            : null,
        alignment: Alignment.center,
        child: Text(
          '$page',
          style: TextStyle(
            fontSize: 14,
            fontWeight: isCurrent ? FontWeight.bold : FontWeight.normal,
          ),
        ),
      ),
    );
  }
}

enum _CardAction { edit, toggleHidden, delete }

class _MyListingCard extends StatelessWidget {
  final Listing listing;
  final bool busy;
  final VoidCallback onEdit;
  final VoidCallback onToggleHidden;
  final VoidCallback onDelete;

  const _MyListingCard({
    required this.listing,
    required this.busy,
    required this.onEdit,
    required this.onToggleHidden,
    required this.onDelete,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;

    return Card(
      margin: EdgeInsets.zero,
      clipBehavior: Clip.antiAlias,
      child: Stack(
        children: [
          Padding(
            padding: const EdgeInsets.all(10),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _MyListingThumbnail(listing: listing),
                const SizedBox(width: 12),
                Expanded(
                  child: SizedBox(
                    height: 92,
                    child: Padding(
                      padding: const EdgeInsets.only(right: 40),
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
                          Wrap(
                            spacing: 10,
                            runSpacing: 4,
                            children: [
                              _Chip(
                                icon: listingCategoryIcon(listing.category),
                                label: listingCategoryLabel(
                                  l10n,
                                  listing.category,
                                ),
                              ),
                              if (listing.isHidden)
                                _Chip(
                                  icon: Icons.visibility_off_outlined,
                                  label: l10n.marketplaceHiddenBadge,
                                  color: colorScheme.error,
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
            bottom: 8,
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
            top: busy ? 14 : 4,
            right: busy ? 16 : 4,
            child: busy
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : PopupMenuButton<_CardAction>(
                    onSelected: (action) {
                      switch (action) {
                        case _CardAction.edit:
                          onEdit();
                        case _CardAction.toggleHidden:
                          onToggleHidden();
                        case _CardAction.delete:
                          onDelete();
                      }
                    },
                    itemBuilder: (_) => [
                      PopupMenuItem(
                        value: _CardAction.edit,
                        child: Text(l10n.generalEdit),
                      ),
                      PopupMenuItem(
                        value: _CardAction.toggleHidden,
                        child: Text(
                          listing.isHidden
                              ? l10n.marketplaceShowListing
                              : l10n.marketplaceHideListing,
                        ),
                      ),
                      PopupMenuItem(
                        value: _CardAction.delete,
                        child: Text(
                          l10n.generalDelete,
                          style: TextStyle(color: colorScheme.error),
                        ),
                      ),
                    ],
                  ),
          ),
        ],
      ),
    );
  }
}

class _MyListingThumbnail extends StatelessWidget {
  final Listing listing;

  const _MyListingThumbnail({required this.listing});

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

class _Chip extends StatelessWidget {
  final IconData icon;
  final String label;
  final Color? color;

  const _Chip({required this.icon, required this.label, this.color});

  @override
  Widget build(BuildContext context) {
    final c = color ?? Theme.of(context).colorScheme.onSurfaceVariant;
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 14, color: c),
        const SizedBox(width: 4),
        Text(
          label,
          style: Theme.of(context).textTheme.bodySmall?.copyWith(color: c),
        ),
      ],
    );
  }
}
