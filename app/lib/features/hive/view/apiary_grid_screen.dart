import 'dart:math';

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';

import '../../../core/api/api_client.dart';
import '../../../core/theme/app_layout.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../l10n/app_localizations.dart';
import '../../apiary/data/apiary_model.dart';
import '../../inspection/data/inspection_model.dart';
import '../../inspection/data/inspection_repository.dart';
import '../cubit/hives_cubit.dart';
import '../data/hive_model.dart';
import '../data/hive_repository.dart';
import '../../treatment/view/bulk_treatment_form_screen.dart';
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
  final TransformationController _transformController =
      TransformationController();

  @override
  void dispose() {
    _transformController.dispose();
    super.dispose();
  }

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

  Future<void> _openBulkTreatment(BuildContext context) async {
    final l10n = AppLocalizations.of(context)!;
    final count = await Navigator.of(context).push<int>(
      MaterialPageRoute(
        builder: (_) => BulkTreatmentFormScreen(apiaryId: widget.apiary.id),
      ),
    );
    if (count != null && context.mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.treatmentBulkSuccess(count))),
      );
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
      appBar: AppBar(
        title: Text(widget.apiary.name),
        actions: const [ProfileIconButton()],
      ),
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
                    child: Text(l10n.generalRetry),
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
            final gridContentWidth = widget.apiary.gridCols * cellSize +
                (widget.apiary.gridCols - 1) * spacing;
            return Column(
              children: [
                Expanded(
                  child: LayoutBuilder(
                    builder: (context, constraints) {
                      final gridContentHeight =
                          widget.apiary.gridRows * cellSize +
                              (widget.apiary.gridRows - 1) * spacing;
                      final contentW = gridContentWidth + padding * 2;
                      final contentH = gridContentHeight + padding * 2;
                      final canvasW =
                          max(constraints.maxWidth, contentW);
                      final canvasH =
                          max(constraints.maxHeight, contentH);
                      return InteractiveViewer(
                        transformationController: _transformController,
                        constrained: false,
                        boundaryMargin:
                            const EdgeInsets.all(double.infinity),
                        minScale: 0.3,
                        maxScale: 4.0,
                        trackpadScrollCausesScale: true,
                        child: SizedBox(
                          width: canvasW,
                          height: canvasH,
                          child: Center(
                            child: Padding(
                              padding: const EdgeInsets.all(padding),
                              child: SizedBox(
                                width: gridContentWidth,
                                child: GridView.builder(
                                  shrinkWrap: true,
                                  physics:
                                      const NeverScrollableScrollPhysics(),
                                  gridDelegate:
                                      SliverGridDelegateWithFixedCrossAxisCount(
                                    crossAxisCount: widget.apiary.gridCols,
                                    crossAxisSpacing: spacing,
                                    mainAxisSpacing: spacing,
                                  ),
                                  itemCount: widget.apiary.gridRows *
                                      widget.apiary.gridCols,
                                  itemBuilder: (context, index) {
                                    final row =
                                        index ~/ widget.apiary.gridCols;
                                    final col =
                                        index % widget.apiary.gridCols;
                                    final hive = hiveMap[(row, col)];
                                    return hive != null
                                        ? _HiveCell(
                                            hive: hive,
                                            dimmed: _activeFilters
                                                    .isNotEmpty &&
                                                !_matches(hive),
                                            onTap: () => _openDetail(
                                                context, hive),
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
                              ),
                            ),
                          ),
                        ),
                      );
                    },
                  ),
                ),
                _FilterBar(
                  activeFilters: _activeFilters,
                  onToggle: _toggleFilter,
                  onCenter: () =>
                      _transformController.value = Matrix4.identity(),
                  onTreatAll: () => _openBulkTreatment(context),
                  l10n: l10n,
                  apiaryId: widget.apiary.id,
                  hives: state.hives,
                  onHiveTap: (hive) async {
                    await Navigator.of(context).push(
                      MaterialPageRoute(
                        builder: (_) => HiveDetailScreen(
                          hive: hive,
                          apiaryId: widget.apiary.id,
                        ),
                      ),
                    );
                    if (context.mounted) context.read<HivesCubit>().load();
                  },
                ),
              ],
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
  final VoidCallback onCenter;
  final VoidCallback onTreatAll;
  final AppLocalizations l10n;
  final int apiaryId;
  final List<Hive> hives;
  final void Function(Hive) onHiveTap;

  const _FilterBar({
    required this.activeFilters,
    required this.onToggle,
    required this.onCenter,
    required this.onTreatAll,
    required this.l10n,
    required this.apiaryId,
    required this.hives,
    required this.onHiveTap,
  });

  void _showFilterSheet(BuildContext context) {
    final localFilters = Set<_HiveFilter>.from(activeFilters);
    final chips = [
      (_HiveFilter.readyForHarvest, l10n.hiveReadyForHarvest),
      (_HiveFilter.queenless, l10n.hiveQueenless),
      (_HiveFilter.sick, l10n.hiveSick),
    ];
    showDialog<void>(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setDialogState) {
          final isWide = MediaQuery.sizeOf(ctx).width >= 600;
          return Dialog(
            shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(20)),
            child: ConstrainedBox(
              constraints: BoxConstraints(maxWidth: isWide ? 480 : 360),
              child: Padding(
                padding:
                    const EdgeInsets.symmetric(horizontal: 24, vertical: 24),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Row(
                      children: [
                        Icon(Icons.tune,
                            size: 20,
                            color: Theme.of(ctx)
                                .colorScheme
                                .onSurfaceVariant),
                        const SizedBox(width: 8),
                        Text(l10n.hiveFilterTooltip,
                            style: Theme.of(ctx).textTheme.titleMedium),
                      ],
                    ),
                    const SizedBox(height: 16),
                    Wrap(
                      alignment: WrapAlignment.center,
                      spacing: 10,
                      runSpacing: 8,
                      children: chips.map((entry) {
                        final (filter, label) = entry;
                        final selected = localFilters.contains(filter);
                        return FilterChip(
                          label: Text(label),
                          padding: const EdgeInsets.symmetric(
                              horizontal: 12, vertical: 8),
                          selected: selected,
                          onSelected: (_) {
                            onToggle(filter);
                            setDialogState(() {
                              if (localFilters.contains(filter)) {
                                localFilters.remove(filter);
                              } else {
                                localFilters.add(filter);
                              }
                            });
                          },
                        );
                      }).toList(),
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

  void _showHiveSheet(BuildContext context) {
    showDialog<void>(
      context: context,
      builder: (_) => _HiveListDialog(
        apiaryId: apiaryId,
        hives: hives,
        onHiveTap: onHiveTap,
        onTreatAll: onTreatAll,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final bannerWidth = AppLayout.bannerWidth(context);

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
                  Badge(
                    isLabelVisible: activeFilters.isNotEmpty,
                    label: Text('${activeFilters.length}'),
                    child: IconButton(
                      icon: const Icon(Icons.tune),
                      iconSize: 28,
                      tooltip: l10n.hiveFilterTooltip,
                      onPressed: () => _showFilterSheet(context),
                    ),
                  ),
                  IconButton(
                    icon: const Icon(Icons.format_list_bulleted),
                    iconSize: 28,
                    tooltip: l10n.hiveListTooltip,
                    onPressed:
                        hives.isEmpty ? null : () => _showHiveSheet(context),
                  ),
                  IconButton(
                    icon: const Icon(Icons.center_focus_strong_outlined),
                    iconSize: 28,
                    tooltip: l10n.apiaryCenterView,
                    onPressed: onCenter,
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _HiveListDialog extends StatefulWidget {
  final int apiaryId;
  final List<Hive> hives;
  final void Function(Hive) onHiveTap;
  final VoidCallback onTreatAll;

  const _HiveListDialog({
    required this.apiaryId,
    required this.hives,
    required this.onHiveTap,
    required this.onTreatAll,
  });

  @override
  State<_HiveListDialog> createState() => _HiveListDialogState();
}

class _HiveListDialogState extends State<_HiveListDialog> {
  final Map<int, Inspection?> _lastInspections = {};

  @override
  void initState() {
    super.initState();
    _loadInspections();
  }

  Future<void> _loadInspections() async {
    final repo = InspectionRepository(api: context.read<ApiClient>());
    final hivesWithInspections =
        widget.hives.where((h) => h.lastInspectedAt != null).toList();
    await Future.wait(
      hivesWithInspections.map((hive) async {
        try {
          final list = await repo.listInspections(
            widget.apiaryId,
            hive.id,
            limit: 1,
          );
          if (mounted) {
            setState(() {
              _lastInspections[hive.id] = list.items.isNotEmpty ? list.items.first : null;
            });
          }
        } catch (_) {}
      }),
    );
  }

  String _hint(AppLocalizations l10n, Inspection inspection) {
    if (inspection.notes.isNotEmpty) return inspection.notes;
    final parts = <String>[];
    if (inspection.queenSeen == 'seen') {
      parts.add(l10n.inspectionQueenStatusSeen);
    } else if (inspection.queenSeen == 'not_seen') {
      parts.add(l10n.inspectionQueenStatusNotSeen);
    }
    if (inspection.broodPattern.isNotEmpty &&
        inspection.broodPattern != 'none') {
      final broodLabel = switch (inspection.broodPattern) {
        'excellent' => l10n.inspectionBroodExcellent,
        'good' => l10n.inspectionBroodGood,
        'poor' => l10n.inspectionBroodPoor,
        _ => inspection.broodPattern,
      };
      parts.add('${l10n.inspectionBroodPattern}: $broodLabel');
    }
    return parts.join(' · ');
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final locale = Localizations.localeOf(context).toString();
    final isWide = MediaQuery.sizeOf(context).width >= 600;

    final sortedHives = [...widget.hives]..sort((a, b) {
        if (a.lastInspectedAt == null && b.lastInspectedAt == null) return 0;
        if (a.lastInspectedAt == null) return -1;
        if (b.lastInspectedAt == null) return 1;
        return a.lastInspectedAt!.compareTo(b.lastInspectedAt!);
      });

    return Dialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
      child: ConstrainedBox(
        constraints: BoxConstraints(maxWidth: isWide ? 480 : 380),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(24, 24, 24, 8),
              child: Row(
                children: [
                  Icon(Icons.format_list_bulleted,
                      size: 20,
                      color: Theme.of(context).colorScheme.onSurfaceVariant),
                  const SizedBox(width: 8),
                  Text(l10n.hiveListTooltip,
                      style: Theme.of(context).textTheme.titleMedium),
                ],
              ),
            ),
            const Divider(height: 1),
            Flexible(
              child: ListView.builder(
                shrinkWrap: true,
                itemCount: sortedHives.length,
                itemBuilder: (_, i) {
                  final hive = sortedHives[i];
                  final dateStr = hive.lastInspectedAt != null
                      ? DateFormat('d MMM yyyy', locale)
                          .format(hive.lastInspectedAt!)
                      : l10n.hiveDetailNoInspections;

                  final inspection = _lastInspections[hive.id];
                  String subtitleText = dateStr;
                  if (isWide &&
                      hive.lastInspectedAt != null &&
                      inspection != null) {
                    final hint = _hint(l10n, inspection);
                    if (hint.isNotEmpty) subtitleText = '$dateStr · $hint';
                  }

                  return ListTile(
                    leading: const Icon(Icons.hive),
                    title: Text(hive.name),
                    subtitle: Text(
                      subtitleText,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    trailing: _HiveStatusIcons(hive: hive),
                    onTap: () {
                      Navigator.of(context).pop();
                      widget.onHiveTap(hive);
                    },
                  );
                },
              ),
            ),
            if (widget.hives.length > 1) ...[
              const Divider(height: 1),
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
                child: ElevatedButton.icon(
                  onPressed: () {
                    Navigator.of(context).pop();
                    widget.onTreatAll();
                  },
                  style: ElevatedButton.styleFrom(
                    visualDensity: VisualDensity.compact,
                  ),
                  icon: const Icon(Icons.medical_services_outlined, size: 18),
                  label: Text(l10n.treatmentTreatAllHives),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _HiveStatusIcons extends StatelessWidget {
  final Hive hive;

  const _HiveStatusIcons({required this.hive});

  @override
  Widget build(BuildContext context) {
    final color = Theme.of(context).colorScheme.onSurfaceVariant;
    final icons = [
      if (hive.queenless) (Icons.female_outlined, color),
      if (hive.readyForHarvest) (Icons.water_drop_outlined, Colors.amber.shade700),
      if (hive.diseases.isNotEmpty) (Icons.coronavirus_outlined, Colors.red.shade400),
    ];
    if (icons.isEmpty) return const SizedBox.shrink();
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: icons
          .map((e) => Icon(e.$1, size: 26, color: e.$2))
          .toList(),
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
