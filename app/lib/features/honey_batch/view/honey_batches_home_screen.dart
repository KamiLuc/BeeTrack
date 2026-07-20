import 'dart:math';

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';

import '../../../core/api/api_client.dart';
import '../../../core/theme/app_layout.dart';
import '../../../core/widgets/app_drawer.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../l10n/app_localizations.dart';
import '../cubit/honey_batches_cubit.dart';
import '../data/honey_batch_model.dart';
import '../data/honey_batch_repository.dart';
import '../data/processing_method.dart';
import 'certification_status_badge.dart';
import 'create_honey_batch_screen.dart';
import 'honey_batch_detail_screen.dart';

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

  Future<void> _openDetail(BuildContext context, HoneyBatchModel batch) async {
    final result = await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (_) => HoneyBatchDetailScreen(batchId: batch.id),
      ),
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
                              batch: batch,
                              onTap: () => _openDetail(context, batch),
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

class HoneyBatchCard extends StatelessWidget {
  final HoneyBatchModel batch;
  final VoidCallback onTap;

  const HoneyBatchCard({super.key, required this.batch, required this.onTap});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;
    final bodyStyle = textTheme.bodyMedium?.copyWith(
      color: colorScheme.onSurfaceVariant,
    );
    final dateStr = DateFormat.yMMMd(
      Localizations.localeOf(context).toString(),
    ).format(batch.gatheringDate);

    return Card(
      margin: EdgeInsets.zero,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          child: Row(
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      dateStr,
                      style: textTheme.bodyLarge?.copyWith(
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      '${batch.honeyType} · ${processingMethodLabel(l10n, batch.processingMethod)}',
                      style: bodyStyle,
                    ),
                    const SizedBox(height: 4),
                    Text('${batch.amountKg.toStringAsFixed(1)} kg', style: bodyStyle),
                  ],
                ),
              ),
              const SizedBox(width: 8),
              CertificationStatusBadge(status: batch.certification?.status),
            ],
          ),
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
