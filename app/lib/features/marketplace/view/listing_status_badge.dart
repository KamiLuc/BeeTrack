import 'package:flutter/material.dart';

import '../../../l10n/app_localizations.dart';
import '../data/listing_model.dart';

class ListingStatusBadge extends StatelessWidget {
  final ListingStatus status;
  final String? rejectionReason;

  const ListingStatusBadge({
    super.key,
    required this.status,
    this.rejectionReason,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;

    final (label, color) = switch (status) {
      ListingStatus.pending => (l10n.marketplaceStatusPending, Colors.amber),
      ListingStatus.approved => (l10n.marketplaceStatusApproved, Colors.green),
      ListingStatus.rejected => (l10n.marketplaceStatusRejected, colorScheme.error),
    };

    final badge = _badge(context, label: label, bg: color.withValues(alpha: 0.15), fg: color);

    if (status == ListingStatus.rejected &&
        rejectionReason != null &&
        rejectionReason!.isNotEmpty) {
      return Tooltip(message: rejectionReason!, child: badge);
    }
    return badge;
  }

  Widget _badge(
    BuildContext context, {
    required String label,
    required Color bg,
    required Color fg,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(color: bg, borderRadius: BorderRadius.circular(20)),
      child: Text(
        label,
        style: Theme.of(context)
            .textTheme
            .labelSmall
            ?.copyWith(color: fg, fontWeight: FontWeight.w600, height: 1),
      ),
    );
  }
}
