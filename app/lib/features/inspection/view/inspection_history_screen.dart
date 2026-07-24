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
import '../../hive/data/hive_repository.dart';
import '../cubit/inspections_cubit.dart';
import '../data/inspection_image_repository.dart';
import '../data/inspection_model.dart';
import '../data/inspection_repository.dart';
import 'inspection_form_screen.dart';
import 'inspection_images_section.dart';
import 'inspection_summary.dart';

class InspectionHistoryScreen extends StatefulWidget {
  final int apiaryId;
  final Hive hive;

  const InspectionHistoryScreen({
    super.key,
    required this.apiaryId,
    required this.hive,
  });

  @override
  State<InspectionHistoryScreen> createState() =>
      _InspectionHistoryScreenState();
}

class _InspectionHistoryScreenState extends State<InspectionHistoryScreen> {
  late Hive _hive;

  @override
  void initState() {
    super.initState();
    _hive = widget.hive;
  }

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => InspectionsCubit(
        repo: InspectionRepository(api: context.read<ApiClient>()),
        apiaryId: widget.apiaryId,
        hiveId: _hive.id,
      )..load(),
      child: _InspectionHistoryView(
        apiaryId: widget.apiaryId,
        hive: _hive,
        onHiveChanged: (hive) => setState(() => _hive = hive),
      ),
    );
  }
}

class _InspectionHistoryView extends StatelessWidget {
  final int apiaryId;
  final Hive hive;
  final ValueChanged<Hive> onHiveChanged;

  const _InspectionHistoryView({
    required this.apiaryId,
    required this.hive,
    required this.onHiveChanged,
  });

  Future<void> _openCreate(BuildContext context) async {
    final state = context.read<InspectionsCubit>().state;
    final inspections =
        state is InspectionsLoaded ? state.inspections : <Inspection>[];
    final previous = inspections.isNotEmpty
        ? inspections.reduce(
            (a, b) => a.inspectedAt.isAfter(b.inspectedAt) ? a : b,
          )
        : null;
    final result = await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (_) => InspectionFormScreen(
          apiaryId: apiaryId,
          hive: hive,
          previousInspection: previous,
        ),
      ),
    );
    if ((result ?? false) && context.mounted) {
      context.read<InspectionsCubit>().load();
      await _refreshHive(context);
    }
  }

  Future<void> _openEdit(BuildContext context, Inspection inspection) async {
    final result = await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (_) => InspectionFormScreen(
          apiaryId: apiaryId,
          hive: hive,
          inspection: inspection,
        ),
      ),
    );
    if ((result ?? false) && context.mounted) {
      context.read<InspectionsCubit>().load();
      await _refreshHive(context);
    }
  }

  Future<void> _refreshHive(BuildContext context) async {
    try {
      final fresh = await HiveRepository(
        api: context.read<ApiClient>(),
      ).getHive(apiaryId, hive.id);
      if (context.mounted) onHiveChanged(fresh);
    } catch (_) {}
  }

  Future<void> _confirmDelete(
    BuildContext context,
    Inspection inspection,
  ) async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(l10n.inspectionDeleteConfirm),
        content: Text(l10n.inspectionDeleteWarning),
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
      context.read<InspectionsCubit>().delete(inspection.id);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.inspectionTitle),
        actions: const [ProfileIconButton()],
      ),
      body: BlocBuilder<InspectionsCubit, InspectionsState>(
        builder: (context, state) {
          if (state is InspectionsLoading || state is InspectionsInitial) {
            return const Center(child: CircularProgressIndicator());
          }
          if (state is InspectionsError) {
            return Center(child: Text(l10n.generalError));
          }
          final loaded = state as InspectionsLoaded;
          final inspections = loaded.inspections;
          if (inspections.isEmpty) {
            return Column(
              children: [
                Expanded(
                  child: Center(child: Text(l10n.inspectionEmpty)),
                ),
                _InspectionBanner(
                  loaded: loaded,
                  onAdd: () => _openCreate(context),
                  onPage: (p) => context.read<InspectionsCubit>().goToPage(p),
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
                          for (final insp in inspections) ...[
                            InspectionCard(
                              apiaryId: apiaryId,
                              inspection: insp,
                              onEdit: () => _openEdit(context, insp),
                              onDelete: () => _confirmDelete(context, insp),
                            ),
                            const SizedBox(height: 8),
                          ],
                        ],
                      ),
                    ),
                  ),
                ),
              ),
              _InspectionBanner(
                loaded: loaded,
                onAdd: () => _openCreate(context),
                onPage: (p) => context.read<InspectionsCubit>().goToPage(p),
              ),
            ],
          );
        },
      ),
    );
  }
}

class InspectionCard extends StatelessWidget {
  final int apiaryId;
  final Inspection inspection;
  final VoidCallback onEdit;
  final VoidCallback onDelete;

  const InspectionCard({
    required this.apiaryId,
    required this.inspection,
    required this.onEdit,
    required this.onDelete,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;
    final dateStr = DateFormat.yMMMd(
      Localizations.localeOf(context).toString(),
    ).add_Hm().format(inspection.inspectedAt);
    final myName = context.read<TokenStorage>().name;
    final bodyStyle = textTheme.bodyMedium?.copyWith(
      color: colorScheme.onSurfaceVariant,
    );

    return LayoutBuilder(
      builder: (context, constraints) {
        // 16*2 padding + ~48 for the popup menu button.
        final textWidth = constraints.maxWidth - 80;
        final truncated = inspection.notes.isNotEmpty &&
            isTextTruncated(inspection.notes, bodyStyle, textWidth);

        return Card(
          child: InkWell(
            onTap: truncated
                ? () => showNoteDialog(
                      context,
                      title: l10n.inspectionNote,
                      note: inspection.notes,
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
                        const SizedBox(height: 4),
                        InspectionSummary(
                          inspection: inspection,
                          currentUserName: myName,
                        ),
                      ],
                    ),
                  ),
                  PopupMenuButton<_CardAction>(
                    onSelected: (action) {
                      if (action == _CardAction.edit) onEdit();
                      if (action == _CardAction.delete) onDelete();
                      if (action == _CardAction.photos) {
                        showInspectionPhotosSheet(
                          context,
                          apiaryId: apiaryId,
                          hiveId: inspection.hiveId,
                          inspection: inspection,
                          imageRepo: InspectionImageRepository(
                            api: context.read<ApiClient>(),
                          ),
                        );
                      }
                    },
                    itemBuilder: (_) => [
                      if (inspection.photoCount > 0)
                        PopupMenuItem(
                          value: _CardAction.photos,
                          child: Text(l10n.inspectionPhotos),
                        ),
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

enum _CardAction { photos, edit, delete }

class _InspectionBanner extends StatelessWidget {
  final InspectionsLoaded loaded;
  final VoidCallback onAdd;
  final ValueChanged<int> onPage;

  const _InspectionBanner({
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
                    tooltip: l10n.inspectionAdd,
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
