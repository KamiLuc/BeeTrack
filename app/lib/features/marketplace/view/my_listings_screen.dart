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

class MyListingsScreen extends StatefulWidget {
  const MyListingsScreen({super.key});

  @override
  State<MyListingsScreen> createState() => _MyListingsScreenState();
}

class _MyListingsScreenState extends State<MyListingsScreen> {
  late final ListingRepository _repo;
  List<Listing>? _listings;
  bool _loading = true;
  bool _hasError = false;
  final Set<int> _busy = {};

  @override
  void initState() {
    super.initState();
    _repo = ListingRepository(api: context.read<ApiClient>());
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _hasError = false;
    });
    try {
      final result = await _repo.searchListings(mine: true, limit: 100);
      if (!mounted) return;
      setState(() {
        _listings = result.items;
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
    if ((updated ?? false) && mounted) _load();
  }

  Future<void> _toggleHidden(Listing listing) async {
    final l10n = AppLocalizations.of(context)!;
    setState(() => _busy.add(listing.id));
    try {
      final updated =
          await _repo.hideListing(listing.id, hidden: !listing.isHidden);
      if (!mounted) return;
      setState(() {
        _listings = [
          for (final l in _listings!) if (l.id == updated.id) updated else l,
        ];
        _busy.remove(listing.id);
      });
    } catch (_) {
      if (mounted) {
        setState(() => _busy.remove(listing.id));
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.generalError)),
        );
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
      setState(() {
        _listings = _listings!.where((l) => l.id != listing.id).toList();
        _busy.remove(listing.id);
      });
    } catch (_) {
      if (mounted) {
        setState(() => _busy.remove(listing.id));
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.generalError)),
        );
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
            Text(l10n.generalError, style: Theme.of(context).textTheme.titleMedium),
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
      child: Padding(
        padding: const EdgeInsets.fromLTRB(16, 12, 8, 12),
        child: Row(
          children: [
            Expanded(
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
                        style: textTheme.titleSmall
                            ?.copyWith(color: colorScheme.primary),
                      ),
                    ],
                  ),
                  const SizedBox(height: 6),
                  Wrap(
                    spacing: 10,
                    runSpacing: 4,
                    children: [
                      _Chip(
                        icon: listingCategoryIcon(listing.category),
                        label: listingCategoryLabel(l10n, listing.category),
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
            if (busy)
              const Padding(
                padding: EdgeInsets.symmetric(horizontal: 8),
                child: SizedBox(
                  width: 20,
                  height: 20,
                  child: CircularProgressIndicator(strokeWidth: 2),
                ),
              )
            else
              PopupMenuButton<_CardAction>(
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
          ],
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
        Text(label, style: Theme.of(context).textTheme.bodySmall?.copyWith(color: c)),
      ],
    );
  }
}
