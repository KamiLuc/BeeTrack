import 'dart:math';

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';

import '../../../core/api/api_client.dart';
import '../../../core/storage/token_storage.dart';
import '../../../core/theme/app_layout.dart';
import '../../../core/widgets/note_dialog.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../l10n/app_localizations.dart';
import '../../hive/data/hive_model.dart';
import '../cubit/feedings_cubit.dart';
import '../data/feeding_model.dart';
import '../data/feeding_repository.dart';
import 'feeding_form_screen.dart';

class FeedingHistoryScreen extends StatelessWidget {
  final int apiaryId;
  final Hive hive;

  const FeedingHistoryScreen({
    super.key,
    required this.apiaryId,
    required this.hive,
  });

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => FeedingsCubit(
        repo: FeedingRepository(api: context.read<ApiClient>()),
        apiaryId: apiaryId,
        hiveId: hive.id,
      )..load(),
      child: _FeedingHistoryView(apiaryId: apiaryId, hive: hive),
    );
  }
}

class _FeedingHistoryView extends StatelessWidget {
  final int apiaryId;
  final Hive hive;

  const _FeedingHistoryView({
    required this.apiaryId,
    required this.hive,
  });

  Future<void> _openCreate(BuildContext context) async {
    final result = await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (_) => FeedingFormScreen(
          apiaryId: apiaryId,
          hive: hive,
        ),
      ),
    );
    if ((result ?? false) && context.mounted) {
      context.read<FeedingsCubit>().load();
    }
  }

  Future<void> _openEdit(BuildContext context, Feeding feeding) async {
    final result = await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (_) => FeedingFormScreen(
          apiaryId: apiaryId,
          hive: hive,
          feeding: feeding,
        ),
      ),
    );
    if ((result ?? false) && context.mounted) {
      context.read<FeedingsCubit>().load();
    }
  }

  Future<void> _confirmDelete(
    BuildContext context,
    Feeding feeding,
  ) async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(l10n.feedingDeleteConfirm),
        content: Text(l10n.feedingDeleteWarning),
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
      context.read<FeedingsCubit>().delete(feeding.id);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final currentUserName = context.read<TokenStorage>().name;
    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.feedingTitle),
        actions: const [ProfileIconButton()],
      ),
      body: BlocBuilder<FeedingsCubit, FeedingsState>(
        builder: (context, state) {
          if (state is FeedingsLoading || state is FeedingsInitial) {
            return const Center(child: CircularProgressIndicator());
          }
          if (state is FeedingsError) {
            return Center(child: Text(l10n.generalError));
          }
          final loaded = state as FeedingsLoaded;
          final feedings = loaded.feedings;
          if (feedings.isEmpty) {
            return Column(
              children: [
                Expanded(
                  child: Center(child: Text(l10n.feedingEmpty)),
                ),
                _FeedingBanner(
                  loaded: loaded,
                  onAdd: () => _openCreate(context),
                  onPage: (p) => context.read<FeedingsCubit>().goToPage(p),
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
                          for (final f in feedings) ...[
                            FeedingCard(
                              feeding: f,
                              currentUserName: currentUserName,
                              onEdit: () => _openEdit(context, f),
                              onDelete: () => _confirmDelete(context, f),
                            ),
                            const SizedBox(height: 8),
                          ],
                        ],
                      ),
                    ),
                  ),
                ),
              ),
              _FeedingBanner(
                loaded: loaded,
                onAdd: () => _openCreate(context),
                onPage: (p) => context.read<FeedingsCubit>().goToPage(p),
              ),
            ],
          );
        },
      ),
    );
  }
}

class FeedingCard extends StatelessWidget {
  final Feeding feeding;
  final String? currentUserName;
  final VoidCallback onEdit;
  final VoidCallback onDelete;

  const FeedingCard({
    super.key,
    required this.feeding,
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
    ).add_Hm().format(feeding.fedAt);

    final otherFeeder =
        feeding.fedByName != null && feeding.fedByName != currentUserName
            ? feeding.fedByName
            : null;

    return LayoutBuilder(
      builder: (context, constraints) {
        // 16*2 padding + ~48 for the popup menu button.
        final textWidth = constraints.maxWidth - 80;
        final truncated = feeding.notes.isNotEmpty &&
            isTextTruncated(feeding.notes, bodyStyle, textWidth);

        return Card(
          child: InkWell(
            onTap: truncated
                ? () => showNoteDialog(
                      context,
                      title: l10n.feedingNote,
                      note: feeding.notes,
                    )
                : null,
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
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
                        if (otherFeeder != null) ...[
                          const SizedBox(height: 2),
                          Text(
                            l10n.feedingFedBy(otherFeeder),
                            style: bodyStyle,
                          ),
                        ],
                        const SizedBox(height: 6),
                        Text(l10n.feedingType, style: labelStyle),
                        const SizedBox(height: 2),
                        Text(
                          '${feeding.feedType} · ${feeding.amount}',
                          style: bodyStyle,
                        ),
                        if (feeding.notes.isNotEmpty) ...[
                          const SizedBox(height: 8),
                          Text(l10n.feedingNote, style: labelStyle),
                          const SizedBox(height: 2),
                          Text(
                            feeding.notes,
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
          ),
        );
      },
    );
  }
}

enum _CardAction { edit, delete }

class _FeedingBanner extends StatelessWidget {
  final FeedingsLoaded loaded;
  final VoidCallback onAdd;
  final ValueChanged<int> onPage;

  const _FeedingBanner({
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
                    tooltip: l10n.feedingAdd,
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
