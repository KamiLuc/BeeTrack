import 'dart:math';

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';

import '../../../core/api/api_client.dart';
import '../../../core/api/api_exception.dart';
import '../../../core/theme/app_layout.dart';
import '../../../core/widgets/app_drawer.dart';
import '../../../core/widgets/delete_dialog.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../l10n/app_localizations.dart';
import '../cubit/honey_batches_cubit.dart';
import '../data/honey_batch_certification_model.dart';
import '../data/honey_batch_certification_request_model.dart';
import '../data/honey_batch_model.dart';
import '../data/honey_batch_repository.dart';
import '../data/processing_method.dart';
import 'certification_request_status_badge.dart';
import 'certification_status_badge.dart';
import 'create_honey_batch_screen.dart';
import 'pdf_preview_screen.dart';
import 'qr_preview_screen.dart';

class HoneyBatchesHomeScreen extends StatelessWidget {
  final ValueChanged<AppSection> onSelectSection;

  const HoneyBatchesHomeScreen({super.key, required this.onSelectSection});

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => HoneyBatchesCubit(
        repo: HoneyBatchRepository(api: context.read<ApiClient>()),
      )..load(),
      child: _HoneyBatchesView(onSelectSection: onSelectSection),
    );
  }
}

class _HoneyBatchesView extends StatelessWidget {
  final ValueChanged<AppSection> onSelectSection;

  const _HoneyBatchesView({required this.onSelectSection});

  Future<void> _openCreate(BuildContext context) async {
    final result = await Navigator.of(context).push<bool>(
      MaterialPageRoute(builder: (_) => const CreateHoneyBatchScreen()),
    );
    if ((result ?? false) && context.mounted) {
      context.read<HoneyBatchesCubit>().load();
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.honeyBatchTitle),
        actions: const [ProfileIconButton()],
      ),
      drawer: AuthenticatedAppDrawer(
        current: AppSection.honeyBatches,
        onSelect: onSelectSection,
      ),
      body: BlocBuilder<HoneyBatchesCubit, HoneyBatchesState>(
        builder: (context, state) {
          if (state is HoneyBatchesLoading || state is HoneyBatchesInitial) {
            return const Center(child: CircularProgressIndicator());
          }
          if (state is HoneyBatchesError) {
            return Center(child: Text(l10n.generalError));
          }
          final loaded = state as HoneyBatchesLoaded;
          final batches = loaded.batches;
          if (batches.isEmpty) {
            return Column(
              children: [
                Expanded(
                  child: Center(child: Text(l10n.honeyBatchEmpty)),
                ),
                _HoneyBatchesBanner(
                  loaded: loaded,
                  onAdd: () => _openCreate(context),
                  onPage: (p) => context.read<HoneyBatchesCubit>().goToPage(p),
                ),
              ],
            );
          }
          return Column(
            children: [
              Expanded(
                child: Center(
                  child: SingleChildScrollView(
                    padding: const EdgeInsets.all(16),
                    child: ConstrainedBox(
                      constraints: AppLayout.formConstraints(context),
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          for (final batch in batches) ...[
                            HoneyBatchCard(
                              key: ValueKey(batch.id),
                              batch: batch,
                            ),
                            const SizedBox(height: 8),
                          ],
                        ],
                      ),
                    ),
                  ),
                ),
              ),
              _HoneyBatchesBanner(
                loaded: loaded,
                onAdd: () => _openCreate(context),
                onPage: (p) => context.read<HoneyBatchesCubit>().goToPage(p),
              ),
            ],
          );
        },
      ),
    );
  }
}

enum _CardAction { edit, certify, openPublicPage, delete }

class HoneyBatchCard extends StatefulWidget {
  final HoneyBatchModel batch;

  const HoneyBatchCard({super.key, required this.batch});

  @override
  State<HoneyBatchCard> createState() => _HoneyBatchCardState();
}

class _HoneyBatchCardState extends State<HoneyBatchCard> {
  bool _certifying = false;
  bool _deleting = false;
  bool _loadingPdf = false;

  Future<void> _viewPdf() async {
    if (widget.batch.pdfFilename.isEmpty) return;
    final l10n = AppLocalizations.of(context)!;
    setState(() => _loadingPdf = true);
    try {
      final repo = HoneyBatchRepository(api: context.read<ApiClient>());
      final bytes = await repo.getPdfBytes(widget.batch.id);
      if (!mounted) return;
      await showPdfPreviewDialog(
        context,
        title: widget.batch.pdfFilename,
        bytes: bytes,
      );
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.generalError)),
        );
      }
    } finally {
      if (mounted) setState(() => _loadingPdf = false);
    }
  }

  void _viewQr() {
    final repo = HoneyBatchRepository(api: context.read<ApiClient>());
    showQrPreviewDialog(
      context,
      title: AppLocalizations.of(context)!.honeyBatchViewQr,
      imageUrl: repo.qrCodeImageUrl(widget.batch.verificationToken),
      downloadUrl: repo.qrCodeDownloadUrl(widget.batch.verificationToken),
      verificationUrl: widget.batch.verificationUrl,
    );
  }

  void _downloadQr() {
    final repo = HoneyBatchRepository(api: context.read<ApiClient>());
    launchQrDownload(repo.qrCodeDownloadUrl(widget.batch.verificationToken));
  }

  Future<void> _certify() async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await showDeleteDialog(
      context,
      title: l10n.honeyBatchCertifyConfirmTitle,
      warning: l10n.honeyBatchCertifyConfirmMessage,
      l10n: l10n,
      withPuzzle: true,
      confirmLabel: l10n.generalConfirm,
    );
    if (!confirmed || !mounted) return;

    setState(() => _certifying = true);
    try {
      await context.read<HoneyBatchesCubit>().requestCertification(widget.batch.id);
    } catch (e) {
      if (mounted) {
        final message = switch (e is ApiException ? e.code : null) {
          'CERTIFICATION_REQUEST_PENDING' => l10n.honeyBatchCertificationRequestPending,
          'BATCH_ALREADY_CERTIFIED' => l10n.honeyBatchAlreadyCertifiedError,
          _ => l10n.generalError,
        };
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(message)),
        );
      }
    } finally {
      if (mounted) setState(() => _certifying = false);
    }
  }

  Future<void> _confirmDelete() async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await showDeleteDialog(
      context,
      title: l10n.honeyBatchDeleteConfirm,
      warning: l10n.honeyBatchDeleteWarning,
      l10n: l10n,
      withPuzzle: true,
    );
    if (!confirmed || !mounted) return;

    setState(() => _deleting = true);
    try {
      await context.read<HoneyBatchesCubit>().delete(widget.batch.id);
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.generalError)),
        );
        setState(() => _deleting = false);
      }
    }
  }

  Future<void> _edit() async {
    final cubit = context.read<HoneyBatchesCubit>();
    final updated = await Navigator.of(context).push<HoneyBatchModel>(
      MaterialPageRoute(
        builder: (_) => CreateHoneyBatchScreen(existingBatch: widget.batch),
      ),
    );
    if (updated != null) cubit.replaceInList(updated);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;
    final labelStyle = textTheme.labelSmall?.copyWith(
      color: colorScheme.primary,
      fontWeight: FontWeight.w600,
      letterSpacing: 0.4,
    );
    final bodyStyle = textTheme.bodyMedium?.copyWith(
      color: colorScheme.onSurfaceVariant,
    );
    final batch = widget.batch;
    final dateStr = DateFormat.yMMMd(
      Localizations.localeOf(context).toString(),
    ).format(batch.gatheringDate);
    final hasPdf = batch.pdfFilename.isNotEmpty;
    final pdfDisplay = hasPdf ? batch.pdfFilename : l10n.honeyBatchNoPdf;
    final canEdit = batch.certification == null;
    final certRequest = batch.certificationRequest;
    final showRequestBadge = batch.certification == null &&
        certRequest != null &&
        (certRequest.status == CertificationRequestStatus.pending ||
            certRequest.status == CertificationRequestStatus.rejected);
    final isRetryable = batch.certification?.status == CertificationStatus.failed ||
        batch.certification?.status == CertificationStatus.reverted;
    final showCertifyAction = isRetryable ||
        (batch.certification == null &&
            hasPdf &&
            certRequest?.status != CertificationRequestStatus.pending);
    final isConfirmed = batch.certification?.status == CertificationStatus.confirmed;

    return Card(
      margin: EdgeInsets.zero,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Expanded(
                  child: Text(
                    dateStr,
                    style: textTheme.bodyLarge?.copyWith(fontWeight: FontWeight.w600),
                  ),
                ),
                showRequestBadge
                    ? CertificationRequestStatusBadge(
                        status: certRequest.status,
                        rejectionReason: certRequest.rejectionReason,
                      )
                    : CertificationStatusBadge(status: batch.certification?.status),
                _deleting || _certifying
                    ? const Padding(
                        padding: EdgeInsets.all(8),
                        child: SizedBox(
                          width: 18,
                          height: 18,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        ),
                      )
                    : PopupMenuButton<_CardAction>(
                        onSelected: (action) {
                          if (action == _CardAction.edit) _edit();
                          if (action == _CardAction.certify) _certify();
                          if (action == _CardAction.openPublicPage) {
                            launchVerificationPage(batch.verificationUrl);
                          }
                          if (action == _CardAction.delete) _confirmDelete();
                        },
                        itemBuilder: (_) => [
                          if (canEdit)
                            PopupMenuItem(
                              value: _CardAction.edit,
                              child: Text(l10n.generalEdit),
                            ),
                          if (showCertifyAction)
                            PopupMenuItem(
                              value: _CardAction.certify,
                              child: Text(
                                isRetryable ? l10n.honeyBatchRetry : l10n.honeyBatchCertify,
                              ),
                            ),
                          if (isConfirmed)
                            PopupMenuItem(
                              value: _CardAction.openPublicPage,
                              child: Text(l10n.honeyBatchOpenPublicPage),
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
            const SizedBox(height: 4),
            Text(
              '${batch.honeyType} · ${processingMethodLabel(l10n, batch.processingMethod)}',
              style: bodyStyle,
            ),
            const SizedBox(height: 2),
            Text('${batch.amountKg.toStringAsFixed(1)} kg', style: bodyStyle),
            const SizedBox(height: 6),
            Text(l10n.honeyBatchPdfLabel, style: labelStyle),
            const SizedBox(height: 2),
            hasPdf
                ? InkWell(
                    onTap: _loadingPdf ? null : _viewPdf,
                    child: Row(
                      children: [
                        Flexible(
                          child: Text(
                            pdfDisplay,
                            style: bodyStyle?.copyWith(color: colorScheme.primary),
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                        const SizedBox(width: 6),
                        if (_loadingPdf)
                          const SizedBox(
                            width: 14,
                            height: 14,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        else
                          Icon(
                            Icons.visibility_outlined,
                            size: 16,
                            color: colorScheme.primary,
                          ),
                      ],
                    ),
                  )
                : Text(pdfDisplay, style: bodyStyle, overflow: TextOverflow.ellipsis),
            if (isConfirmed) ...[
              const SizedBox(height: 12),
              Row(
                children: [
                  Expanded(
                    child: OutlinedButton(
                      onPressed: _viewQr,
                      child: Text(l10n.honeyBatchViewQr),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: OutlinedButton(
                      onPressed: _downloadQr,
                      child: Text(l10n.honeyBatchDownloadQr),
                    ),
                  ),
                ],
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _HoneyBatchesBanner extends StatelessWidget {
  final HoneyBatchesLoaded loaded;
  final VoidCallback onAdd;
  final ValueChanged<int> onPage;

  const _HoneyBatchesBanner({
    required this.loaded,
    required this.onAdd,
    required this.onPage,
  });

  List<_PageItem> _buildPageItems() {
    final cur = loaded.currentPage;
    final last = loaded.totalPages;
    if (last <= 1) return [];

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
    final l10n = AppLocalizations.of(context)!;
    final bannerWidth = AppLayout.bannerWidth(context);
    final cur = loaded.currentPage;
    final last = loaded.totalPages;
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
                mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                children: [
                  IconButton(
                    icon: const Icon(Icons.add),
                    iconSize: 28,
                    tooltip: l10n.honeyBatchAdd,
                    onPressed: onAdd,
                  ),
                  if (last > 1) ...[
                    IconButton(
                      icon: const Icon(Icons.chevron_left),
                      iconSize: 24,
                      onPressed: cur > 1 ? () => onPage(cur - 1) : null,
                    ),
                    for (final item in pageItems)
                      item.isEllipsis
                          ? const Padding(
                              padding: EdgeInsets.symmetric(horizontal: 2),
                              child: Text('…', style: TextStyle(fontSize: 16)),
                            )
                          : _PageButton(
                              page: item.page!,
                              isCurrent: item.page == cur,
                              onTap: () => onPage(item.page!),
                            ),
                    IconButton(
                      icon: const Icon(Icons.chevron_right),
                      iconSize: 24,
                      onPressed: cur < last ? () => onPage(cur + 1) : null,
                    ),
                  ],
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
  const _PageItem.ellipsis()
      : page = null,
        isEllipsis = true;
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
