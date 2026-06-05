import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';

import '../../../core/api/api_client.dart';
import '../../../core/theme/app_layout.dart';
import '../../../l10n/app_localizations.dart';
import '../../hive/data/hive_model.dart';
import '../../hive/data/hive_repository.dart';
import '../cubit/inspections_cubit.dart';
import '../data/inspection_model.dart';
import '../data/inspection_repository.dart';
import 'inspection_form_screen.dart';
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
      final hives = await HiveRepository(
        api: context.read<ApiClient>(),
      ).listHives(apiaryId);
      final fresh = hives.firstWhere(
        (h) => h.id == hive.id,
        orElse: () => hive,
      );
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
      appBar: AppBar(title: Text(l10n.inspectionTitle)),
      body: BlocBuilder<InspectionsCubit, InspectionsState>(
        builder: (context, state) {
          if (state is InspectionsLoading || state is InspectionsInitial) {
            return const Center(child: CircularProgressIndicator());
          }
          if (state is InspectionsError) {
            return Center(child: Text(l10n.generalError));
          }
          final inspections = (state as InspectionsLoaded).inspections;
          if (inspections.isEmpty) {
            return Center(
              child: SizedBox(
                width: 200,
                child: ElevatedButton(
                  onPressed: () => _openCreate(context),
                  child: Text(l10n.inspectionAdd),
                ),
              ),
            );
          }
          return Center(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: ConstrainedBox(
                constraints: AppLayout.formConstraints(context),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    for (final insp in inspections) ...[
                      _InspectionCard(
                        inspection: insp,
                        onEdit: () => _openEdit(context, insp),
                        onDelete: () => _confirmDelete(context, insp),
                      ),
                      const SizedBox(height: 8),
                    ],
                    const SizedBox(height: 8),
                    SizedBox(
                      width: 200,
                      child: ElevatedButton(
                        onPressed: () => _openCreate(context),
                        child: Text(l10n.inspectionAdd),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          );
        },
      ),
    );
  }
}

class _InspectionCard extends StatelessWidget {
  final Inspection inspection;
  final VoidCallback onEdit;
  final VoidCallback onDelete;

  const _InspectionCard({
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
    ).format(inspection.inspectedAt);

    return Card(
      child: InkWell(
        onTap: onEdit,
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
                    InspectionSummary(inspection: inspection),
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
  }
}


enum _CardAction { edit, delete }
