import 'package:flutter/material.dart';

import '../../../l10n/app_localizations.dart';
import '../data/honey_batch_certification_model.dart';

class CertificationStatusBadge extends StatelessWidget {
  final CertificationStatus? status;

  const CertificationStatusBadge({super.key, required this.status});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;

    if (status == null) {
      return _badge(
        context,
        label: l10n.honeyBatchNotCertified,
        bg: colorScheme.surfaceContainerHighest,
        fg: colorScheme.onSurfaceVariant,
      );
    }

    if (!status!.isTerminal) {
      return _badge(
        context,
        label: l10n.honeyBatchInProgress,
        bg: Colors.amber.withValues(alpha: 0.15),
        fg: Colors.amber,
        showSpinner: true,
      );
    }

    final (label, color) = switch (status!) {
      CertificationStatus.confirmed => (l10n.honeyBatchStatusConfirmed, Colors.green),
      CertificationStatus.failed => (l10n.honeyBatchStatusFailed, colorScheme.error),
      CertificationStatus.reverted => (l10n.honeyBatchStatusReverted, colorScheme.error),
      _ => throw StateError('unreachable: $status is not terminal'),
    };

    return _badge(context, label: label, bg: color.withValues(alpha: 0.15), fg: color);
  }

  Widget _badge(
    BuildContext context, {
    required String label,
    required Color bg,
    required Color fg,
    bool showSpinner = false,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(color: bg, borderRadius: BorderRadius.circular(20)),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (showSpinner) ...[
            SizedBox(
              width: 10,
              height: 10,
              child: CircularProgressIndicator(strokeWidth: 1.5, color: fg),
            ),
            const SizedBox(width: 6),
          ],
          Text(
            label,
            style: Theme.of(context)
                .textTheme
                .labelSmall
                ?.copyWith(color: fg, fontWeight: FontWeight.w600),
          ),
        ],
      ),
    );
  }
}
