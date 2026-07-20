import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';

import '../../../core/api/api_client.dart';
import '../../../core/theme/app_layout.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../l10n/app_localizations.dart';
import '../data/honey_batch_certification_model.dart';
import '../data/honey_batch_model.dart';
import '../data/honey_batch_repository.dart';
import '../data/processing_method.dart';
import 'certification_status_badge.dart';

class HoneyBatchDetailScreen extends StatefulWidget {
  final int batchId;

  const HoneyBatchDetailScreen({super.key, required this.batchId});

  @override
  State<HoneyBatchDetailScreen> createState() => _HoneyBatchDetailScreenState();
}

class _HoneyBatchDetailScreenState extends State<HoneyBatchDetailScreen> {
  late Future<HoneyBatchModel> _future;
  bool _certifying = false;
  bool _deleting = false;

  @override
  void initState() {
    super.initState();
    _future = _load();
  }

  Future<HoneyBatchModel> _load() {
    return HoneyBatchRepository(api: context.read<ApiClient>())
        .getBatch(widget.batchId);
  }

  Future<void> _certify() async {
    final l10n = AppLocalizations.of(context)!;
    setState(() => _certifying = true);
    try {
      final batch = await HoneyBatchRepository(api: context.read<ApiClient>())
          .requestCertification(widget.batchId);
      if (!mounted) return;
      setState(() {
        _future = Future.value(batch);
        _certifying = false;
      });
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.generalError)),
        );
        setState(() => _certifying = false);
      }
    }
  }

  Future<void> _confirmDelete() async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(l10n.honeyBatchDeleteConfirm),
        content: Text(l10n.honeyBatchDeleteWarning),
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
    if ((confirmed ?? false) && mounted) _delete();
  }

  Future<void> _delete() async {
    final l10n = AppLocalizations.of(context)!;
    setState(() => _deleting = true);
    try {
      await HoneyBatchRepository(api: context.read<ApiClient>())
          .deleteBatch(widget.batchId);
      if (!mounted) return;
      Navigator.of(context).pop(true);
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.generalError)),
        );
        setState(() => _deleting = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.honeyBatchDetailTitle),
        actions: [
          _deleting
              ? const Padding(
                  padding: EdgeInsets.all(16),
                  child: SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  ),
                )
              : IconButton(
                  icon: const Icon(Icons.delete_outline),
                  tooltip: l10n.generalDelete,
                  onPressed: _confirmDelete,
                ),
          const ProfileIconButton(),
        ],
      ),
      body: FutureBuilder<HoneyBatchModel>(
        future: _future,
        builder: (context, snapshot) {
          if (snapshot.connectionState != ConnectionState.done) {
            return const Center(child: CircularProgressIndicator());
          }
          if (snapshot.hasError) {
            return Center(child: Text(l10n.generalError));
          }
          final batch = snapshot.data!;
          return Center(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(24),
              child: ConstrainedBox(
                constraints: AppLayout.formConstraints(context),
                child: _HoneyBatchDetail(
                  batch: batch,
                  certifying: _certifying,
                  onCertify: _certify,
                ),
              ),
            ),
          );
        },
      ),
    );
  }
}

class _HoneyBatchDetail extends StatelessWidget {
  final HoneyBatchModel batch;
  final bool certifying;
  final VoidCallback onCertify;

  const _HoneyBatchDetail({
    required this.batch,
    required this.certifying,
    required this.onCertify,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final textTheme = Theme.of(context).textTheme;
    final dateStr = DateFormat.yMMMd(
      Localizations.localeOf(context).toString(),
    ).format(batch.gatheringDate);
    final cert = batch.certification;
    final truncatedHash = batch.pdfFileHash.isEmpty
        ? l10n.honeyBatchNoPdf
        : batch.pdfFileHash.length > 16
            ? '${batch.pdfFileHash.substring(0, 16)}…'
            : batch.pdfFileHash;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(dateStr, style: textTheme.headlineSmall),
            CertificationStatusBadge(status: cert?.status),
          ],
        ),
        const SizedBox(height: 16),
        _DetailRow(label: l10n.honeyBatchHoneyType, value: batch.honeyType),
        _DetailRow(
          label: l10n.honeyBatchProcessingMethod,
          value: processingMethodLabel(l10n, batch.processingMethod),
        ),
        _DetailRow(
          label: l10n.honeyBatchAmountKg,
          value: batch.amountKg.toStringAsFixed(1),
        ),
        _DetailRow(
          label: l10n.honeyBatchPdfLabel,
          value: truncatedHash,
          valueStyle: const TextStyle(fontFamily: 'monospace'),
        ),
        const SizedBox(height: 24),
        _CertificationAction(
          cert: cert,
          certifying: certifying,
          onCertify: onCertify,
        ),
      ],
    );
  }
}

class _DetailRow extends StatelessWidget {
  final String label;
  final String value;
  final TextStyle? valueStyle;

  const _DetailRow({required this.label, required this.value, this.valueStyle});

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            label,
            style: textTheme.labelSmall?.copyWith(
              color: colorScheme.primary,
              fontWeight: FontWeight.w600,
              letterSpacing: 0.4,
            ),
          ),
          const SizedBox(height: 2),
          Text(value, style: valueStyle ?? textTheme.bodyMedium),
        ],
      ),
    );
  }
}

class _CertificationAction extends StatelessWidget {
  final HoneyBatchCertificationModel? cert;
  final bool certifying;
  final VoidCallback onCertify;

  const _CertificationAction({
    required this.cert,
    required this.certifying,
    required this.onCertify,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    if (certifying) {
      return const Center(
        child: SizedBox(
          width: 24,
          height: 24,
          child: CircularProgressIndicator(strokeWidth: 2),
        ),
      );
    }

    if (cert == null) {
      return FilledButton(
        onPressed: onCertify,
        child: Text(l10n.honeyBatchCertify),
      );
    }

    final isRetryable = cert!.status == CertificationStatus.failed ||
        cert!.status == CertificationStatus.reverted;
    if (isRetryable) {
      return FilledButton(
        onPressed: onCertify,
        child: Text(l10n.honeyBatchRetry),
      );
    }

    if (cert!.status == CertificationStatus.confirmed) {
      return Row(
        children: [
          Expanded(
            child: OutlinedButton(
              // TODO(HC-FE-05): navigate to QR display screen
              onPressed: () {},
              child: Text(l10n.honeyBatchViewQr),
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: OutlinedButton(
              // TODO(HC-FE-05): download QR code
              onPressed: () {},
              child: Text(l10n.honeyBatchDownloadQr),
            ),
          ),
        ],
      );
    }

    return Row(
      children: [
        const SizedBox(
          width: 18,
          height: 18,
          child: CircularProgressIndicator(strokeWidth: 2),
        ),
        const SizedBox(width: 12),
        Text(l10n.honeyBatchInProgress),
      ],
    );
  }
}
