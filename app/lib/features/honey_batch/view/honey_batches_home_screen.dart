import 'dart:math';

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';

import '../../../core/api/api_client.dart';
import '../../../core/theme/app_layout.dart';
import '../../../core/widgets/app_drawer.dart';
import '../../../core/widgets/delete_dialog.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../l10n/app_localizations.dart';
import '../cubit/honey_batches_cubit.dart';
import '../data/honey_batch_certification_model.dart';
import '../data/honey_batch_model.dart';
import '../data/honey_batch_repository.dart';
import '../data/processing_method.dart';
import 'certification_status_badge.dart';
import 'create_honey_batch_screen.dart';

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
                            HoneyBatchCard(batch: batch),
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

enum _CardAction { edit, delete }

class HoneyBatchCard extends StatefulWidget {
  final HoneyBatchModel batch;

  const HoneyBatchCard({super.key, required this.batch});

  @override
  State<HoneyBatchCard> createState() => _HoneyBatchCardState();
}

class _HoneyBatchCardState extends State<HoneyBatchCard> {
  bool _certifying = false;
  bool _deleting = false;

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
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.generalError)),
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
    final pdfDisplay =
        batch.pdfFilename.isEmpty ? l10n.honeyBatchNoPdf : batch.pdfFilename;
    final canEdit = batch.certification == null;

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
                CertificationStatusBadge(status: batch.certification?.status),
                _deleting
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
                          if (action == _CardAction.delete) _confirmDelete();
                        },
                        itemBuilder: (_) => [
                          if (canEdit)
                            PopupMenuItem(
                              value: _CardAction.edit,
                              child: Text(l10n.generalEdit),
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
            Text(pdfDisplay, style: bodyStyle, overflow: TextOverflow.ellipsis),
            const SizedBox(height: 12),
            _CertificationAction(
              cert: batch.certification,
              certifying: _certifying,
              onCertify: _certify,
            ),
          ],
        ),
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
      return const SizedBox(
        height: 36,
        child: Center(
          child: SizedBox(
            width: 20,
            height: 20,
            child: CircularProgressIndicator(strokeWidth: 2),
          ),
        ),
      );
    }

    if (cert == null) {
      return Align(
        alignment: Alignment.center,
        child: FractionallySizedBox(
          widthFactor: 0.4,
          child: FilledButton(
            onPressed: onCertify,
            child: Text(l10n.honeyBatchCertify),
          ),
        ),
      );
    }

    final isRetryable = cert!.status == CertificationStatus.failed ||
        cert!.status == CertificationStatus.reverted;
    if (isRetryable) {
      return Align(
        alignment: Alignment.center,
        child: FractionallySizedBox(
          widthFactor: 0.4,
          child: FilledButton(
            onPressed: onCertify,
            child: Text(l10n.honeyBatchRetry),
          ),
        ),
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
          width: 16,
          height: 16,
          child: CircularProgressIndicator(strokeWidth: 2),
        ),
        const SizedBox(width: 10),
        Text(l10n.honeyBatchInProgress),
      ],
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
