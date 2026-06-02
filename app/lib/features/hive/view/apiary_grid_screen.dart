import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../l10n/app_localizations.dart';
import '../../apiary/data/apiary_model.dart';
import '../cubit/hives_cubit.dart';
import '../data/hive_model.dart';
import '../data/hive_repository.dart';
import 'add_hive_screen.dart';
import 'edit_hive_screen.dart';

class ApiaryGridScreen extends StatelessWidget {
  final Apiary apiary;

  const ApiaryGridScreen({super.key, required this.apiary});

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => HivesCubit(
        repo: HiveRepository(api: context.read()),
        apiaryId: apiary.id,
      )..load(),
      child: _ApiaryGridView(apiary: apiary),
    );
  }
}

class _ApiaryGridView extends StatelessWidget {
  final Apiary apiary;

  const _ApiaryGridView({required this.apiary});

  Future<void> _openEdit(BuildContext context, Hive hive) async {
    await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => EditHiveScreen(apiaryId: apiary.id, hive: hive),
      ),
    );
    if (context.mounted) context.read<HivesCubit>().load();
  }

  Future<void> _confirmDelete(BuildContext context, Hive hive) async {
    final l10n = AppLocalizations.of(context)!;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(l10n.hiveDeleteConfirm),
        content: Text(l10n.hiveDeleteWarning),
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
      context.read<HivesCubit>().delete(hive.id);
    }
  }

  Future<void> _openAddHive(
    BuildContext context,
    HivesLoaded state,
    int row,
    int col,
  ) async {
    final l10n = AppLocalizations.of(context)!;
    final defaultName = l10n.hiveDefaultName(state.hives.length + 1);
    await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => AddHiveScreen(
          apiaryId: apiary.id,
          gridRow: row,
          gridCol: col,
          defaultName: defaultName,
        ),
      ),
    );
    if (context.mounted) context.read<HivesCubit>().load();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      appBar: AppBar(title: Text(apiary.name)),
      body: BlocBuilder<HivesCubit, HivesState>(
        builder: (context, state) {
          if (state is HivesInitial || state is HivesLoading) {
            return const Center(child: CircularProgressIndicator());
          }
          if (state is HivesError) {
            return Center(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(l10n.generalError),
                  const SizedBox(height: 12),
                  ElevatedButton(
                    onPressed: () => context.read<HivesCubit>().load(),
                    child: const Text('Retry'),
                  ),
                ],
              ),
            );
          }
          if (state is HivesLoaded) {
            final hiveMap = {
              for (final h in state.hives) (h.gridRow, h.gridCol): h,
            };
            const double cellSize = 80;
            const double spacing = 8;
            const double padding = 16;
            final maxGridWidth = apiary.gridCols * cellSize +
                (apiary.gridCols - 1) * spacing +
                padding * 2;
            return RefreshIndicator(
              onRefresh: () => context.read<HivesCubit>().load(),
              child: Center(
                child: SingleChildScrollView(
                  padding: const EdgeInsets.all(padding),
                  child: ConstrainedBox(
                    constraints: BoxConstraints(maxWidth: maxGridWidth),
                    child: GridView.builder(
                      shrinkWrap: true,
                      physics: const NeverScrollableScrollPhysics(),
                      gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                        crossAxisCount: apiary.gridCols,
                        crossAxisSpacing: spacing,
                        mainAxisSpacing: spacing,
                      ),
                      itemCount: apiary.gridRows * apiary.gridCols,
                      itemBuilder: (context, index) {
                        final row = index ~/ apiary.gridCols;
                        final col = index % apiary.gridCols;
                        final hive = hiveMap[(row, col)];
                        return hive != null
                            ? _HiveCell(
                                hive: hive,
                                onEdit: () => _openEdit(context, hive),
                                onDelete: () => _confirmDelete(context, hive),
                              )
                            : _EmptyCell(
                                onTap: () => _openAddHive(context, state, row, col),
                              );
                      },
                    ),
                  ),
                ),
              ),
            );
          }
          return const SizedBox.shrink();
        },
      ),
    );
  }
}

class _HiveCell extends StatelessWidget {
  final Hive hive;
  final VoidCallback onEdit;
  final VoidCallback onDelete;

  const _HiveCell({
    required this.hive,
    required this.onEdit,
    required this.onDelete,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;
    final primary = colorScheme.primary;

    return Card(
      margin: EdgeInsets.zero,
      color: primary.withAlpha(40),
      clipBehavior: Clip.antiAlias,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: const BorderSide(color: Colors.black),
      ),
      child: PopupMenuButton<_HiveCellAction>(
        padding: EdgeInsets.zero,
        onSelected: (action) {
          if (action == _HiveCellAction.edit) onEdit();
          if (action == _HiveCellAction.delete) onDelete();
        },
        itemBuilder: (_) => [
          PopupMenuItem(
            value: _HiveCellAction.edit,
            child: Text(l10n.generalEdit),
          ),
          PopupMenuItem(
            value: _HiveCellAction.delete,
            child: Text(
              l10n.generalDelete,
              style: TextStyle(color: colorScheme.error),
            ),
          ),
        ],
        child: Padding(
          padding: const EdgeInsets.all(8),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                Icons.hive,
                size: 28,
                color: hive.active ? primary : colorScheme.outline,
              ),
              const SizedBox(height: 4),
              Flexible(
                child: Text(
                  hive.name,
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        fontWeight: FontWeight.w600,
                      ),
                  textAlign: TextAlign.center,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              if (!hive.active)
                Text(
                  l10n.hiveInactive,
                  style: Theme.of(context).textTheme.labelSmall?.copyWith(
                        color: colorScheme.outline,
                      ),
                ),
            ],
          ),
        ),
      ),
    );
  }
}

enum _HiveCellAction { edit, delete }

class _EmptyCell extends StatelessWidget {
  final VoidCallback? onTap;

  const _EmptyCell({this.onTap});

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(12),
      child: Container(
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: Theme.of(context).colorScheme.outlineVariant,
          ),
        ),
      ),
    );
  }
}
