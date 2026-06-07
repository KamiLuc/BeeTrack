import 'dart:math';

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';

import '../../../core/api/api_client.dart';
import '../../../core/storage/token_storage.dart';
import '../../../core/theme/app_layout.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../l10n/app_localizations.dart';
import '../../hive/data/hive_model.dart';
import '../cubit/treatments_cubit.dart';
import '../data/treatment_model.dart';
import '../data/treatment_repository.dart';
import 'treatment_form_screen.dart';

class TreatmentHistoryScreen extends StatelessWidget {
  final int apiaryId;
  final Hive hive;

  const TreatmentHistoryScreen({
    super.key,
    required this.apiaryId,
    required this.hive,
  });

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => TreatmentsCubit(
        repo: TreatmentRepository(api: context.read<ApiClient>()),
        apiaryId: apiaryId,
        hiveId: hive.id,
      )..load(),
      child: _TreatmentHistoryView(apiaryId: apiaryId, hive: hive),
    );
  }
}

class _TreatmentHistoryView extends StatelessWidget {
  final int apiaryId;
  final Hive hive;

  const _TreatmentHistoryView({
    required this.apiaryId,
    required this.hive,
  });

  Future<void> _openCreate(BuildContext context) async {
    final result = await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (_) => TreatmentFormScreen(
          apiaryId: apiaryId,
          hive: hive,
        ),
      ),
    );
    if ((result ?? false) && context.mounted) {
      context.read<TreatmentsCubit>().load();
    }
  }

  Future<void> _openEdit(BuildContext context, Treatment treatment) async {
    final result = await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (_) => TreatmentFormScreen(
          apiaryId: apiaryId,
          hive: hive,
          treatment: treatment,
        ),
      ),
    );
    if ((result ?? false) && context.mounted) {
      context.read<TreatmentsCubit>().load();
    }
  }

  Future<void> _confirmDelete(
    BuildContext context,
    Treatment treatment,
  ) async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(l10n.treatmentDeleteConfirm),
        content: Text(l10n.treatmentDeleteWarning),
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
    if ((confirmed ?? false) && context.mounted) {
      context.read<TreatmentsCubit>().delete(treatment.id);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final currentUserName = context.read<TokenStorage>().name;
    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.treatmentTitle),
        actions: const [ProfileIconButton()],
      ),
      body: BlocBuilder<TreatmentsCubit, TreatmentsState>(
        builder: (context, state) {
          if (state is TreatmentsLoading || state is TreatmentsInitial) {
            return const Center(child: CircularProgressIndicator());
          }
          if (state is TreatmentsError) {
            return Center(child: Text(l10n.generalError));
          }
          final loaded = state as TreatmentsLoaded;
          final treatments = loaded.treatments;
          if (treatments.isEmpty) {
            return Column(
              children: [
                Expanded(
                  child: Center(child: Text(l10n.treatmentEmpty)),
                ),
                _TreatmentBanner(
                  loaded: loaded,
                  onAdd: () => _openCreate(context),
                  onPage: (p) => context.read<TreatmentsCubit>().goToPage(p),
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
                          for (final t in treatments) ...[
                            TreatmentCard(
                              treatment: t,
                              currentUserName: currentUserName,
                              onEdit: () => _openEdit(context, t),
                              onDelete: () => _confirmDelete(context, t),
                            ),
                            const SizedBox(height: 8),
                          ],
                        ],
                      ),
                    ),
                  ),
                ),
              ),
              _TreatmentBanner(
                loaded: loaded,
                onAdd: () => _openCreate(context),
                onPage: (p) => context.read<TreatmentsCubit>().goToPage(p),
              ),
            ],
          );
        },
      ),
    );
  }
}

class TreatmentCard extends StatelessWidget {
  final Treatment treatment;
  final String? currentUserName;
  final VoidCallback onEdit;
  final VoidCallback onDelete;

  const TreatmentCard({
    super.key,
    required this.treatment,
    this.currentUserName,
    required this.onEdit,
    required this.onDelete,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;
    final bodyStyle = textTheme.bodyMedium?.copyWith(
      color: colorScheme.onSurfaceVariant,
    );
    final labelStyle = textTheme.labelSmall?.copyWith(
      color: colorScheme.primary,
      fontWeight: FontWeight.w600,
      letterSpacing: 0.4,
    );
    final dateStr = DateFormat.yMMMd(
      Localizations.localeOf(context).toString(),
    ).add_Hm().format(treatment.treatedAt);

    final doseCount = int.tryParse(treatment.dose);
    final doseDisplay = doseCount != null
        ? l10n.treatmentDoseCount(doseCount)
        : treatment.dose;

    final otherTreater =
        treatment.treatedByName != null &&
                treatment.treatedByName != currentUserName
            ? treatment.treatedByName
            : null;

    return Card(
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
                  if (otherTreater != null) ...[
                    const SizedBox(height: 2),
                    Text(
                      l10n.treatmentTreatedBy(otherTreater),
                      style: bodyStyle,
                    ),
                  ],
                  const SizedBox(height: 6),
                  Text(l10n.treatmentMedicine, style: labelStyle),
                  const SizedBox(height: 2),
                  Text(
                    '${treatment.medicineName} · $doseDisplay',
                    style: bodyStyle,
                  ),
                  if (treatment.notes.isNotEmpty) ...[
                    const SizedBox(height: 8),
                    Text(l10n.treatmentNote, style: labelStyle),
                    const SizedBox(height: 2),
                    Text(
                      treatment.notes,
                      style: bodyStyle,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ],
                ],
              ),
            ),
            PopupMenuButton<_CardAction>(
              onSelected: (action) {
                if (action == _CardAction.edit) onEdit();
                if (action == _CardAction.delete) onDelete();
              },
              itemBuilder: (_) => [
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
      ),
    );
  }
}

enum _CardAction { edit, delete }

class _TreatmentBanner extends StatelessWidget {
  final TreatmentsLoaded loaded;
  final VoidCallback onAdd;
  final ValueChanged<int> onPage;

  const _TreatmentBanner({
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
                    tooltip: l10n.treatmentAdd,
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
