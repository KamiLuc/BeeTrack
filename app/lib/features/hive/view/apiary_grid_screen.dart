import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';

import '../../../l10n/app_localizations.dart';
import '../../apiary/data/apiary_model.dart';
import '../cubit/hives_cubit.dart';
import '../data/hive_model.dart';
import '../data/hive_repository.dart';
import 'add_hive_screen.dart';
import 'hive_detail_screen.dart';

enum _HiveFilter { readyForHarvest, queenless, sick }

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

class _ApiaryGridView extends StatefulWidget {
  final Apiary apiary;

  const _ApiaryGridView({required this.apiary});

  @override
  State<_ApiaryGridView> createState() => _ApiaryGridViewState();
}

class _ApiaryGridViewState extends State<_ApiaryGridView> {
  final Set<_HiveFilter> _activeFilters = {};

  void _toggleFilter(_HiveFilter filter) {
    setState(() {
      if (_activeFilters.contains(filter)) {
        _activeFilters.remove(filter);
      } else {
        _activeFilters.add(filter);
      }
    });
  }

  bool _matches(Hive hive) {
    if (_activeFilters.isEmpty) return true;
    if (_activeFilters.contains(_HiveFilter.readyForHarvest) &&
        hive.readyForHarvest) return true;
    if (_activeFilters.contains(_HiveFilter.queenless) && hive.queenless) {
      return true;
    }
    if (_activeFilters.contains(_HiveFilter.sick) && hive.diseases.isNotEmpty) {
      return true;
    }
    return false;
  }

  Future<void> _openDetail(BuildContext context, Hive hive) async {
    await Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => HiveDetailScreen(hive: hive, apiaryId: widget.apiary.id),
      ),
    );
    if (context.mounted) context.read<HivesCubit>().load();
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
          apiaryId: widget.apiary.id,
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
      appBar: AppBar(title: Text(widget.apiary.name)),
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
                  Text(
                    l10n.generalError,
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(height: 12),
                  ElevatedButton(
                    onPressed: () => context.read<HivesCubit>().load(),
                    style: ElevatedButton.styleFrom(
                      minimumSize: const Size(120, 40),
                    ),
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
            final maxGridWidth = widget.apiary.gridCols * cellSize +
                (widget.apiary.gridCols - 1) * spacing +
                padding * 2;
            return RefreshIndicator(
              onRefresh: () => context.read<HivesCubit>().load(),
              child: Center(
                child: SingleChildScrollView(
                  padding: const EdgeInsets.fromLTRB(padding, 8, padding, padding),
                  child: ConstrainedBox(
                    constraints: BoxConstraints(maxWidth: maxGridWidth),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        _FilterBar(
                          activeFilters: _activeFilters,
                          onToggle: _toggleFilter,
                          l10n: l10n,
                        ),
                        const SizedBox(height: 8),
                        GridView.builder(
                            shrinkWrap: true,
                            physics: const NeverScrollableScrollPhysics(),
                            gridDelegate:
                                SliverGridDelegateWithFixedCrossAxisCount(
                              crossAxisCount: widget.apiary.gridCols,
                              crossAxisSpacing: spacing,
                              mainAxisSpacing: spacing,
                            ),
                            itemCount:
                                widget.apiary.gridRows * widget.apiary.gridCols,
                            itemBuilder: (context, index) {
                              final row = index ~/ widget.apiary.gridCols;
                              final col = index % widget.apiary.gridCols;
                              final hive = hiveMap[(row, col)];
                              return hive != null
                                  ? _HiveCell(
                                      hive: hive,
                                      dimmed: _activeFilters.isNotEmpty &&
                                          !_matches(hive),
                                      onTap: () =>
                                          _openDetail(context, hive),
                                    )
                                  : _EmptyCell(
                                      onTap: () => _openAddHive(
                                          context, state, row, col),
                                      onDrop: (h) => context
                                          .read<HivesCubit>()
                                          .move(h.id, row, col),
                                    );
                            },
                          ),
                        ],
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

class _FilterBar extends StatelessWidget {
  final Set<_HiveFilter> activeFilters;
  final void Function(_HiveFilter) onToggle;
  final AppLocalizations l10n;

  const _FilterBar({
    required this.activeFilters,
    required this.onToggle,
    required this.l10n,
  });

  @override
  Widget build(BuildContext context) {
    final chips = [
      (_HiveFilter.readyForHarvest, l10n.hiveReadyForHarvest),
      (_HiveFilter.queenless, l10n.hiveQueenless),
      (_HiveFilter.sick, l10n.hiveSick),
    ];
    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      child: Row(
        children: chips.map((entry) {
          final (filter, label) = entry;
          final selected = activeFilters.contains(filter);
          return Padding(
            padding: const EdgeInsets.only(right: 8),
            child: FilterChip(
              label: Text(label),
              selected: selected,
              onSelected: (_) => onToggle(filter),
            ),
          );
        }).toList(),
      ),
    );
  }
}

class _HiveCell extends StatelessWidget {
  final Hive hive;
  final bool dimmed;
  final VoidCallback onTap;

  const _HiveCell({
    required this.hive,
    required this.dimmed,
    required this.onTap,
  });

  Widget _buildContent(BuildContext context, {double opacity = 1.0}) {
    final l10n = AppLocalizations.of(context)!;
    final colorScheme = Theme.of(context).colorScheme;
    final primary = colorScheme.primary;
    final locale = Localizations.localeOf(context).toString();

    return Opacity(
      opacity: opacity,
      child: Card(
        margin: EdgeInsets.zero,
        color: primary.withAlpha(40),
        clipBehavior: Clip.antiAlias,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(12),
          side: const BorderSide(color: Colors.black),
        ),
        child: Padding(
          padding: const EdgeInsets.all(6),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                Icons.hive,
                size: 24,
                color: hive.active ? primary : colorScheme.outline,
              ),
              const SizedBox(height: 2),
              Flexible(
                child: Text(
                  hive.name,
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        fontWeight: FontWeight.w600,
                        fontSize: 12,
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
                        fontSize: 9,
                      ),
                ),
              if (hive.lastInspectedAt != null)
                Text(
                  DateFormat('d MMM', locale).format(hive.lastInspectedAt!),
                  style: Theme.of(context).textTheme.labelSmall?.copyWith(
                        color: colorScheme.onSurfaceVariant,
                        fontSize: 9,
                      ),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
            ],
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final effectiveOpacity = dimmed ? 0.3 : 1.0;
    return LongPressDraggable<Hive>(
      data: hive,
      delay: const Duration(milliseconds: 300),
      feedback: SizedBox(
        width: 80,
        height: 80,
        child: Material(
          color: Colors.transparent,
          elevation: 8,
          borderRadius: BorderRadius.circular(12),
          child: _buildContent(context),
        ),
      ),
      childWhenDragging: _buildContent(context, opacity: 0.35),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: _buildContent(context, opacity: effectiveOpacity),
      ),
    );
  }
}

class _EmptyCell extends StatelessWidget {
  final VoidCallback? onTap;
  final void Function(Hive)? onDrop;

  const _EmptyCell({this.onTap, this.onDrop});

  @override
  Widget build(BuildContext context) {
    return DragTarget<Hive>(
      onWillAcceptWithDetails: (_) => true,
      onAcceptWithDetails: (details) => onDrop?.call(details.data),
      builder: (context, candidateData, _) {
        final hovering = candidateData.isNotEmpty;
        return InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(12),
          child: AnimatedContainer(
            duration: const Duration(milliseconds: 150),
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(12),
              border: Border.all(
                color: hovering
                    ? Theme.of(context).colorScheme.primary
                    : Theme.of(context).colorScheme.outlineVariant,
                width: hovering ? 2 : 1,
              ),
              color: hovering
                  ? Theme.of(context).colorScheme.primary.withAlpha(20)
                  : null,
            ),
          ),
        );
      },
    );
  }
}
